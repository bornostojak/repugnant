package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const ConfigFileName = "rpg.conf.yaml"

// Config controls generation and publishing for one documented repository.
type Config struct {
	Version int           `yaml:"version"`
	Langs   []string      `yaml:"langs,omitempty"`
	Output  OutputConfig  `yaml:"output"`
	Hooks   HookConfig    `yaml:"hooks"`
	Project ProjectConfig `yaml:"project"`
}

// EnabledLanguages honors an explicit config list. Otherwise it infers likely
// languages from conventional project-root manifests; a repository with no
// recognizable manifest remains permissive so small scripts still work.
func (c Config) EnabledLanguages(root string) map[string]bool {
	if len(c.Langs) > 0 {
		return languageSet(c.Langs)
	}
	hints := map[string][]string{
		"go": {"go.mod", "go.sum"}, "typescript": {"tsconfig.json"}, "javascript": {"package.json"},
		"rust": {"Cargo.toml"}, "python": {"pyproject.toml", "requirements.txt"}, "ruby": {"Gemfile"},
		"dart": {"pubspec.yaml"}, "cpp": {"CMakeLists.txt"},
	}
	result := map[string]bool{}
	for language, files := range hints {
		for _, file := range files {
			if _, err := os.Stat(filepath.Join(root, file)); err == nil {
				result[language] = true
				break
			}
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
func languageSet(langs []string) map[string]bool {
	out := map[string]bool{}
	for _, language := range langs {
		out[strings.ToLower(strings.TrimSpace(language))] = true
	}
	return out
}

type OutputConfig struct {
	Docs DocsOutputConfig `yaml:"docs"`
	Web  WebOutputConfig  `yaml:"web"`
}

type DocsOutputConfig struct {
	Enabled bool   `yaml:"enabled"`
	Dir     string `yaml:"dir"`
}

type WebOutputConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint,omitempty"`
}

type HookConfig struct {
	OnPublishFailure string `yaml:"on_publish_failure"`
}
type ProjectConfig struct {
	Slug   string `yaml:"slug,omitempty"`
	APIURL string `yaml:"api_url,omitempty"`
	APIKey string `yaml:"api_key,omitempty"`
}

func DefaultConfig() Config {
	return Config{
		Version: 1,
		Output: OutputConfig{
			Docs: DocsOutputConfig{Enabled: true, Dir: "docs"},
			Web:  WebOutputConfig{Enabled: false},
		},
		Hooks: HookConfig{OnPublishFailure: "block"},
	}
}

func Load(root string) (Config, error) {
	path := filepath.Join(root, ConfigFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("read %s: %w", ConfigFileName, err)
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", ConfigFileName, err)
	}
	if err := config.Validate(); err != nil {
		return Config{}, err
	}
	return config, nil
}

func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("rpg.conf.yaml: unsupported version %d", c.Version)
	}
	if c.Output.Docs.Enabled && c.Output.Docs.Dir == "" {
		return fmt.Errorf("rpg.conf.yaml: output.docs.dir is required when local docs are enabled")
	}
	if c.Output.Web.Enabled && c.Output.Web.Endpoint == "" {
		return fmt.Errorf("rpg.conf.yaml: output.web.endpoint is required when web output is enabled")
	}
	if c.Hooks.OnPublishFailure != "block" && c.Hooks.OnPublishFailure != "allow_pending" {
		return fmt.Errorf("rpg.conf.yaml: hooks.on_publish_failure must be block or allow_pending")
	}
	if c.Output.Web.Enabled && (c.Project.Slug == "" || c.Project.APIURL == "" || c.Project.APIKey == "") {
		return fmt.Errorf("rpg.conf.yaml: project.slug, project.api_url, and project.api_key are required for web output")
	}
	return nil
}
