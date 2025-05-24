package deps

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func CheckDependencies() error {
	// Define toolsPath
	toolsPath := filepath.Join("toolssparky")

	// List of system and Go tools to check
	tools := []struct {
		name    string
		checkCmd []string
	}{
		{"subfinder", []string{"subfinder", "-h"}},
		{"assetfinder", []string{"assetfinder", "-h"}},
		{"amass", []string{"amass", "-h"}},
		{"httpx", []string{"httpx", "-h"}},
		{"dnsx", []string{"dnsx", "-h"}},
		{"jsbeautifier", []string{"jsbeautifier", "-v"}},
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
	pythonTools := []string{"linkfinder", "secretfinder"}
	for _, tool := range pythonTools {
		// Create a temporary Python script to test the module
		tempScript := fmt.Sprintf(`
import sys
try:
    import %s
    print("Module %s is installed")
    sys.exit(0)
except ImportError:
    print("Module %s is not installed")
    sys.exit(1)
`, tool, tool, tool)

		tempFile := filepath.Join(toolsPath, fmt.Sprintf("check_%s.py", tool))
		if err := os.WriteFile(tempFile, []byte(tempScript), 0644); err != nil {
			return fmt.Errorf("failed to create temp script for %s: %v", tool, err)
		}
		defer os.Remove(tempFile)

		cmd := exec.Command(python, tempFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("tool %s not found in virtual environment, run with -id to install: %s", tool, string(output))
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
	if _, err := exec.LookPath("npm"); err != nil {
		fmt.Println("[*] npm not found. Please install Node.js and npm (e.g., sudo apt install nodejs npm)")
		return err
	}

	// Ensure GOPATH is set
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		gopath = filepath.Join(os.Getenv("HOME"), "go")
		os.Setenv("GOPATH", gopath)
		fmt.Printf("[*] GOPATH set to %s\n", gopath)
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
		fmt.Printf("[*] Checking %s...\n", tool.name)
		if _, err := exec.LookPath(tool.name); err == nil {
			// Try a simple command to verify it works
			var cmd *exec.Cmd
			if tool.name == "assetfinder" || tool.name == "anew" || tool.name == "waybackurls" {
				cmd = exec.Command(tool.name, "-h")
			} else {
				cmd = exec.Command(tool.name, "--version")
			}
			if err := cmd.Run(); err == nil {
				fmt.Printf("[*] %s is already installed and working\n", tool.name)
				continue
			}
		}
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
			if err := exec.Command("nslookup", "-version").Run(); err == nil {
				if err := exec.Command("whois", "--version").Run(); err == nil {
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
		// Check if the module is working
		tempScript := fmt.Sprintf(`
import sys
try:
    import %s
    print("Module %s is installed")
    sys.exit(0)
except ImportError:
    print("Module %s is not installed")
    sys.exit(1)
`, tool.name, tool.name, tool.name)

		tempFile := filepath.Join(toolsPath, fmt.Sprintf("check_%s.py", tool))
		if err := os.WriteFile(tempFile, []byte(tempScript), 0644); err != nil {
			return fmt.Errorf("failed to create temp script for %s: %v", tool, err)
		}
		defer os.Remove(tempFile)

		cmd := exec.Command(python, tempFile)
		if err := cmd.Run(); err == nil {
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
