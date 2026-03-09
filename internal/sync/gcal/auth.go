// Package gcal provides a Google Calendar adapter for nullcal.
// Credentials are read from environment variables (or a .env file):
//
//	GOOGLE_CLIENT_ID     — OAuth2 client ID
//	GOOGLE_CLIENT_SECRET — OAuth2 client secret
//
// Token is stored at $XDG_CONFIG_HOME/nullcal/token.json and refreshed
// automatically on subsequent runs.
package gcal

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	googlecalendar "google.golang.org/api/calendar/v3"
)

// callbackPath is the path registered as redirect URI in Google Cloud Console.
const callbackPath = "/api/auth/callback/google"

// callbackAddr is the address the temporary auth listener binds to.
// Must match the port in the redirect URI you registered.
const callbackAddr = "localhost:7331"

// oauthConfig builds the OAuth2 config from environment variables.
func oauthConfig() (*oauth2.Config, error) {
	id := os.Getenv("GOOGLE_CLIENT_ID")
	secret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if id == "" || secret == "" {
		return nil, fmt.Errorf("GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set (hint: create a .env file)")
	}
	return &oauth2.Config{
		ClientID:     id,
		ClientSecret: secret,
		RedirectURL:  "http://" + callbackAddr + callbackPath,
		Scopes:       []string{googlecalendar.CalendarReadonlyScope},
		Endpoint:     google.Endpoint,
	}, nil
}

// tokenPath returns the path where the OAuth2 token is cached.
func tokenPath() string {
	cfgDir := os.Getenv("XDG_CONFIG_HOME")
	if cfgDir == "" {
		home, _ := os.UserHomeDir()
		cfgDir = filepath.Join(home, ".config")
	}
	return filepath.Join(cfgDir, "nullcal", "token.json")
}

// loadToken reads a cached token from disk.
func loadToken(path string) (*oauth2.Token, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err := json.NewDecoder(f).Decode(tok); err != nil {
		return nil, err
	}
	return tok, nil
}

// saveToken persists a token to disk (creating parent dirs if needed).
func saveToken(path string, tok *oauth2.Token) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tok)
}

// AuthClient returns an authenticated *http.Client using cached token or
// performing the OAuth2 consent flow when no token exists.
func AuthClient(ctx context.Context) (*http.Client, error) {
	cfg, err := oauthConfig()
	if err != nil {
		return nil, err
	}

	path := tokenPath()
	tok, err := loadToken(path)
	if err != nil {
		// No cached token – start HTTP callback consent flow.
		tok, err = runCallbackFlow(ctx, cfg)
		if err != nil {
			return nil, err
		}
		if saveErr := saveToken(path, tok); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save token: %v\n", saveErr)
		}
	}

	return cfg.Client(ctx, tok), nil
}

// runCallbackFlow starts a temporary HTTP server on callbackAddr, opens the
// browser to the Google consent page, and waits for the redirect callback to
// receive the authorization code and exchange it for a token.
func runCallbackFlow(ctx context.Context, cfg *oauth2.Config) (*oauth2.Token, error) {
	state := "nullcal-auth"
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()
	srv := &http.Server{Addr: callbackAddr, Handler: mux}

	mux.HandleFunc(callbackPath, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("state"); got != state {
			errCh <- fmt.Errorf("oauth state mismatch: got %q", got)
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}
		code := r.URL.Query().Get("code")
		if code == "" {
			errCh <- fmt.Errorf("no code in callback: %s", r.URL.RawQuery)
			http.Error(w, "missing code", http.StatusBadRequest)
			return
		}
		// Friendly confirmation page.
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, `<!DOCTYPE html><html><body style="font-family:monospace;background:#111;color:#e0e0e0;padding:40px">
<h2 style="color:#50fa7b">✓ nullcal authorised</h2>
<p>You can close this tab and return to the terminal.</p>
</body></html>`)
		codeCh <- code
	})

	// Start the temporary listener in a goroutine.
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("auth server: %w", err)
		}
	}()

	authURL := cfg.AuthCodeURL(state, oauth2.AccessTypeOffline)
	fmt.Println("\n── Google Calendar Auth ──────────────────────────────────────")
	fmt.Println("Opening browser for Google consent. If it doesn't open, visit:")
	fmt.Println(authURL)
	fmt.Println("──────────────────────────────────────────────────────────────")

	// Try to open the browser.
	openBrowserFn(authURL)

	// Wait for code or error.
	var code string
	select {
	case code = <-codeCh:
	case err := <-errCh:
		_ = srv.Shutdown(context.Background())
		return nil, err
	case <-ctx.Done():
		_ = srv.Shutdown(context.Background())
		return nil, ctx.Err()
	}

	_ = srv.Shutdown(context.Background())

	tok, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	return tok, nil
}
