package recon

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lediusa/sparky/pkg/config"
)

func SubdomainDiscovery(domain, outputFile string, cfg *config.Config) error {
	// Define the tools and their arguments
	tools := []struct {
		name     string
		args     []string
		tempFile string
	}{
		{
			name:     "subfinder",
			args:     []string{"-d", domain, "-all", "-silent"},
			tempFile: outputFile + ".subfinder",
		},
		{
			name:     "assetfinder",
			args:     []string{"--subs-only", domain},
			tempFile: outputFile + ".assetfinder",
		},
		{
			name:     "amass",
			args:     []string{"enum", "-d", domain, "-o", outputFile + ".amass"},
			tempFile: outputFile + ".amass",
		},
	}

	// Run each tool and save its output to a temporary file
	for _, tool := range tools {
		cmd := exec.Command(tool.name, tool.args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("[!] Warning: %s failed with error: %v\nOutput: %s\n", tool.name, err, string(output))
			// Continue to the next tool even if one fails
			continue
		}

		// For tools other than amass, we need to write the output manually
		if tool.name != "amass" {
			if err := os.WriteFile(tool.tempFile, output, 0644); err != nil {
				fmt.Printf("[!] Warning: Failed to write output for %s to %s: %v\n", tool.name, tool.tempFile, err)
				continue
			}
		}
	}

	// Combine the results into a single file with unique subdomains
	combinedOutput, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create final output file %s: %v", outputFile, err)
	}
	defer combinedOutput.Close()

	// Use sort -u to combine and deduplicate
	sortCmd := exec.Command("sh", "-c", fmt.Sprintf("cat %s.subfinder %s.assetfinder %s.amass 2>/dev/null | sort -u", outputFile, outputFile, outputFile))
	sortCmd.Stdout = combinedOutput
	if err := sortCmd.Run(); err != nil {
		return fmt.Errorf("failed to combine and deduplicate subdomains: %v", err)
	}

	// Clean up temporary files explicitly
	for _, tool := range tools {
		if err := os.Remove(tool.tempFile); err != nil {
			fmt.Printf("[!] Warning: Failed to remove temporary file %s: %v\n", tool.tempFile, err)
		}
	}

	// Check if the final output file is empty
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return fmt.Errorf("failed to check final output file %s: %v", outputFile, err)
	}
	if fileInfo.Size() == 0 {
		fmt.Println("[*] Subdomain discovery completed: No subdomains found.")
		return fmt.Errorf("no subdomains found for %s", domain)
	}

	// Count the number of subdomains in the final file
	file, err := os.Open(outputFile)
	if err != nil {
		return fmt.Errorf("failed to open final output file %s for counting: %v", outputFile, err)
	}
	defer file.Close()

	subdomainCount := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if scanner.Text() != "" { // Ignore empty lines
			subdomainCount++
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading final output file %s: %v", outputFile, err)
	}

	// Print completion message with the number of subdomains found
	fmt.Printf("[*] Subdomain discovery completed: %d unique subdomains found for %s.\n", subdomainCount, domain)

	return nil
}
