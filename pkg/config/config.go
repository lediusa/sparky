```go
package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Paths struct {
		Subdomains     string `yaml:"subdomains"`
		Resolvers      string `yaml:"resolvers"`
		BackupLogin    string `yaml:"backup_login"`
		JsSmartFuzzing string `yaml:"js-smart-fuzzing"`
		OutputDir      string `yaml:"output_dir"`
		NucleiTemplates string `yaml:"nuclei_templates"`
	} `yaml:"paths"`
	Settings struct {
		Threads int `yaml:"threads"`
		Timeout int `yaml:"timeout"`
	} `yaml:"settings"`
	Tools map[string]string `yaml:"tools"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config/config.yaml")
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Tools == nil {
		cfg.Tools = make(map[string]string)
	}
	cfg.Tools["linkfinder"] = "toolssparky/linkfinder/LinkFinder.py"
	cfg.Tools["secretfinder"] = "toolssparky/SecretFinder/SecretFinder.py"

	return &cfg, nil
}
```
