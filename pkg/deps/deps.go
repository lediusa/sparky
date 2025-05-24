package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckDependencies() error {
	toolsPath := filepath.Join("toolssparky")

	// List of system and Go tools to check
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
		{"hakrawler", []string{"hakrawler", "-h"}},
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
		name string
		url  string
	}{
		{"subfinder", "github.com/projectdiscovery/subfinder/cmd/subfinder@latest"},
		{"assetfinder", "github.com/tomnomnom/assetfinder@latest"},
		{"amass", "github.com/OWASP/Amass/v3/...@latest"},
		{"httpx", "github.com/projectdiscovery/httpx/cmd/httpx@latest"},
		{"dnsx", "github.com/projectdiscovery/dnsx/cmd/dnsx@latest"},
		{"katana", "github.com/projectdiscovery/katana/cmd/katana@latest"},
		{"waybackurls", "github.com/tomnomnom/waybackurls@latest"},
		{"ffuf", "github.com/ffuf/ffuf@latest"},
		{"hakrawler", "github.com/hakluke/hakrawler@latest"},
		{"anew", "github.com/tomnomnom/anew@latest"},
		{"gf", "github.com/tomnomnom/gf@latest"},
		{"nuclei", "github.com/projectdiscovery/nuclei/v2/cmd/nuclei@latest"},
	}
	for _, tool := range goTools {
		// Check if the tool is installed and working
		var cmd *exec.Cmd
		if tool.name == "assetfinder" || tool.name == "anew" || tool.name == "waybackurls" {
			cmd = exec.Command(tool.name, "-h")
		} else {
			cmd = exec.Command(tool.name, "--version")
		}
		if _, err := exec.LookPath(tool.name); err == nil && cmd.Run() == nil {
			continue // Skip if already installed and working
		}

		fmt.Printf("[*] Installing %s...\n", tool.name)
		cmd = exec.Command("go", "install", tool.url)
		cmd.Env = append(os.Environ(), fmt.Sprintf("GOPATH=%s", gopath))
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install %s: %v", tool.name, err)
		}
		// Verify installation
		if _, err := exec.LookPath(tool.name); err != nil {
			return fmt.Errorf("tool %s was installed but not found in PATH", tool.name)
		}
	}

	// Install sqlmap using apt
	if _, err := exec.LookPath("sqlmap"); err != nil || exec.Command("sqlmap", "--version").Run() != nil {
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
			fmt.Printf("[*] Installing requirements for %s...\n", tool.name)
			if err := exec.Command(pip, "install", "-r", filepath.Join(toolPath, "requirements.txt")).Run(); err != nil {
				return fmt.Errorf("failed to install requirements for %s: %v", tool.name, err)
			}
			// Run setup.py for linkfinder
			if tool.name == "linkfinder" {
				fmt.Println("[*] Running setup.py for linkfinder...")
				if err := exec.Command(python, filepath.Join(toolPath, "setup.py"), "install").Run(); err != nil {
					return fmt.Errorf("failed to run setup.py for %s: %v", tool.name, err)
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
