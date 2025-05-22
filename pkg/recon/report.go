```go
package recon

import (
	"fmt"
	"os"
	"path/filepath"
)

func GenerateReport(domain, outputDir string, newSubdomains []string) {
	fmt.Printf("[*] Results for %s:\n", domain)
	fmt.Printf("  [*] Active Subdomains: %s/active_subdomains.txt\n", outputDir)
	if len(newSubdomains) > 0 {
		fmt.Printf("  [*] New Vhost Subdomains: %v\n", newSubdomains)
	}
	fmt.Printf("  [*] Non-CDN IPs: %s/ip_list.txt\n", outputDir)
	fmt.Printf("  [*] Smart Fuzzing Results: %s/smart_fuzzing_200.txt\n", outputDir)
	fmt.Printf("  [*] JS URLs: %s/js_urls.txt\n", outputDir)
	fmt.Printf("  [*] SSRF Injection Points: %s/gf/gf_ssrf.txt\n", outputDir)
	fmt.Printf("  [*] SQLi Injection Points: %s/gf/gf_sqli.txt\n", outputDir)
	fmt.Printf("  [*] JS Analysis: %s/js_analysis.txt\n", outputDir)
	fmt.Printf("  [*] Output Directory: %s/\n", outputDir)

	jsonReport := fmt.Sprintf(`{
	"domain": "%s",
	"new_vhost_subdomains": %q
}`, domain, newSubdomains)
	os.WriteFile(filepath.Join(outputDir, "report.json"), []byte(jsonReport), 0644)
}
```
