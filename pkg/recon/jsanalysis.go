package recon

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"github.com/lediusa/sparky/pkg/config"
)

func ExtractJSUrls(inputFile, outputFile string) error {
	cmd := exec.Command("grep", "-iE", "\\.js$", inputFile)
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	return os.WriteFile(outputFile, out, 0644)
}

func AnalyzeJSFiles(jsUrlsFile, outputFile string, cfg *config.Config) error {
	file, err := os.Open(jsUrlsFile)
	if err != nil {
		return err
	}
	defer file.Close()

	jsDir := filepath.Join(filepath.Dir(outputFile), "js_files")
	if err := os.MkdirAll(jsDir, 0755); err != nil {
		return err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		if url == "" {
			continue
		}
		filename := filepath.Join(jsDir, filepath.Base(url))
		cmd := exec.Command("curl", "-s", url)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		os.WriteFile(filename, out, 0644)

		beautified := filename + ".beautified"
		cmd = exec.Command("jsbeautifier", filename)
		out, err = cmd.Output()
		if err != nil {
			continue
		}
		os.WriteFile(beautified, out, 0644)

		for _, tool := range []string{"linkfinder", "secretfinder"} {
			cmd = exec.Command("python3", cfg.Tools[tool], "-i", beautified, "-o", "cli")
			cmd.Dir = filepath.Dir(cfg.Tools[tool])
			out, err = cmd.Output()
			if err != nil {
				continue
			}
			os.WriteFile(outputFile, out, 0644|os.ModeAppend)
		}
	}
	return nil
}
