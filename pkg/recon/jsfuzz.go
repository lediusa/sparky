package recon

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/lediusa/sparky/pkg/config"
	"strings"
)

func JSFuzzing(urlsFile, outputFile string, cfg *config.Config) ([]string, error) {
	file, err := os.Open(urlsFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Extract JS URLs
	jsURLs := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if strings.HasSuffix(url, ".js") {
			jsURLs = append(jsURLs, url)
		}
	}

	// Extract unique JS paths
	jsPaths := make(map[string]bool)
	for _, url := range jsURLs {
		parts := strings.Split(url, "/")
		if len(parts) < 4 {
			continue
		}
		path := strings.Join(parts[3:len(parts)-1], "/")
		if path != "" {
			jsPaths[path] = true
		}
	}

	// Fuzz each path
	var newJSFiles []string
	for path := range jsPaths {
		for _, jsURL := range jsURLs {
			if strings.Contains(jsURL, path) {
				baseURL := strings.Split(jsURL, "/")
				domain := strings.Join(baseURL[:3], "/")
				fuzzURL := fmt.Sprintf("%s/%s/FUZZ", domain, path)
				cmd := exec.Command("ffuf", "-w", cfg.Paths.JsSmartFuzzing, "-u", fuzzURL, "-mc", "200", "-o", outputFile+".tmp", "-of", "csv")
				if err := cmd.Run(); err != nil {
					continue
				}

				// Parse ffuf output
				data, err := os.ReadFile(outputFile + ".tmp")
				if err != nil {
					continue
				}
				lines := strings.Split(string(data), "\n")
				for _, line := range lines[1:] {
					if line == "" {
						continue
					}
					cols := strings.Split(line, ",")
					if len(cols) < 2 {
						continue
					}
					fuzzedURL := fmt.Sprintf("%s/%s/%s", domain, path, cols[0])
					newJSFiles = append(newJSFiles, fuzzedURL)
				}
			}
		}
	}

	// Save new JS files
	if len(newJSFiles) > 0 {
		os.WriteFile(outputFile, []byte(strings.Join(newJSFiles, "\n")), 0644)
	}

	return newJSFiles, nil
}
