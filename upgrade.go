package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
)

type Release struct {
	TagName string `json:"tag_name"`
	HTMLURL string `json:"html_url"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// Fetches release info from github api
func fetchRelease() (*Release, error) {
	release := &Release{}
	resp, err := http.Get("https://api.github.com/repos/godwhoa/oodle/releases/latest")
	if err != nil {
		return release, fmt.Errorf("Failed to github api: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return release, json.Unmarshal(body, release)
}

var archAsset = map[string]string{
	"amd64": "oodle_linux",
	"arm":   "oodle_linux_armv7",
}

// Extracts asset download url
func extractAsset(r *Release) (string, error) {
	for _, asset := range r.Assets {
		if asset.Name == archAsset[runtime.GOARCH] {
			return asset.BrowserDownloadURL, nil
		}
	}
	return "", fmt.Errorf("Could not find the asset \"oodle_linux\"")
}

// Downloads a file to tmp dir. then returns file name.
func downloadTmp(url string) (string, error) {
	cdir, _ := os.Getwd()
	f, err := ioutil.TempFile(cdir, "oodle")
	if err != nil {
		return "", fmt.Errorf("Failed to create tmp file: %v", err)
	}
	defer f.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("Failed to download: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	if _, err = io.Copy(f, resp.Body); err != nil {
		return "", fmt.Errorf("Error while downloading: %v", err)
	}

	return f.Name(), nil
}

// Upgrades the binary
// 1. fetch github api
// 2. extract download url
// 3. download it to tmp dir.
// 4. get running exe.'s path
// 5. rename tmp to that
// 6. remove tmp
// 7. chmod +x
func upgrade() error {
	release, err := fetchRelease()
	if err != nil {
		return err
	}

	downloadURL, err := extractAsset(release)
	if err != nil {
		return err
	}

	fmt.Println("Downloading " + downloadURL)
	newoodlePath, err := downloadTmp(downloadURL)
	if err != nil {
		return err
	}

	fmt.Println("Upgrading...")
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("Could not find executable path: %v", err)
	}
	if err := os.Rename(newoodlePath, exePath); err != nil {
		return err
	}
	os.Remove(newoodlePath)
	exec.Command("chmod", "+x", exePath).Output()
	fmt.Printf("Upgraded to %s! Patched executable: %s\n", release.TagName, exePath)
	return nil
}
