package gcal

import (
	"fmt"
	"os/exec"
	"runtime"
)

// openBrowserFn opens url in the user's default browser (best-effort).
func openBrowserFn(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url) //nolint:gosec
	case "darwin":
		cmd = exec.Command("open", url) //nolint:gosec
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url) //nolint:gosec
	default:
		fmt.Printf("Open this URL in your browser:\n%s\n", url)
		return
	}
	if err := cmd.Start(); err != nil {
		fmt.Printf("Couldn't open browser (%v). Open this URL manually:\n%s\n", err, url)
	}
}
