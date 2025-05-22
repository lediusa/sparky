```go
package recon

import (
	"os"
	"os/exec"
	"sparky/pkg/config"
)

func SubdomainDiscovery(domain, outputFile string, cfg *config.Config) error {
	tools := []struct {
		name string
		args []string
	}{
		{"subfinder", []string{"-d", domain, "-all", "-silent"}},
		{"assetfinder", []string{"--subs-only", domain}},
		{"amass", []string{"enum", "-d", domain}},
	}
	for _, tool := range tools {
		cmd := exec.Command(tool.name, tool.args...)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		os.WriteFile(outputFile+"."+tool.name, out, 0644)
	}

	cmd := exec.Command("sh", "-c", "cat "+outputFile+".* | sort -u > "+outputFile)
	if err := cmd.Run(); err != nil {
		return err
	}

	os.RemoveAll(outputFile + ".*")
	return nil
}
```
