package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	githubOwner = "v4nsh0x"
	githubRepo  = "pengu"
	releaseAPI  = "https://api.github.com/repos/" + githubOwner + "/" + githubRepo + "/releases/latest"
)

// githubRelease represents the relevant fields from a GitHub release.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset represents a downloadable file attached to a release.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// handleUpdate is the entry point for the `pengu update` command.
func handleUpdate() {
	fmt.Println("Checking for updates...")
	fmt.Println()

	currentVersion := version
	fmt.Printf("  Current version: v%s\n", currentVersion)

	// Fetch latest release from GitHub
	release, err := fetchLatestRelease()
	if err != nil {
		fmt.Printf("\nUpdate failed:\n%s\n", err)
		os.Exit(1)
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	fmt.Printf("  Latest version:  v%s\n", latestVersion)
	fmt.Println()

	// Compare versions
	if latestVersion == currentVersion {
		fmt.Println("🐧 Pengu is already up to date!")
		return
	}

	// Resolve the correct asset for this OS/arch
	assetName := resolveAssetName()
	asset, err := findAsset(release.Assets, assetName)
	if err != nil {
		fmt.Printf("Update failed:\n%s\n", err)
		os.Exit(1)
	}

	fmt.Printf("Downloading %s...\n", asset.Name)

	// Download to a temporary file
	tmpPath, err := downloadAsset(asset.BrowserDownloadURL)
	if err != nil {
		fmt.Printf("Update failed:\n%s\n", err)
		os.Exit(1)
	}
	defer os.Remove(tmpPath)

	// Replace the current executable
	fmt.Println("Replacing executable...")
	err = replaceBinary(tmpPath)
	if err != nil {
		fmt.Printf("Update failed:\n%s\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Printf("✅ Update completed successfully (v%s → v%s).\n", currentVersion, latestVersion)
	fmt.Println("Restart Pengu to use the latest version.")
}

// fetchLatestRelease queries the GitHub Releases API and returns the latest release metadata.
func fetchLatestRelease() (*githubRelease, error) {
	req, err := http.NewRequest("GET", releaseAPI, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "Pengu-CLI/"+version)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not fetch latest release from GitHub\n  %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("no releases found on GitHub repository")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release githubRelease
	err = json.NewDecoder(resp.Body).Decode(&release)
	if err != nil {
		return nil, fmt.Errorf("could not parse GitHub release response: %v", err)
	}

	return &release, nil
}

// resolveAssetName builds the expected binary name for the current OS and architecture.
// Naming convention: pengu-<os>-<arch>[.exe]
func resolveAssetName() string {
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	name := fmt.Sprintf("pengu-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name
}

// findAsset locates the matching asset from a release's asset list.
func findAsset(assets []githubAsset, name string) (*githubAsset, error) {
	for _, a := range assets {
		if a.Name == name {
			return &a, nil
		}
	}

	available := make([]string, len(assets))
	for i, a := range assets {
		available[i] = a.Name
	}
	return nil, fmt.Errorf("no release asset found for your platform (%s/%s)\n  Expected: %s\n  Available: %s",
		runtime.GOOS, runtime.GOARCH, name, strings.Join(available, ", "))
}

// downloadAsset downloads a URL to a temporary file and returns the temp file path.
func downloadAsset(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("download failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "pengu-update-*")
	if err != nil {
		return "", fmt.Errorf("could not create temporary file: %v", err)
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write downloaded binary: %v", err)
	}

	return tmpFile.Name(), nil
}

// replaceBinary safely replaces the currently running executable with the downloaded update.
func replaceBinary(newBinaryPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable path: %v", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve executable path: %v", err)
	}

	// On Windows, we cannot overwrite a running executable directly.
	// Strategy: rename current → .old, move new → current, delete .old
	if runtime.GOOS == "windows" {
		oldPath := execPath + ".old"

		// Clean up any leftover .old file from a previous update
		os.Remove(oldPath)

		// Rename current exe to .old
		err = os.Rename(execPath, oldPath)
		if err != nil {
			return fmt.Errorf("could not rename current executable: %v", err)
		}

		// Copy new binary to the original path
		err = copyFile(newBinaryPath, execPath)
		if err != nil {
			// Rollback: restore original
			os.Rename(oldPath, execPath)
			return fmt.Errorf("could not write new executable: %v", err)
		}

		// Schedule cleanup of .old (best effort — Windows may lock it)
		os.Remove(oldPath)
	} else {
		// On Unix systems: copy new binary over the old one, then set permissions
		err = copyFile(newBinaryPath, execPath)
		if err != nil {
			return fmt.Errorf("could not replace executable: %v", err)
		}

		// Preserve executable permissions
		err = os.Chmod(execPath, 0755)
		if err != nil {
			return fmt.Errorf("could not set executable permissions: %v", err)
		}
	}

	return nil
}

// copyFile copies src to dst, overwriting dst if it exists.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
