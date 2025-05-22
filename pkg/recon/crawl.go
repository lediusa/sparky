package recon

import (
	"os"
	"os/exec"
	"strings"
)

func CrawlSubdomains(inputFile, outputFile string) error {
	tools := []struct {
		name string
		args []string
	}{
		{"waybackurls", []string{}},
		{"katana", []string{"-list", inputFile, "-jc"}},
		{"hakrawler", []string{"-d", "2", "-t", "8"}},
	}
	for _, tool := range tools {
		args := append(tool.args, "<", inputFile)
		cmd := exec.Command("sh", "-c", tool.name+" "+strings.Join(args, " ")+" > "+outputFile+"."+tool.name)
		if err := cmd.Run(); err != nil {
			continue
		}
	}

	cmd := exec.Command("sh", "-c", "cat "+outputFile+".* | sort -u | grep -vE '\\.(css|png|jpg|jpeg|gif|svg|woff|woff2|ttf|ico)$' > "+outputFile)
	if err := cmd.Run(); err != nil {
		return err
	}

	os.RemoveAll(outputFile + ".*")
	return nil
}
