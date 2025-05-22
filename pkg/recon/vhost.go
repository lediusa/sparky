```go
package recon

import (
	"bufio"
	"os"
	"os/exec"
	"sparky/pkg/config"
	"strings"
)

func VhostDiscovery(domain string, ips []string, outputFile string, cfg *config.Config) error {
	for _, ip := range ips {
		cmd := exec.Command("ffuf", "-w", cfg.Paths.Subdomains, "-u", "http://"+ip+"/", "-H", "Host: FUZZ."+domain, "-o", outputFile, "-fc", "404")
		if err := cmd.Run(); err != nil {
			return err
		}
	}
	return nil
}

func CompareSubdomains(activeSubdomains, vhostFile, newSubdomainsFile string) ([]string, error) {
	cmd := exec.Command("grep", "-oE", "[a-zA-Z0-9.-]+\\."+strings.Split(activeSubdomains, "/")[len(strings.Split(activeSubdomains, "/"))-2], vhostFile)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	vhostSubs := strings.Split(string(out), "\n")
	activeSubs := make(map[string]bool)
	file, err := os.Open(activeSubdomains)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		activeSubs[scanner.Text()] = true
	}

	var newSubs []string
	for _, sub := range vhostSubs {
		if sub != "" && !activeSubs[sub] {
			newSubs = append(newSubs, sub)
		}
	}

	if len(newSubs) > 0 {
		os.WriteFile(newSubdomainsFile, []byte(strings.Join(newSubs, "\n")), 0644)
	}
	return newSubs, nil
}
```
