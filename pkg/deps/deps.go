package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckDependencies() error {
	// List of tools to check
	tools := []string{
		"subfinder", "assetfinder", "amass", "httpx", "dnsx", "jsbeautifier",
		"katana", "waybackurls", "sqlmap", "ffuf", "hakrawler", "anew",
		"gf", "nuclei", "nslookup", "whois",
	}
	for _, tool := range tools {
		if _, err := exec.LookPath(tool); err != nil {
			return fmt.Errorf("tool %s not found, run with -id to install", tool)
		}
		// Try running the tool to ensure it works
		if err := exec.Command(tool, "--version").Run(); err != nil {
			return fmt.Errorf("tool %s is installed but not working correctly, run with -id to reinstall: %v", tool, err)
		}
	}

	// Check virtual environment
	venvPath := filepath.Join("toolssparky", "venv", "bin")
	if _, err := os.Stat(filepath.Join(venvPath, "python3")); os.IsNotExist(err) {
		return fmt.Errorf("virtual environment not found in %s, run with -id to install", venvPath)
	}

	// Check Python tools in virtual environment
	python := filepath.Join(venvPath, "python3")
	if err := exec.Command(python, "-c", "import linkfinder").Run(); err != nil {
		return fmt.Errorf("LinkFinder not found in virtual environment, run with -id to install")
	}
	if err := exec.Command(python, "-c", "import secretfinder").Run(); err != nil {
		return fmt.Errorf("SecretFinder not found in virtual environment, run with -id to install")
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
	if _, err := exec.LookPath("npm"); err != nil {
		fmt.Println("[*] npm not found. Please install Node.js and npm (e.g., sudo apt install nodejs npm)")
		return err
	}

	fmt.Println("[*] Installing Go dependencies...")
	if err := exec.Command("go", "mod", "tidy").Run(); err != nil {
		return fmt.Errorf("failed to tidy Go modules: %v", err)
	}

	toolsPath := filepath.Join("toolssparky")
	if err := os.MkdirAll(toolsPath, 0755); err != nil {
		return fmt.Errorf("failed to create tools directory: %v", err)
	}

	// Install Go-based tools
	goTools := []string{"subfinder", "assetfinder", "amass", "httpx", "dnsx", "katana", "waybackurls", "ffuf", "hakrawler", "anew", "gf", "nuclei"}
	for _, tool := range goTools {
		fmt.Printf("[*] Checking %s...\n", tool)
		if _, err := exec.LookPath(tool); err == nil {
			if err := exec.Command(tool, "--version").Run(); err == nil {
				fmt.Printf("[*] %s is already installed and working\n", tool)
				continue
			}
		}
		fmt.Printf("[*] Installing %s...\n", tool)
		if err := exec.Command("go", "install", fmt.Sprintf("github.com/projectdiscovery/%s/cmd/%s@latest", tool, tool)).Run(); err != nil {
			return fmt.Errorf("failed to install %s: %v", tool, err)
		}
	}

	// Install jsbeautifier with npm
	fmt.Println("[*] Checking jsbeautifier...")
	if _, err := exec.LookPath("jsbeautifier"); err == nil {
		if err := exec.Command("jsbeautifier", "-v").Run(); err == nil {
			fmt.Println("[*] jsbeautifier is already installed and working")
		} else {
			fmt.Println("[*] Installing jsbeautifier...")
			if err := exec.Command("npm", "install", "-g", "js-beautify").Run(); err != nil {
				return fmt.Errorf("failed to install jsbeautifier: %v", err)
			}
		}
	} else {
		fmt.Println("[*] Installing jsbeautifier...")
		if err := exec.Command("npm", "install", "-g", "js-beautify").Run(); err != nil {
			return fmt.Errorf("failed to install jsbeautifier: %v", err)
		}
	}

	// Install sqlmap using apt
	fmt.Println("[*] Checking sqlmap...")
	if _, err := exec.LookPath("sqlmap"); err == nil {
		if err := exec.Command("sqlmap", "--version").Run(); err == nil {
			fmt.Println("[*] sqlmap is already installed and working")
		} else {
			fmt.Println("[*] Installing sqlmap...")
			if err := exec.Command("sudo", "apt", "install", "-y", "sqlmap").Run(); err != nil {
				return fmt.Errorf("failed to install sqlmap (try running with sudo or manually): %v", err)
			}
		}
	} else {
		fmt.Println("[*] Installing sqlmap...")
		if err := exec.Command("sudo", "apt", "install", "-y", "sqlmap").Run(); err != nil {
			return fmt.Errorf("failed to install sqlmap (try running with sudo or manually): %v", err)
		}
	}

	// Install nslookup and whois (system tools)
	fmt.Println("[*] Checking nslookup and whois...")
	if _, err := exec.LookPath("nslookup"); err == nil {
		if _, err := exec.LookPath("whois"); err == nil {
			fmt.Println("[*] nslookup and whois are already installed and working")
		} else {
			fmt.Println("[*] Installing nslookup and whois...")
			if err := exec.Command("sudo", "apt", "install", "-y", "dnsutils", "whois").Run(); err != nil {
				return fmt.Errorf("failed to install dnsutils and whois (try running with sudo or manually): %v", err)
			}
		}
	} else {
		fmt.Println("[*] Installing nslookup and whois...")
		if err := exec.Command("sudo", "apt", "install", "-y", "dnsutils", "whois").Run(); err != nil {
			return fmt.Errorf("failed to install dnsutils and whois (try running with sudo or manually): %v", err)
		}
	}

	// Create a virtual environment for Python tools
	venvPath := filepath.Join(toolsPath, "venv")
	fmt.Println("[*] Checking virtual environment for Python tools...")
	if _, err := os.Stat(filepath.Join(venvPath, "bin", "python3")); err == nil {
		fmt.Println("[*] Virtual environment already exists")
	} else {
		fmt.Println("[*] Creating virtual environment for Python tools...")
		if err := exec.Command("python3", "-m", "venv", venvPath).Run(); err != nil {
			return fmt.Errorf("failed to create virtual environment: %v", err)
		}
	}

	// Install Python-based tools in the virtual environment
	pip := filepath.Join(venvPath, "bin", "pip3")
	python := filepath.Join(venvPath, "bin", "python3")
	pythonTools := []struct {
		name string
		url  string
	}{
		{"linkfinder", "https://github.com/GerbenJavado/LinkFinder.git"},
		{"secretfinder", "https://github.com/m4ll0k/SecretFinder.git"},
	}
	for _, tool := range pythonTools {
		toolPath := filepath.Join(toolsPath, tool.name)
		fmt.Printf("[*] Checking %s...\n", tool.name)
		if err := exec.Command(python, "-c", fmt.Sprintf("import %s", tool.name)).Run(); err == nil {
			fmt.Printf("[*] %s is already installed and working in virtual environment\n", tool.name)
			continue
		}

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
	}

	fmt.Println("[*] Dependencies installed successfully")
	return nil
}
