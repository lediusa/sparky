package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckDependencies() error {
	toolsPath := filepath.Join("toolssparky")

	// List of system and Go tools to check with appropriate test commands
	tools := []struct {
		name     string
		checkCmd []string
	}{
		{"subfinder", []string{"subfinder", "--version"}},
		{"assetfinder", []string{"assetfinder", "-h"}},
		{"amass", []string{"amass", "-version"}},
		{"httpx", []string{"httpx", "--version"}},
		{"dnsx", []string{"dnsx", "--version"}},
		{"katana", []string{"katana", "--version"}},
		{"waybackurls", []string{"waybackurls", "-h"}},
		{"sqlmap", []string{"sqlmap", "--version"}},
		{"ffuf", []string{"ffuf", "-h"}},
		{"hakrawler", []string{"hakrawler", "-h"}}, // Use -h with empty arg to avoid error
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
		// Test if the tool actually works
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
		{"jsbeautifier", "jsbeautifier"},
	}
	for _, tool := range pythonTools {
		// Check if the module is installed by trying to run a simple command
		var cmd *exec.Cmd
		if tool.name == "jsbeautifier" {
			cmd = exec.Command(python, "-m", "jsbeautifier", "--version")
		} else {
			cmd = exec.Command(python, "-c", fmt.Sprintf("import %s", tool.name))
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("tool %s not found in virtual environment, run with -id to install: %v", tool.name, err)
		}
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
		{"hakrawler", "github.com/hakluke/hakrawler@latest", []string{"hakrawler", "-h"}},
		{"anew", "github.com/tomnomnom/anew@latest", []string{"anew", "-h"}},
		{"gf", "github.com/tomnomnom/gf@latest", []string{"gf", "-h"}},
		{"nuclei", "github.com/projectdiscovery/nuclei/v3/cmd/nuclei@latest", []string{"nuclei", "-h"}},
	}
	for _, tool := range goTools {
		// Check if the tool is installed and working
		_, err := exec.LookPath(tool.name)
		if err != nil {
			// If command not found, install it
			fmt.Printf("[*] Installing %s...\n", tool.name)
			cmd := exec.Command("go", "install", tool.url)
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install %s: %v", tool.name, err)
			}
			// Verify installation
			if _, err := exec.LookPath(tool.name); err != nil {
				return fmt.Errorf("tool %s was installed but not found in PATH", tool.name)
			}
			continue
		}

		// If the tool exists, test if it works
		cmd := exec.Command(tool.checkCmd[0], tool.checkCmd[1:]...)
		err = cmd.Run()
		if err != nil {
			// If the tool doesn't work, reinstall it
			fmt.Printf("[*] Reinstalling %s (not working correctly)...\n", tool.name)
			cmd := exec.Command("go", "install", tool.url)
			cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to reinstall %s: %v", tool.name, err)
			}
			// Verify reinstallation
			if _, err := exec.LookPath(tool.name); err != nil {
				return fmt.Errorf("tool %s was reinstalled but not found in PATH", tool.name)
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
	// Upgrade pip in the virtual environment to avoid dependency issues
	fmt.Println("[*] Upgrading pip in virtual environment...")
	if err := exec.Command(pip, "install", "--upgrade", "pip").Run(); err != nil {
		return fmt.Errorf("failed to upgrade pip in virtual environment: %v", err)
	}

	pythonTools := []struct {
		name      string
		url       string
		pipName   string // For tools installed directly via pip
		checkFile string // For tools that have a script file
	}{
		{"linkfinder", "https://github.com/GerbenJavado/LinkFinder.git", "", "linkfinder.py"},
		{"secretfinder", "https://github.com/m4ll0k/SecretFinder.git", "", "SecretFinder.py"},
		{"jsbeautifier", "", "jsbeautifier", ""},
	}
	for _, tool := range pythonTools {
		// Check if the tool is installed and working
		var cmd *exec.Cmd
		if tool.name == "jsbeautifier" {
			cmd = exec.Command(python, "-m", "jsbeautifier", "--version")
		} else {
			cmd = exec.Command(python, "-c", fmt.Sprintf("import %s", tool.name))
		}
		if cmd.Run() == nil {
			continue // Skip if already installed and working
		}

		if tool.url != "" {
			toolPath := filepath.Join(toolsPath, tool.name)
			if _, err := os.Stat(toolPath); os.IsNotExist(err) {
				fmt.Printf("[*] Cloning %s...\n", tool.name)
				if err := exec.Command("git", "clone", "--depth", "1", tool.url, toolPath).Run(); err != nil {
					return fmt.Errorf("failed to clone %s: %v", tool.name, err)
				}
			}
			// Install dependencies
			requirementsPath := filepath.Join(toolPath, "requirements.txt")
			if _, err := os.Stat(requirementsPath); err == nil {
				fmt.Printf("[*] Installing requirements for %s...\n", tool.name)
				if err := exec.Command(pip, "install", "-r", requirementsPath).Run(); err != nil {
					return fmt.Errorf("failed to install requirements for %s: %v", tool.name, err)
				}
			}
			// Run setup.py for linkfinder
			if tool.name == "linkfinder" {
				setupPath := filepath.Join(toolPath, "setup.py")
				if _, err := os.Stat(setupPath); err == nil {
					fmt.Println("[*] Running setup.py for linkfinder...")
					// Install setuptools to ensure setup.py runs correctly
					if err := exec.Command(pip, "install", "setuptools").Run(); err != nil {
						return fmt.Errorf("failed to install setuptools for linkfinder: %v", err)
					}
					if err := exec.Command(python, setupPath, "install").Run(); err != nil {
						// Capture more detailed error output
						cmd := exec.Command(python, setupPath, "install")
						output, err := cmd.CombinedOutput()
						return fmt.Errorf("failed to run setup.py for %s: %v\nOutput: %s", tool.name, err, string(output))
					}
				}
			}
		} else if tool.pipName != "" {
			// Install tools like jsbeautifier directly via pip
			fmt.Printf("[*] Installing %s...\n", tool.name)
			if err := exec.Command(pip, "install", tool.pipName).Run(); err != nil {
				return fmt.Errorf("failed to install %s: %v", tool.name, err)
			}
		}
	}

	fmt.Println("[*] Dependencies installed successfully")
	return nil
}
