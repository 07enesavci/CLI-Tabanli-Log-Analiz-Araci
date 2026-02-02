package tailer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"log-analyzer/backend/internal/rules"
)

type Alert struct {
	Timestamp   time.Time
	Source      string
	LogFile     string
	Line        string
	MatchedRules []string
	Severity    string
}

type Tailer struct {
	ruleManager *rules.Manager
	alerts      chan Alert
	stopChan    chan struct{}
	wg          sync.WaitGroup
	mu          sync.Mutex
	watchers    map[string]*fileWatcher
}

type fileWatcher struct {
	file    *os.File
	path    string
	stop    chan struct{}
	mu      sync.Mutex
	lastPos int64
}

func NewTailer(ruleManager *rules.Manager) *Tailer {
	return &Tailer{
		ruleManager: ruleManager,
		alerts:      make(chan Alert, 100),
		stopChan:    make(chan struct{}),
		watchers:    make(map[string]*fileWatcher),
	}
}

func (t *Tailer) StartWatching(filePath string) error {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if _, exists := t.watchers[filePath]; exists {
		return fmt.Errorf("already watching %s", filePath)
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			dir := filepath.Dir(filePath)
			if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
				return fmt.Errorf("failed to create dir %s: %w", dir, mkErr)
			}
			if createErr := os.WriteFile(filePath, nil, 0644); createErr != nil {
				return fmt.Errorf("failed to create file: %w", createErr)
			}
			file, err = os.Open(filePath)
		}
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
	}
	fileInfo, err := file.Stat()
	if err != nil {
		file.Close()
		return fmt.Errorf("failed to stat file: %w", err)
	}
	
	file.Seek(fileInfo.Size(), 0)
	
	watcher := &fileWatcher{
		file:    file,
		path:    filePath,
		stop:    make(chan struct{}),
		lastPos: fileInfo.Size(),
	}
	
	t.watchers[filePath] = watcher
	t.wg.Add(1)
	go t.watchFile(watcher)
	
	return nil
}

func (t *Tailer) watchFile(watcher *fileWatcher) {
	defer t.wg.Done()
	defer watcher.file.Close()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-watcher.stop:
			return
		case <-t.stopChan:
			return
		case <-ticker.C:
			t.readNewLines(watcher)
		}
	}
}

func (t *Tailer) readNewLines(watcher *fileWatcher) {
	watcher.mu.Lock()
	defer watcher.mu.Unlock()
	fileInfo, err := os.Stat(watcher.path)
	if err != nil {
		return
	}
	currentSize := fileInfo.Size()
	if currentSize > watcher.lastPos {
		toRead := currentSize - watcher.lastPos
		if _, err := watcher.file.Seek(watcher.lastPos, 0); err != nil {
			t.reopenWatcher(watcher, currentSize)
			return
		}
		buf := make([]byte, toRead)
		n, err := io.ReadFull(watcher.file, buf)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			t.reopenWatcher(watcher, currentSize)
			return
		}
		watcher.lastPos += int64(n)
		chunk := string(buf[:n])
		for _, rawLine := range strings.Split(chunk, "\n") {
			line := strings.TrimSpace(rawLine)
			if line == "" {
				continue
			}
			matchedRules := t.ruleManager.MatchRules(line)
			if len(matchedRules) > 0 {
				var ruleNames []string
				maxSeverity := "low"
				for _, rule := range matchedRules {
					ruleNames = append(ruleNames, rule.Name)
					ruleSevLevel := severityLevel(rule.Severity)
					maxSevLevel := severityLevel(maxSeverity)
					if ruleSevLevel > maxSevLevel {
						maxSeverity = rule.Severity
					}
				}
				alert := Alert{
					Timestamp:    time.Now(),
					Source:       watcher.path,
					LogFile:      watcher.path,
					Line:         line,
					MatchedRules: ruleNames,
					Severity:     maxSeverity,
				}
				select {
				case t.alerts <- alert:
				default:
				}
			}
		}
	} else if currentSize < watcher.lastPos {
		t.reopenWatcher(watcher, currentSize)
	}
}

func (t *Tailer) reopenWatcher(watcher *fileWatcher, newSize int64) {
	watcher.file.Close()
	if newFile, err := os.Open(watcher.path); err == nil {
		if fi, err := newFile.Stat(); err == nil {
			newSize = fi.Size()
		}
		watcher.file = newFile
		watcher.lastPos = newSize
		watcher.file.Seek(newSize, 0)
	}
}

func (t *Tailer) StopWatching(filePath string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	if watcher, exists := t.watchers[filePath]; exists {
		close(watcher.stop)
		delete(t.watchers, filePath)
	}
}

func (t *Tailer) Stop() {
	close(t.stopChan)
	
	t.mu.Lock()
	for _, watcher := range t.watchers {
		close(watcher.stop)
	}
	t.mu.Unlock()
	
	t.wg.Wait()
	close(t.alerts)
}

func (t *Tailer) Alerts() <-chan Alert {
	return t.alerts
}

func (t *Tailer) GetWatchedFiles() []string {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	files := make([]string, 0, len(t.watchers))
	for path := range t.watchers {
		files = append(files, path)
	}
	return files
}

func (t *Tailer) IsWatching(filePath string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	
	_, exists := t.watchers[filePath]
	return exists
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
