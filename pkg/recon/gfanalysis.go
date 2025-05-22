```go
package recon

import (
	"os"
	"os/exec"
)

func AnalyzeWithGF(inputFile, ssrfFile, sqliFile string) error {
	cmd := exec.Command("gf", "ssrf")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		data, _ := os.ReadFile(inputFile)
		stdin.Write(data)
	}()
	out, err := cmd.Output()
	if err != nil {
		return err
	}
	os.WriteFile(ssrfFile, out, 0644)

	cmd = exec.Command("gf", "sqli")
	stdin, err = cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		data, _ := os.ReadFile(inputFile)
		stdin.Write(data)
	}()
	out, err = cmd.Output()
	if err != nil {
		return err
	}
	os.WriteFile(sqliFile, out, 0644)

	return nil
}
```
