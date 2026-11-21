package updater

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	RepoOwner = "vknow360"
	RepoName  = "FreeTab"
)

// GitHubRelease represents a release object returned by the GitHub API.
type GitHubRelease struct {
	TagName     string        `json:"tag_name"`
	Name        string        `json:"name"`
	Body        string        `json:"body"`
	PublishedAt string        `json:"published_at"`
	HTMLURL     string        `json:"html_url"`
	Assets      []GitHubAsset `json:"assets"`
}

// GitHubAsset represents a downloadable asset in a GitHub release.
type GitHubAsset struct {
	Name        string `json:"name"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"browser_download_url"`
}

// UpdateInfo contains metadata about available updates.
type UpdateInfo struct {
	CurrentVersion  string `json:"currentVersion"`
	LatestVersion   string `json:"latestVersion"`
	UpdateAvailable bool   `json:"updateAvailable"`
	ReleaseName     string `json:"releaseName"`
	ReleaseNotes    string `json:"releaseNotes"`
	ReleaseURL      string `json:"releaseUrl"`
	DownloadURL     string `json:"downloadUrl"`
	AssetName       string `json:"assetName"`
	AssetSize       int64  `json:"assetSize"`
	StatusMessage   string `json:"statusMessage"`
}

// FetchLatestRelease fetches the latest GitHub release for the repository.
func FetchLatestRelease(client *http.Client) (*GitHubRelease, error) {
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", RepoOwner, RepoName)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", fmt.Sprintf("%s-Updater", RepoName))

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error checking for updates: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no release found on GitHub")
	}

	if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("GitHub API rate limit reached. Please try again later.")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release GitHubRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release metadata: %w", err)
	}

	return &release, nil
}

// CompareVersions returns true if latest is strictly newer than current.
func CompareVersions(current, latest string) bool {
	current = strings.TrimPrefix(strings.TrimSpace(current), "v")
	latest = strings.TrimPrefix(strings.TrimSpace(latest), "v")

	curParts := strings.Split(current, ".")
	latParts := strings.Split(latest, ".")

	maxLen := len(curParts)
	if len(latParts) > maxLen {
		maxLen = len(latParts)
	}

	for i := 0; i < maxLen; i++ {
		var curNum, latNum int
		if i < len(curParts) {
			curNum, _ = strconv.Atoi(curParts[i])
		}
		if i < len(latParts) {
			latNum, _ = strconv.Atoi(latParts[i])
		}

		if latNum > curNum {
			return true
		}
		if latNum < curNum {
			return false
		}
	}

	return false
}

// FindAsset locates the best binary asset for Windows in a release.
func FindAsset(release *GitHubRelease) (*GitHubAsset, error) {
	if len(release.Assets) == 0 {
		return nil, fmt.Errorf("release %s contains no downloadable assets", release.TagName)
	}

	// 1. Exact match "FreeTab.exe"
	for _, a := range release.Assets {
		if strings.EqualFold(a.Name, "FreeTab.exe") {
			return &a, nil
		}
	}

	// 2. Contains "freetab" and ends in .exe
	for _, a := range release.Assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, "freetab") && strings.HasSuffix(lower, ".exe") {
			return &a, nil
		}
	}

	// 3. Any .exe asset
	for _, a := range release.Assets {
		if strings.HasSuffix(strings.ToLower(a.Name), ".exe") {
			return &a, nil
		}
	}

	// 4. Any .zip asset containing "freetab"
	for _, a := range release.Assets {
		lower := strings.ToLower(a.Name)
		if strings.Contains(lower, "freetab") && strings.HasSuffix(lower, ".zip") {
			return &a, nil
		}
	}

	return nil, fmt.Errorf("no compatible Windows executable or zip found in release %s", release.TagName)
}

// CheckForUpdate checks if an update is available on GitHub.
func CheckForUpdate(currentVersion string, logFn func(string)) (*UpdateInfo, error) {
	info := &UpdateInfo{
		CurrentVersion: currentVersion,
	}

	if logFn != nil {
		logFn(fmt.Sprintf("[INFO] Checking for updates on GitHub (current: v%s)...", currentVersion))
	}

	release, err := FetchLatestRelease(nil)
	if err != nil {
		info.StatusMessage = err.Error()
		if logFn != nil {
			logFn(fmt.Sprintf("[WARN] Update check: %s", err.Error()))
		}
		return info, err
	}

	latestVer := strings.TrimPrefix(release.TagName, "v")
	info.LatestVersion = latestVer
	info.ReleaseName = release.Name
	info.ReleaseNotes = release.Body
	info.ReleaseURL = release.HTMLURL

	if !CompareVersions(currentVersion, latestVer) {
		info.UpdateAvailable = false
		info.StatusMessage = fmt.Sprintf("FreeTab is up to date (v%s)", currentVersion)
		if logFn != nil {
			logFn(fmt.Sprintf("FreeTab is up to date (v%s)", currentVersion))
		}
		return info, nil
	}

	asset, err := FindAsset(release)
	if err != nil {
		info.UpdateAvailable = false
		info.StatusMessage = fmt.Sprintf("New version v%s available, but: %v", latestVer, err)
		if logFn != nil {
			logFn(fmt.Sprintf("[WARN] %s", info.StatusMessage))
		}
		return info, nil
	}

	info.UpdateAvailable = true
	info.DownloadURL = asset.DownloadURL
	info.AssetName = asset.Name
	info.AssetSize = asset.Size
	info.StatusMessage = fmt.Sprintf("Update available: v%s → v%s", currentVersion, latestVer)

	if logFn != nil {
		logFn(fmt.Sprintf("🎉 %s (Release: %s)", info.StatusMessage, release.Name))
	}

	return info, nil
}

// DownloadAndInstall downloads the update asset and replaces the running executable.
func DownloadAndInstall(downloadURL string, assetName string, logFn func(string)) error {
	if logFn != nil {
		logFn(fmt.Sprintf("[INFO] Downloading update from: %s", downloadURL))
	}

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download server returned HTTP %d", resp.StatusCode)
	}

	// Create temp file for downloaded asset
	tmpFile, err := os.CreateTemp("", "freetab-download-*"+filepath.Ext(assetName))
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpDownloadPath := tmpFile.Name()
	defer os.Remove(tmpDownloadPath)

	written, err := io.Copy(tmpFile, resp.Body)
	tmpFile.Close()
	if err != nil {
		return fmt.Errorf("failed to save download data: %w", err)
	}

	if logFn != nil {
		logFn(fmt.Sprintf("[INFO] Downloaded %.2f MB. Preparing installation...", float64(written)/(1024*1024)))
	}

	binaryPath := tmpDownloadPath

	// If the asset is a .zip archive, extract the executable
	if strings.HasSuffix(strings.ToLower(assetName), ".zip") {
		extractedPath, err := extractExecutableFromZip(tmpDownloadPath)
		if err != nil {
			return fmt.Errorf("failed to extract executable from zip: %w", err)
		}
		defer os.Remove(extractedPath)
		binaryPath = extractedPath
	}

	// Replace the currently running executable
	if logFn != nil {
		logFn("[INFO] Replacing current executable with updated version...")
	}

	if err := ReplaceExecutable(binaryPath); err != nil {
		return fmt.Errorf("installation failed: %w", err)
	}

	if logFn != nil {
		logFn("🎉 Update installed successfully! Application is ready to restart.")
	}

	return nil
}

// extractExecutableFromZip finds and extracts the .exe from a downloaded zip file.
func extractExecutableFromZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var targetFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".exe") {
			targetFile = f
			if strings.EqualFold(filepath.Base(f.Name), "FreeTab.exe") {
				break
			}
		}
	}

	if targetFile == nil {
		return "", fmt.Errorf("no executable found inside the downloaded zip archive")
	}

	rc, err := targetFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	tmpExe, err := os.CreateTemp("", "freetab-extracted-*.exe")
	if err != nil {
		return "", err
	}
	defer tmpExe.Close()

	if _, err := io.Copy(tmpExe, rc); err != nil {
		os.Remove(tmpExe.Name())
		return "", err
	}

	return tmpExe.Name(), nil
}

// ReplaceExecutable safely replaces the current running binary on Windows.
func ReplaceExecutable(newBinaryPath string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not determine current executable path: %w", err)
	}

	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("could not resolve executable symlink: %w", err)
	}

	// Ensure staging file is on the exact same drive/volume so os.Rename is atomic
	execDir := filepath.Dir(execPath)
	stagingPath := filepath.Join(execDir, fmt.Sprintf("FreeTab_staging_%d.exe", time.Now().UnixNano()))

	srcFile, err := os.Open(newBinaryPath)
	if err != nil {
		return fmt.Errorf("cannot read new binary: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(stagingPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("cannot create staging file in app directory: %w", err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		dstFile.Close()
		os.Remove(stagingPath)
		return fmt.Errorf("cannot copy new binary to app directory: %w", err)
	}
	dstFile.Close()

	oldPath := execPath + ".old"
	_ = os.Remove(oldPath) // remove previous backup if any

	// Rename running executable to .old
	if err := os.Rename(execPath, oldPath); err != nil {
		os.Remove(stagingPath)
		return fmt.Errorf("cannot rename running executable: %w", err)
	}

	// Rename staged binary into final executable location
	if err := os.Rename(stagingPath, execPath); err != nil {
		// Rollback
		_ = os.Rename(oldPath, execPath)
		os.Remove(stagingPath)
		return fmt.Errorf("cannot install new executable: %w", err)
	}

	return nil
}

// RestartApp restarts the updated executable and exits the current process.
func RestartApp() error {
	execPath, err := os.Executable()
	if err != nil {
		return err
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return err
	}

	cmd := exec.Command(execPath)
	cmd.Dir = filepath.Dir(execPath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch updated process: %w", err)
	}

	// Exit current process
	os.Exit(0)
	return nil
}
