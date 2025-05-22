package recon

import (
	"os"
	"os/exec"
	"sparky/pkg/config"
)

func ScanNuclei(inputFile, outputFile string, cfg *config.Config) error {
	cmd := exec.Command("nuclei", "-l", inputFile, "-t", cfg.Paths.NucleiTemplates)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, out, 0644)
}
