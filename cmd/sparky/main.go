package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"github.com/lediusa/sparky/pkg/config"
	"github.com/lediusa/sparky/pkg/deps"
	"github.com/lediusa/sparky/pkg/recon"
	"strings"
)

const logo = `
    ____  ____  ____  ____  
   /    \/    \/    \/    \ 
  /__      ___/____    ___/ 
 /    \  /    /    \  /    \
/     / /    /     / /     /
\    /_/    /\    /_/     /
 \_________/  \_________/
 ==========================
       S P A R K Y

 Sparky - Reconnaissance for Bug Hunters
`

const modulePath = "github.com/lediusa/sparky/cmd/sparky"

func main() {
	fmt.Println(logo)

	domain := flag.String("d", "", "Single domain to process (e.g., example.com)")
	file := flag.String("f", "", "File containing list of domains")
	installDeps := flag.Bool("id", false, "Install dependencies")
	update := flag.Bool("update", false, "Check and update to the latest version")
	vhost := flag.Bool("vhost", false, "Enable virtual host discovery")
	smartFuzz := flag.Bool("sm", false, "Enable smart fuzzing on 403/404 subdomains")
	sqli := flag.Bool("sqli", false, "Enable SQLi scanning with sqlmap")
	nuclei := flag.Bool("nuclei", false, "Enable nuclei scanning")
	jsFuzz := flag.Bool("jsf", false, "Enable JavaScript fuzzing (also -jsfuzzing)")
	jsFuzzLong := flag.Bool("jsfuzzing", false, "Enable JavaScript fuzzing (same as -jsf)")
	wcd := flag.Bool("wcd", false, "Enable Web Cache Deception testing")
	threads := flag.Int("threads", 1, "Number of concurrent threads")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("Usage: sparky [options]")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *update {
		currentVersion := "v1.0.0"
		cmd := exec.Command("git", "ls-remote", "--tags", "https://github.com/lediusa/sparky.git")
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error checking updates: %v\n", err)
			os.Exit(1)
		}
		tags := strings.Split(string(out), "\n")
		latestTag := ""
		for _, tag := range tags {
			if strings.HasPrefix(tag, "refs/tags/") {
				version := strings.TrimPrefix(tag, "refs/tags/")
				if version > latestTag && version > currentVersion {
					latestTag = version
				}
			}
		}
		if latestTag != "" && latestTag > currentVersion {
			fmt.Printf("New version available: %s. Run 'go install %s@%s' to update.\n", latestTag, modulePath, latestTag)
		} else {
			fmt.Println("You are on the latest version.")
		}
		os.Exit(0)
	}

	if *installDeps {
		if err := deps.InstallDependencies(); err != nil {
			fmt.Printf("Error installing dependencies: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("[*] Dependencies installed successfully")
		os.Exit(0)
	}

	if err := deps.CheckDependencies(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[*] Loading configuration from config.yaml")

	var domains []string
	if *domain != "" {
		domains = append(domains, *domain)
		fmt.Printf("[*] Processing single domain: %s\n", *domain)
	} else if *file != "" {
		domains, err = recon.ReadDomainsFromFile(*file)
		if err != nil {
			fmt.Printf("Error reading domains file: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[*] Loading domains from %s\n", *file)
		fmt.Printf("[*] Domains to process: %v\n", domains)
	} else {
		fmt.Println("Error: Either -d or -f must be specified")
		flag.PrintDefaults()
		os.Exit(1)
	}

	opts := recon.Options{
		Vhost:      *vhost,
		SmartFuzz:  *smartFuzz,
		SQLi:       *sqli,
		Nuclei:     *nuclei,
		JSFuzz:     *jsFuzz || *jsFuzzLong,
		WCD:        *wcd,
		Threads:    *threads,
		Config:     cfg,
		OutputBase: os.Getenv("PWD"),
	}

	recon.RunRecon(domains, opts)
}
