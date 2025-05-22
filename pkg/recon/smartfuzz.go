```go
package recon

import (
	"bufio"
	"os"
	"os/exec"
	"sparky/pkg/config"
	"strings"
)

func SmartFuzzing(inputFile, fuzzFile, fuzz200File string, cfg *config.Config) ([]string, error) {
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var results []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		subdomain := scanner.Text()
		if subdomain == "" {
			continue
		}
		cmd := exec.Command("ffuf", "-w", cfg.Paths.BackupLogin, "-u", "https://"+subdomain+"/FUZZ", "-o", fuzzFile, "-fc", "403,404")
		if err := cmd.Run(); err != nil {
			continue
		}

		cmd = exec.Command("grep", "-oE", "https://"+subdomain+"/[a-zA-Z0-9./_-]+", fuzzFile)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		results = append(results, strings.Split(string(out), "\n")...)
	}

	if len(results) > 0 {
		os.WriteFile(fuzz200File, []byte(strings.Join(results, "\n")), 0644)
	}
	return results, nil
}
```
