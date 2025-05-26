package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CheckDependencies() error {
	toolsPath := filepath.Join("toolssparky")

	// Create toolssparky directory if it doesn't exist
	if _, err := os.Stat(toolsPath); os.IsNotExist(err) {
		if err := os.MkdirAll(toolsPath, 0755); err != nil {
			return fmt.Errorf("failed to create tools directory %s: %v", toolsPath, err)
		}
	}

	// List of system and Go tools to check with appropriate test commands
	tools := []struct {
		name     string
		checkCmd []string
	}{
		{"subfinder", []string{"subfinder", "-h"}},
		{"assetfinder", []string{"assetfinder", "-h"}},
		{"amass", []string{"amass", "-h"}},
		{"httpx", []string{"httpx", "-h"}},
		{"dnsx", []string{"dnsx", "-h"}},
		{"katana", []string{"katana", "-h"}},
		{"waybackurls", []string{"waybackurls", "-h"}},
		{"sqlmap", []string{"sqlmap", "--version"}},
		{"ffuf", []string{"ffuf", "-h"}},
		{"hakrawler", nil},
		{"anew", []string{"anew", "-h"}},
		{"gf", []string{"gf", "-h"}},
		{"nuclei", []string{"nuclei", "-h"}},
		{"nslookup", []string{"nslookup", "-version"}},
		{"whois", []string{"whois", "--version"}},
	}

	for _, tool := range tools {
		if _, err := exec.LookPath(tool.name); err != nil {
			return fmt.Errorf("tool %s not found, run with -id to install", tool.name)
		}

		// Special case for hakrawler: run without flags and check output
		if tool.name == "hakrawler" {
			cmd := exec.Command(tool.name)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if strings.Contains(string(output), "No urls detected. Hint: cat urls.txt | hakrawler") {
					continue // Tool is installed and working as expected
				}
				return fmt.Errorf("tool %s is installed but not working correctly, run with -id to reinstall: %v\nOutput: %s", tool.name, err, string(output))
			}
			continue
		}

		// Test other tools with their respective commands
		cmd := exec.Command(tool.checkCmd[0], tool.checkCmd[1:]...)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tool %s is installed but not working correctly, run with -id to reinstall: %v", tool.name, err)
		}
	}

	// Check virtual environment
	venvPath := filepath.Join(toolsPath, "venv", "bin")
	if _, err := os.Stat(filepath.Join(venvPath, "python3")); os.IsNotExist(err) {
		return fmt.Errorf("virtual environment not found in %s, run with -id to install", venvPath)
	}

	// Check Python tools in virtual environment
	python := filepath.Join(venvPath, "python3")
	pythonTools := []struct {
		name      string
		checkFile string
	}{
		{"linkfinder", "linkfinder.py"},
		{"secretfinder", "SecretFinder.py"},
	}
	for _, tool := range pythonTools {
		toolPath := filepath.Join(toolsPath, tool.name, tool.checkFile)
		cmd := exec.Command(python, toolPath, "-h")
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tool %s not found or not working in virtual environment, run with -id to install: %v", tool.name, err)
		}
	}

	// Check js-beautify (Node.js tool)
	if _, err := exec.LookPath("js-beautify"); err != nil {
		return fmt.Errorf("tool js-beautify not found, run with -id to install")
	}
	cmd := exec.Command("js-beautify", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("tool js-beautify is installed but not working correctly, run with -id to reinstall: %v", err)
	}

	return nil
}

func InstallDependencies() error {
	// Check required base tools
	if _, err := exec.LookPath("go"); err != nil {
		fmt.Println("[*] Go not found. Please install Go manually from https://golang.org")
		return err
	}
	if _, err := exec.LookPath("git"); err != nil {
		fmt.Println("[*] Git not found. Please install Git manually (e.g., sudo apt install git)")
		return err
	}
	if _, err := exec.LookPath("python3"); err != nil {
		fmt.Println("[*] Python3 not found. Please install Python3 (e.g., sudo apt install python3)")
		return err
	}
	// Check for Node.js and npm (required for js-beautify)
	if _, err := exec.LookPath("node"); err != nil {
		fmt.Println("[*] Node.js not found. Please install Node.js manually (e.g., sudo apt install nodejs)")
		return err
	}
	if _, err := exec.LookPath("npm"); err != nil {
		fmt.Println("[*] npm not found. Please install npm manually (e.g., sudo apt install npm)")
		return err
	}

	// Ensure GOPATH is set
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
		os.Setenv("GOPATH", gopath)
		fmt.Println("[*] GOPATH set to", gopath)
	}
	os.Setenv("PATH", fmt.Sprintf("%s:%s/bin", os.Getenv("PATH"), gopath))

	fmt.Println("[*] Installing Go dependencies...")
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		return fmt.Errorf("failed to tidy Go modules: %v", err)
	}

	toolsPath := filepath.Join("toolssparky")
	if err := os.MkdirAll(toolsPath, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %v", err)
	}

	// Install Go-based tools
	goTools := []struct {
		name     string
		url      string
		checkCmd []string
	}{
		{"subfinder", "github.com/projectdiscovery/subfinder/v2/cmd/subfinder@latest", []string{"subfinder", "-h"}},
		{"assetfinder", "github.com/tomnomnom/assetfinder@latest", []string{"assetfinder", "-h"}},
		{"amass", "github.com/owasp-amass/amass/v4/...@master", []string{"amass", "-h"}},
		{"httpx", "github.com/projectdiscovery/httpx/cmd/httpx@latest", []string{"httpx", "-h"}},
		{"dnsx", "github.com/projectdiscovery/dnsx/cmd/dnsx@latest", []string{"dnsx", "-h"}},
		{"katana", "github.com/projectdiscovery/katana/cmd/katana@latest", []string{"katana", "-h"}},
		{"waybackurls", "github.com/tomnomnom/waybackurls@latest", []string{"waybackurls", "-h"}},
		{"ffuf", "github.com/ffuf/ffuf/v2@latest", []string{"ffuf", "-h"}},
		{"hakrawler", "github.com/hakluke/hakrawler@latest", nil},
		{"anew", "github.com/tomnomnom/anew@latest", []string{"anew", "-h"}},
		{"gf", "github.com/tomnomnom/gf@latest", []string{"gf", "-h"}},
		{"nuclei", "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", []string{"nuclei", "-h"}},
	}
	for _, tool := range goTools {
		_, err := exec.LookPath(tool.name)
		if err != nil {
			fmt.Printf("[*] Installing %s...\n", tool.name)
			cmd := exec.Command("go", "install", tool.url)
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install %s: %v", tool.name, err)
			}
			continue
		}

		// Special case for hakrawler: run without flags and check output
		if tool.name == "hakrawler" {
			cmd := exec.Command(tool.name)
			output, err := cmd.CombinedOutput()
			if err != nil {
				if strings.Contains(string(output), "No urls detected. Hint: cat urls.txt | hakrawler") {
					continue // Tool is installed and working as expected
				}
				fmt.Printf("[*] Reinstalling %s (not working correctly)...\n", tool.name)
				cmd = exec.Command("go", "install", tool.url)
				cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("failed to reinstall %s: %v", tool.name, err)
				}
			}
			continue
		}

		// Test other tools with their respective commands
		cmd := exec.Command(tool.checkCmd[0], tool.checkCmd[1:]...)
		err = cmd.Run()
		if err != nil {
			fmt.Printf("[*] Reinstalling %s (not working correctly)...\n", tool.name)
			cmd = exec.Command("go", "install", tool.url)
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to reinstall %s: %v", tool.name, err)
			}
		}
	}

	// Install sqlmap using apt
	_, err := exec.LookPath("sqlmap")
	if err != nil || exec.Command("sqlmap", "--version").Run() != nil {
		fmt.Println("[*] Installing sqlmap...")
		if err := exec.Command("sudo", "apt", "install", "-y", "sqlmap").Run(); err != nil {
			return fmt.Errorf("failed to install sqlmap (try running with sudo or manually): %v", err)
		}
	}

	// Install nslookup and whois (system tools)
	needsNslookup := false
	needsWhois := false

	if _, err := exec.LookPath("nslookup"); err != nil {
		needsNslookup = true
	} else if err := exec.Command("nslookup", "-version").Run(); err != nil {
		needsNslookup = true
	}

	if _, err := exec.LookPath("whois"); err != nil {
		needsWhois = true
	} else if err := exec.Command("whois", "--version").Run(); err != nil {
		needsWhois = true
	}

	if needsNslookup || needsWhois {
		fmt.Println("[*] Installing nslookup and whois...")
		if err := exec.Command("sudo", "apt", "install", "-y", "dnsutils", "whois").Run(); err != nil {
			return fmt.Errorf("failed to install dnsutils and whois (try running with sudo or manually): %v", err)
		}
	}

	// Create a virtual environment for Python tools
	venvPath := filepath.Join(toolsPath, "venv")
	if _, err := os.Stat(filepath.Join(venvPath, "bin", "python3")); os.IsNotExist(err) {
		fmt.Println("[*] Creating virtual environment for Python tools...")
		if err := exec.Command("python3", "-m", "venv", venvPath).Run(); err != nil {
			return fmt.Errorf("failed to create virtual environment: %v", err)
		}
	}

	// Install Python-based tools in the virtual environment
	pip := filepath.Join(venvPath, "bin", "pip3")
	python := filepath.Join(venvPath, "bin", "python3")
	fmt.Println("[*] Upgrading pip in virtual environment...")
	if err := exec.Command(pip, "install", "--upgrade", "pip").Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip in virtual environment: %v", err)
	}

	pythonTools := []struct {
		name      string
		url       string
		checkFile string
	}{
		{"linkfinder", "https://github.com/GerbenJavado/LinkFinder.git", "linkfinder.py"},
		{"secretfinder", "https://github.com/m4ll0k/SecretFinder.git", "SecretFinder.py"},
	}
	for _, tool := range pythonTools {
		toolPath := filepath.Join(toolsPath, tool.name)
		if _, err := os.Stat(toolPath); os.IsNotExist(err) {
			fmt.Printf("[*] Cloning %s...\n", tool.name)
			if err := exec.Command("git", "clone", "--depth", "1", tool.url, toolPath).Run(); err != nil {
				return fmt.Errorf("failed to clone %s: %v", tool.name, err)
			}
		}
		requirementsPath := filepath.Join(toolPath, "requirements.txt")
		if _, err := os.Stat(requirementsPath); err == nil {
			fmt.Printf("[*] Installing requirements for %s...\n", tool.name)
			if err := exec.Command(pip, "install", "-r", requirementsPath).Run(); err != nil {
				return fmt.Errorf("failed to install requirements for %s: %v", tool.name, err)
			}
		}
		// Check for setup.py and run it if exists
		setupPath := filepath.Join(toolPath, "setup.py")
		if _, err := os.Stat(setupPath); err == nil {
			fmt.Printf("[*] Running setup.py for %s...\n", tool.name)
			if err := exec.Command(pip, "install", "setuptools").Run(); err != nil {
				return fmt.Errorf("failed to install setuptools for %s: %v", tool.name, err)
			}
			// Capture detailed output for debugging
			cmd := exec.Command(python, setupPath, "install")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to run setup.py for %s: %v\nOutput: %s", tool.name, err, string(output))
			}
		}
	}

	// Install js-beautify using npm
	_, err = exec.LookPath("js-beautify")
	if err != nil {
		fmt.Println("[*] Installing js-beautify...")
		if err := exec.Command("npm", "install", "-g", "js-beautify").Run(); err != nil {
			return fmt.Errorf("failed to install js-beautify: %v", err)
		}
	} else {
		// Test if js-beautify works, reinstall if it doesn't
		cmd := exec.Command("js-beautify", "--version")
		if err := cmd.Run(); err != nil {
			fmt.Println("[*] Reinstalling js-beautify (not working correctly)...")
			if err := exec.Command("npm", "install", "-g", "js-beautify").Run(); err != nil {
				return fmt.Errorf("failed to reinstall js-beautify: %v", err)
			}
		}
	}

	fmt.Println("[*] Dependencies installed successfully")
	return nil
}
