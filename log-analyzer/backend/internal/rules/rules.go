package rules

import (
	"fmt"
	"os"
	"regexp"
	"sync"

	"gopkg.in/yaml.v3"
)

type Rule struct {
	Name          string `yaml:"name" json:"name"`
	Pattern       string `yaml:"pattern" json:"pattern"`
	ExcludePattern string `yaml:"exclude_pattern" json:"exclude_pattern,omitempty"`
	Severity      string `yaml:"severity" json:"severity"`
	Description   string `yaml:"description" json:"description"`
	Enabled       bool   `yaml:"enabled" json:"enabled"`
	regex         *regexp.Regexp
	excludeRegex  *regexp.Regexp
}

type LogFile struct {
	Path    string `yaml:"path" json:"path"`
	Type    string `yaml:"type" json:"type"`
	Enabled bool   `yaml:"enabled" json:"enabled"`
}

type Config struct {
	Rules    []Rule    `yaml:"rules" json:"rules"`
	LogFiles []LogFile `yaml:"log_files" json:"log_files"`
}

type Manager struct {
	config     *Config
	configPath string
	mu         sync.RWMutex
}

func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
	}
	
	if err := m.LoadConfig(); err != nil {
		return nil, err
	}
	
	return m, nil
}

func (m *Manager) LoadConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config: %w", err)
	}
	for i := range config.Rules {
		if config.Rules[i].Enabled {
			regex, err := regexp.Compile(config.Rules[i].Pattern)
			if err != nil {
				return fmt.Errorf("invalid regex pattern for rule %s: %w", config.Rules[i].Name, err)
			}
			config.Rules[i].regex = regex
			if config.Rules[i].ExcludePattern != "" {
				excludeRegex, err := regexp.Compile(config.Rules[i].ExcludePattern)
				if err != nil {
					return fmt.Errorf("invalid exclude pattern for rule %s: %w", config.Rules[i].Name, err)
				}
				config.Rules[i].excludeRegex = excludeRegex
			}
		}
	}
	
	m.config = &config
	return nil
}

func (m *Manager) GetRules() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	rules := make([]Rule, len(m.config.Rules))
	copy(rules, m.config.Rules)
	return rules
}

func (m *Manager) GetEnabledRules() []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var enabled []Rule
	for _, rule := range m.config.Rules {
		if rule.Enabled {
			enabled = append(enabled, rule)
		}
	}
	return enabled
}

func (m *Manager) GetLogFiles() []LogFile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	files := make([]LogFile, len(m.config.LogFiles))
	copy(files, m.config.LogFiles)
	return files
}

func (m *Manager) GetEnabledLogFiles() []LogFile {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var enabled []LogFile
	for _, file := range m.config.LogFiles {
		if file.Enabled {
			enabled = append(enabled, file)
		}
	}
	return enabled
}

func (m *Manager) MatchRules(line string) []Rule {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var matches []Rule
	for _, rule := range m.config.Rules {
		if rule.Enabled && rule.regex != nil && rule.regex.MatchString(line) {
			if rule.excludeRegex != nil && rule.excludeRegex.MatchString(line) {
				continue
			}
			matches = append(matches, rule)
		}
	}
	return matches
}

