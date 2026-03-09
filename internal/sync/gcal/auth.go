// Package gcal provides a Google Calendar adapter for nullcal.
// Credentials are read from environment variables (or a .env file):
//
//   GOOGLE_CLIENT_ID     — OAuth2 client ID
//   GOOGLE_CLIENT_SECRET — OAuth2 client secret
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
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
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
		// No cached token – start consent flow.
		tok, err = runConsentFlow(cfg)
		if err != nil {
			return nil, err
		}
		if saveErr := saveToken(path, tok); saveErr != nil {
			fmt.Fprintf(os.Stderr, "warning: could not save token: %v\n", saveErr)
		}
	}

	return cfg.Client(ctx, tok), nil
}

// runConsentFlow opens the browser / prints the URL and waits for the
// authorization code that the user pastes back.
func runConsentFlow(cfg *oauth2.Config) (*oauth2.Token, error) {
	url := cfg.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Println("\n── Google Calendar Auth ──────────────────────────────")
	fmt.Println("Open this URL in your browser, then paste the code below:")
	fmt.Println(url)
	fmt.Print("\nAuthorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("reading authorization code: %w", err)
	}

	tok, err := cfg.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	return tok, nil
}
