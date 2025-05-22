package recon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type WCDResult struct {
	URL            string `json:"url"`
	Cached         bool   `json:"cached"`
	CacheHeaders   map[string]string `json:"cache_headers"`
}

func TestWebCacheDeception(urlsFile, outputFile string) ([]WCDResult, error) {
	file, err := os.Open(urlsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Collect static files
	staticFiles := []string{}
	staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".woff", ".woff2", ".ttf", ".ico"}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		for _, ext := range staticExts {
			if strings.HasSuffix(url, ext) {
				staticFiles = append(staticFiles, url)
				break
			}
		}
	}

	// Test each static file for caching
	var results []WCDResult
	for _, url := range staticFiles {
		cmd := exec.Command("curl", "-s", "-I", url)
		out, err := cmd.Output()
		if err != nil {
			continue
		}

		headers := make(map[string]string)
		cached := false
		lines := strings.Split(string(out), "\n")
		for _, line := range lines {
			if strings.Contains(line, ":") {
				parts := strings.SplitN(line, ":", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					value := strings.TrimSpace(parts[1])
					headers[key] = value
					if key == "Cache-Control" && (strings.Contains(value, "public") || strings.Contains(value, "max-age")) {
						cached = true
					}
					if key == "ETag" || key == "Last-Modified" || key == "Age" {
						cached = true
					}
				}
			}
		}

		result := WCDResult{
			URL:          url,
			Cached:       cached,
			CacheHeaders: headers,
		}
		results = append(results, result)
	}

	// Save results to JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return nil, err
	}

	// Return only cached files
	var cachedFiles []WCDResult
	for _, result := range results {
		if result.Cached {
			cachedFiles = append(cachedFiles, result)
		}
	}
	return cachedFiles, nil
}
