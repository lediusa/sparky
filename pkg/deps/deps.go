```go
package deps

import (
	"fmt"
	"os"
	"os/exec"
)

func CheckDependencies() error {
	tools := []string{
		"subfinder", "assetfinder", "amass", "httpx", "dnsx", "jsbeautifier",
		"katana", "waybackurls", "sqlmap", "ffuf", "hakrawler", "anew",
		"gf", "nuclei", "nslookup", "whois",
	}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("tool %s not found, run with -id", tool)
		}
	}

	if _, err := os.Stat("toolssparky/linkfinder/LinkFinder.py"); os.IsNotExist(err) {
		return fmt.Errorf("LinkFinder not found, run with -id")
	}
	if _, err := os.Stat("toolssparky/SecretFinder/SecretFinder.py"); os.IsNotExist(err) {
		return fmt.Errorf("SecretFindermill not found, run with -id")
	}

	return nil
}

func InstallDependencies() error {
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("[*] Go not found. Please install Go manually from https://golang.org")
		return err
	}

	fmt.Println("[*] Installing Go dependencies...")
	cmd := exec.Command("go", "mod", "tidy")
	if err := cmd.Run(); err != nil {
		return err
	}

	tools := []struct {
		name string
		url  string
	}{
		{"linkfinder", "https://github.com/GerbenJavado/LinkFinder.git"},
		{"SecretFinder", "https://github.com/m4ll0k/SecretFinder.git"},
	}
	for _, tool := range tools {
		if _, err := os.Stat("toolssparky/" + tool.name); os.IsNotExist(err) {
			fmt.Printf("[*] Installing %s...\n", tool.name)
			cmd := exec.Command("git", "clone", tool.url, "toolssparky/"+tool.name)
			if err := cmd.Run(); err != nil {
				return err
			}
			if tool.name == "linkfinder" || tool.name == "SecretFinder" {
				cmd := exec.Command("pip3", "install", "-r", "toolssparky/"+tool.name+"/requirements.txt")
				if err := cmd.Run(); err != nil {
					return err
				}
			}
		}
	}

	fmt.Println("[*] Building Sparky...")
	cmd = exec.Command("go", "build", "-o", "sparky", "cmd/sparky/main.go")
	if err := cmd.Run(); err != nil {
		return err
	}

	fmt.Println("[*] Moving Sparky to /usr/local/bin...")
	cmd = exec.Command("sudo", "mv", "sparky", "/usr/local/bin/sparky")
	if err := cmd.Run(); err != nil {
		fmt.Println("[*] Failed to move to /usr/local/bin. Please move manually: sudo mv sparky /usr/local/bin/sparky")
		return err
	}

	fmt.Println("[*] Ensure other tools are installed manually: subfinder, ffuf, etc.")
	return nil
}
```
