package recon

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "github.com/lediusa/sparky/pkg/config"
)

// Options defines the configuration for reconnaissance tasks.
type Options struct {
    Vhost      bool
    SmartFuzz  bool
    SQLi       bool
    Nuclei     bool
    JSFuzz     bool
    WCD        bool
    Threads    int
    Config     *config.Config
    OutputBase string
}

// RunRecon processes a list of domains concurrently.
func RunRecon(domains []string, opts Options) {
    var wg sync.WaitGroup
    sem := make(chan struct{}, opts.Threads)

    for _, domain := range domains {
        wg.Add(1)
        sem <- struct{}{}
        go func(d string) {
            defer wg.Done()
            defer func() { <-sem }()
            reconDomain(d, opts)
        }(domain)
    }
    wg.Wait()
    fmt.Println("[*] Finished processing all domains.")
}

// reconDomain performs reconnaissance tasks for a single domain.
func reconDomain(domain string, opts Options) {
    outputDir := filepath.Join(opts.OutputBase, fmt.Sprintf("recon_%s", domain))
    if err := os.MkdirAll(outputDir, 0755); err != nil {
        fmt.Printf("Error creating output dir for %s: %v\n", domain, err)
        return
    }
    fmt.Printf("[*] Starting recon for %s\n", domain)

    // Subdomain discovery
    subdomainsFile := filepath.Join(outputDir, "subdomains.txt")
    if err := SubdomainDiscovery(domain, subdomainsFile, opts.Config); err != nil {
        fmt.Printf("Error in subdomain discovery for %s: %v\n", domain, err)
        return
    }
    fmt.Printf("[*] Subdomain discovery for %s...\n", domain)

    // Filter active subdomains
    activeSubdomainsFile := filepath.Join(outputDir, "active_subdomains.txt")
    if err := FilterActiveSubdomains(subdomainsFile, activeSubdomainsFile); err != nil {
        fmt.Printf("Error filtering active subdomains: %v\n", err)
        return
    }
    fmt.Printf("[*] Filtering active subdomains with dnsx...\n")

    // Resolve IPs
    ipsFile := filepath.Join(outputDir, "ips.txt")
    if err := ResolveIPs(activeSubdomainsFile, ipsFile); err != nil {
        fmt.Printf("Error resolving IPs: %v\n", err)
        return
    }
    fmt.Printf("[*] Resolving IPs for subdomains using nslookup...\n")

    // Identify non-CDN IPs
    nonCDNIPs, err := IdentifyNonCDNIPs(ipsFile, outputDir)
    if err != nil {
        fmt.Printf("Error identifying non-CDN IPs: %v\n", err)
        return
    }
    fmt.Printf("[*] Found %d non-CDN IPs\n", len(nonCDNIPs))

    // Virtual host discovery
    newSubdomains := []string{}
    if opts.Vhost && len(nonCDNIPs) > 0 {
        vhostFile := filepath.Join(outputDir, "vhost.txt")
        newSubdomainsFile := filepath.Join(outputDir, "new_vhost_subdomains.txt")
        if err := VhostDiscovery(domain, nonCDNIPs, vhostFile, opts.Config); err != nil {
            fmt.Printf("Error in vhost discovery: %v\n", err)
            return
        }
        newSubdomains, err = CompareSubdomains(activeSubdomainsFile, vhostFile, newSubdomainsFile)
        if err != nil {
            fmt.Printf("Error comparing subdomains: %v\n", err)
            return
        }
        fmt.Printf("[*] Virtual host discovery for %s...\n", domain)
        if len(newSubdomains) > 0 {
            fmt.Printf("[*] New subdomains found: %v\n", newSubdomains)
        }
    }

    // Check for 403/404 subdomains
    forbiddenFile := filepath.Join(outputDir, "forbidden_subdomains.txt")
    if err := CheckForbiddenSubdomains(activeSubdomainsFile, forbiddenFile); err != nil {
        fmt.Printf("Error checking 403/404 subdomains: %v\n", err)
        return
    }
    fmt.Printf("[*] Checking subdomains for 403/404 status...\n")

    // Smart fuzzing on 403/404 subdomains
    if opts.SmartFuzz {
        fuzzFile := filepath.Join(outputDir, "smart_fuzzing.txt")
        fuzz200File := filepath.Join(outputDir, "smart_fuzzing_200.txt")
        fuzzResults, err := SmartFuzzing(forbiddenFile, fuzzFile, fuzz200File, opts.Config)
        if err != nil {
            fmt.Printf("Error in smart fuzzing: %v\n", err)
            return
        }
        if len(fuzzResults) > 0 {
            fmt.Printf("[*] Smart fuzzing on 403/404 subdomains...\n")
            fmt.Printf("[*] Found accessible paths: %v\n", fuzzResults)
        }
    }

    // Crawl subdomains
    urlsFile := filepath.Join(outputDir, "urls.txt")
    if err := CrawlSubdomains(activeSubdomainsFile, urlsFile); err != nil {
        fmt.Printf("Error crawling subdomains: %v\n", err)
        return
    }
    fmt.Printf("[*] Crawling subdomains for %s...\n", domain)

    // Extract JS URLs
    jsUrlsFile := filepath.Join(outputDir, "js_urls.txt")
    if err := ExtractJSUrls(urlsFile, jsUrlsFile); err != nil {
        fmt.Printf("Error extracting JS URLs: %v\n", err)
        return
    }
    fmt.Printf("[*] Extracting JavaScript URLs...\n")

    // JS fuzzing
    var newJSFiles []string
    if opts.JSFuzz {
        newJSFile := filepath.Join(outputDir, "new_js_files.txt")
        newJSFiles, err = JSFuzzing(urlsFile, newJSFile, opts.Config)
        if err != nil {
            fmt.Printf("Error in JS fuzzing: %v\n", err)
            return
        }
        if len(newJSFiles) > 0 {
            fmt.Printf("[*] JS fuzzing completed. New JS files found: %v\n", newJSFiles)
        } else {
            fmt.Println("[*] JS fuzzing completed. No new JS files found.")
        }
    }

    // Analyze JS files
    jsAnalysisFile := filepath.Join(outputDir, "js_analysis.txt")
    if err := AnalyzeJSFiles(jsUrlsFile, jsAnalysisFile, opts.Config); err != nil {
        fmt.Printf("Error analyzing JS files: %v\n", err)
        return
    }
    fmt.Printf("[*] JS Analysis for %s...\n", domain)

    // Analyze URLs with gf
    gfDir := filepath.Join(outputDir, "gf")
    if err := os.MkdirAll(gfDir, 0755); err != nil {
        fmt.Printf("Error creating gf dir: %v\n", err)
        return
    }
    gfSSRF := filepath.Join(gfDir, "gf_ssrf.txt")
    gfSQLi := filepath.Join(gfDir, "gf_sqli.txt")
    if err := AnalyzeWithGF(urlsFile, gfSSRF, gfSQLi); err != nil {
        fmt.Printf("Error analyzing with gf: %v\n", err)
        return
    }
    fmt.Printf("[*] Analyzing URLs for SSRF and SQLi...\n")

    // SQLi scan
    if opts.SQLi {
        sqliResultsFile := filepath.Join(outputDir, "sqli_results.txt")
        if err := ScanSQLi(gfSQLi, sqliResultsFile); err != nil {
            fmt.Printf("Error in SQLi scan: %v\n", err)
            return
        }
        fmt.Printf("[*] SQLi Scan Results for %s...\n", domain)
    }

    // Nuclei scan
    if opts.Nuclei {
        nucleiResultsFile := filepath.Join(outputDir, "nuclei_results.txt")
        if err := ScanNuclei(activeSubdomainsFile, nucleiResultsFile, opts.Config); err != nil {
            fmt.Printf("Error in nuclei scan: %v\n", err)
            return
        }
        fmt.Printf("[*] Nuclei Scan Results for %s...\n", domain)
    }

    // Web Cache Deception testing
    if opts.WCD {
        wcdDir := filepath.Join(outputDir, "wcd")
        if err := os.MkdirAll(wcdDir, 0755); err != nil {
            fmt.Printf("Error creating wcd dir: %v\n", err)
            return
        }
        wcdResultsFile := filepath.Join(wcdDir, "wcd_results.json")
        cachedFiles, err := TestWebCacheDeception(urlsFile, wcdResultsFile)
        if err != nil {
            fmt.Printf("Error in Web Cache Deception testing: %v\n", err)
            return
        }
        if len(cachedFiles) > 0 {
            fmt.Printf("[*] Web Cache Deception test completed. Found %d potentially cached files. Check %s for details.\n", len(cachedFiles), wcdResultsFile)
        } else {
            fmt.Printf("[*] Web Cache Deception test completed. No cached files found.\n")
        }
    }

    // Generate report
    GenerateReport(domain, outputDir, newSubdomains)
    fmt.Printf("[*] Finished processing %s\n", domain)
}

// FilterActiveSubdomains simulates filtering active subdomains.
func FilterActiveSubdomains(inputFile, outputFile string) error {
    return nil
}

// ResolveIPs simulates resolving IPs for subdomains.
func ResolveIPs(inputFile, outputFile string) error {
    return nil
}

// IdentifyNonCDNIPs simulates identifying non-CDN IPs.
func IdentifyNonCDNIPs(ipsFile, outputDir string) ([]string, error) {
    return []string{}, nil
}

// CheckForbiddenSubdomains simulates checking for 403/404 subdomains.
func CheckForbiddenSubdomains(inputFile, outputFile string) error {
    return nil
}
