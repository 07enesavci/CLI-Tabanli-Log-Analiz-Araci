package analyzer

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"log-analyzer/backend/internal/parser"
	"log-analyzer/backend/internal/rules"
)

type LogEntry struct {
	Timestamp    string
	Source       string
	LogFile      string
	Line         string
	Summary      string
	MatchedRules []string
	Severity     string
}

type Analyzer struct {
	ruleManager *rules.Manager
}

func NewAnalyzer(ruleManager *rules.Manager) *Analyzer {
	return &Analyzer{
		ruleManager: ruleManager,
	}
}

func (a *Analyzer) AnalyzeFile(filePath string) ([]LogEntry, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	var entries []LogEntry
	scanner := bufio.NewScanner(file)
	
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		
		matchedRules := a.ruleManager.MatchRules(line)
		if len(matchedRules) > 0 {
			var ruleNames []string
			maxSeverity := "low"
			
			for _, rule := range matchedRules {
				ruleNames = append(ruleNames, rule.Name)
				if severityLevel(rule.Severity) > severityLevel(maxSeverity) {
					maxSeverity = rule.Severity
				}
			}
			
			timestamp := extractTimestamp(line)
			summary := parser.ParseLogLineToSummary(line)
			if summary == "" {
				summary = line
			}
			entries = append(entries, LogEntry{
				Timestamp:    timestamp,
				Source:       filepath.Base(filePath),
				LogFile:      filePath,
				Line:         line,
				Summary:      summary,
				MatchedRules: ruleNames,
				Severity:     maxSeverity,
			})
		}
	}
	
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	
	return entries, nil
}

func (a *Analyzer) AnalyzeMultipleFiles(filePaths []string) ([]LogEntry, error) {
	var allEntries []LogEntry
	
	for _, filePath := range filePaths {
		entries, err := a.AnalyzeFile(filePath)
		if err != nil {
			continue
		}
		allEntries = append(allEntries, entries...)
	}
	
	return allEntries, nil
}

func (a *Analyzer) ExportToCSV(entries []LogEntry, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create CSV file: %w", err)
	}
	defer file.Close()
	
	writer := csv.NewWriter(file)
	defer writer.Flush()
	header := []string{"Timestamp", "Source", "LogFile", "Severity", "MatchedRules", "Summary", "LogLine"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}
	for _, entry := range entries {
		summary := entry.Summary
		if summary == "" {
			summary = entry.Line
		}
		record := []string{
			entry.Timestamp,
			entry.Source,
			entry.LogFile,
			entry.Severity,
			strings.Join(entry.MatchedRules, "; "),
			summary,
			entry.Line,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}
	
	return nil
}

func extractTimestamp(line string) string {
	parts := strings.Fields(line)
	if len(parts) >= 3 {
		if len(parts[0]) <= 3 && len(parts[1]) <= 2 {
			return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2])
		}
		if strings.Contains(parts[0], "T") || strings.Contains(parts[0], "-") {
			return parts[0]
		}
	}
	return time.Now().Format(time.RFC3339)
}

func severityLevel(severity string) int {
	s := strings.ToLower(severity)
	switch s {
	case "critical", "kritik":
		return 4
	case "high", "yüksek":
		return 3
	case "medium", "orta":
		return 2
	case "low", "düşük":
		return 1
	default:
		return 0
	}
}
