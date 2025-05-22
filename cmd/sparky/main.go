```go
package main

import (
	"flag"
	"fmt"
	"os"
	"sparky/pkg/config"
	"sparky/pkg/deps"
	"sparky/pkg/recon"
)

const logo = `
\033[32m
   _____ _               
  / ____| |              
 | (___ | |__   __ _ _ __   ___ _   _ 
  \___ \| '_ \ / _` + "`" + ` | '_ \ / __| | | |
  ____) | | | | (_| | |_) | (__| |_| |
 |_____/|_| |_|__,_|_.__/ \___|\__,_|

\033[37mSparky - Reconnaissance for Bug Hunters\033[0m
`

func main() {
	fmt.Println(logo)

	domain := flag.String("d", "", "Single domain to process (e.g., example.com)")
	file := flag.String("f", "", "File containing list of domains")
	installDeps := flag.Bool("id", false, "Install dependencies")
	vhost := flag.Bool("vhost", false, "Enable virtual host discovery")
	smartFuzz := flag.Bool("sm", false, "Enable smart fuzzing on 403/404 subdomains")
	sqli := flag.Bool("sqli", false, "Enable SQLi scanning with sqlmap")
	nuclei := flag.Bool("nuclei", false, "Enable nuclei scanning")
	threads := flag.Int("threads", 1, "Number of concurrent threads")
	help := flag.Bool("h", false, "Show help")
	flag.Parse()

	if *help {
		fmt.Println("Usage: ./sparky [options]")
		flag.PrintDefaults()
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
		Threads:    *threads,
		Config:     cfg,
		OutputBase: os.Getenv("PWD"),
	}

	recon.RunRecon(domains, opts)
}
```
