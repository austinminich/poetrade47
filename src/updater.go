package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"poetrade47/src/helpers"
	"runtime"
	"strings"
	"time"
)

const (
	repoOwner = "austinminich"
	repoName  = "poetrade47"
	version   = "dev" // default version
)

type gitRelease struct {
	TagName    string     `json:"tag_name"`
	Prerelease bool       `json:"prerelease"`
	Assets     []struct { // Digest is checksum for downloads
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Digest             string `json:"digest"`
	} `json:"assets"`
}

func CleanupOldVersions() {
	// Need to consider if it's an .appimage as well
	exe, err := os.Executable()
	if err != nil {
		helpers.DebugLog("Error getting executable path: %v", err)
		return
	}

	_ = os.Remove(exe + ".old")
}

func CheckForUpdates() error {
	println("Cleaning old versions... (Windows L)")
	CleanupOldVersions()
	println("Checking for updates...")
	client := &http.Client{Timeout: 10 * time.Second}

	// Fetch the latest release information from GitHub
	url := "https://api.github.com/repos/" + repoOwner + "/" + repoName + "/releases"
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("Error fetching latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code: %d", resp.StatusCode)
	}

	var release gitRelease // TODO: There's a possibility that a pre-release version is older than the newest stable release
	if allowPrereleases(version) {
		// If the current version is a pre-release, we want to check for the latest pre-release as well
		var releases []gitRelease
		err = json.NewDecoder(resp.Body).Decode(&releases)
		if err != nil {
			return fmt.Errorf("Error decoding releases JSON: %w", err)
		}

		for _, r := range releases {
			if r.Prerelease && r.TagName != version {
				release = r
				break
			}
		}
	} else {
		// We are not a pre-release version, so we only want to check for the latest stable release
		err = json.NewDecoder(resp.Body).Decode(&release)
		if err != nil {
			return fmt.Errorf("Error decoding release JSON: %w", err)
		}
	}

	// Compare versions
	if release.TagName == "" || release.TagName == version {
		println("No updates available. Current version:", version)
		return nil
	}
	println("Update available. New version:", release.TagName)

	// Find asset matching OS version
	expectedAssetName := "poetrade47"
	switch runtime.GOOS {
	case "windows":
		expectedAssetName += ".exe"
	case "linux":
		expectedAssetName += ".AppImage"
	}

	var downloadURL, expectedHash string
	for _, asset := range release.Assets {
		if asset.Name == expectedAssetName {
			downloadURL = asset.BrowserDownloadURL
			expectedHash = strings.TrimPrefix(asset.Digest, "sha256:")
			break
		}
	}

	if downloadURL == "" {
		return fmt.Errorf("No suitable update found for OS: %s", runtime.GOOS)
	}
	println("Downloading update from:", downloadURL)

	// Download new file
	if version == "dev" { // Dev version
		println("Dev version detected, skipping download.")
		return nil
	}
	resp, err = client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("Error downloading update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Unexpected status code while downloading update: %d", resp.StatusCode)
	}

	// Verify checksum
	if expectedHash == "" {
		return fmt.Errorf("Asset is missing a sha256 digest")
	}

	hash, err := getLocalBinaryHash() // TODO: Test
	if err != nil {
		return fmt.Errorf("Error calculating local binary hash: %w", err)
	}
	if hash != expectedHash {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, hash)
	}
	return nil
}

func UpdatePopUp() {

}

// getLocalBinaryHash returns the lowercase hex SHA256 hash of the running binary or AppImage.
func getLocalBinaryHash() (string, error) {
	targetPath, err := getTargetBinPath()
	if err != nil {
		return "", fmt.Errorf("could not locate binary: %w", err)
	}

	f, err := os.Open(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to open local binary: %w", err)
	}
	defer f.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, f); err != nil {
		return "", fmt.Errorf("failed to hash local binary: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func replaceBinary(exePath, newFilePath string) error {
	oldPath := exePath + ".old"
	_ = os.Remove(oldPath) // TODO: Check why this is done twice

	if err := os.Rename(exePath, oldPath); err != nil {
		return fmt.Errorf("Error renaming current executable: %w", err)
	}

	if err := os.Rename(newFilePath, exePath); err != nil {
		// Attempt to restore the old executable if the rename fails
		_ = os.Rename(oldPath, exePath)
		return fmt.Errorf("Error replacing executable with new version: %w", err)
	}

	// Linux can do this immediately, windows will clean old files on next launch
	_ = os.Remove(oldPath)
	return nil
}

func getTargetBinPath() (string, error) {
	// 1. Check if we're running appimage (linux)
	if appImagePath := os.Getenv("APPIMAGE"); appImagePath != "" {
		return appImagePath, nil
	} else {
		exePath, err := os.Executable()
		if err != nil {
			return "", fmt.Errorf("Error getting executable path: %w", err)
		}
		return filepath.EvalSymlinks(exePath)
	}
}

func allowPrereleases(ver string) bool {
	return ver == "dev" || strings.Contains(ver, "-")
}
