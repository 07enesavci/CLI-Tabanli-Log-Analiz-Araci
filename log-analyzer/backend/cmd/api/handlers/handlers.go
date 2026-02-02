package handlers

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"log-analyzer/backend/internal/analyzer"
	"log-analyzer/backend/internal/parser"
	"log-analyzer/backend/internal/rules"
	"log-analyzer/backend/internal/tailer"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Handler struct {
	ruleManager   *rules.Manager
	analyzer      *analyzer.Analyzer
	tailer        *tailer.Tailer
	alerts        []AlertResponse
	mu            sync.RWMutex
	wsConnections map[*websocket.Conn]struct{}
	wsMu          sync.RWMutex
	upgrader      websocket.Upgrader
}

type AlertResponse struct {
	Timestamp    time.Time `json:"timestamp"`
	Source       string    `json:"source"`
	LogFile      string    `json:"logFile"`
	Line         string    `json:"line"`
	Summary      string    `json:"summary"`
	MatchedRules []string  `json:"matchedRules"`
	Severity     string    `json:"severity"`
}

type AnalyzeRequest struct {
	Files []string `json:"files"`
}

type TailRequest struct {
	Files []string `json:"files"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Error   string `json:"error,omitempty"`
}

const (
	authUsername = "admin"
	authPassword = "admin"
	authToken    = "log-analyzer-auth-token"
)

type StatsResponse struct {
	TotalAlerts      int            `json:"totalAlerts"`
	SeverityCount    map[string]int `json:"severityCount"`
	ActiveRules      int            `json:"activeRules"`
	WatchedFiles     int            `json:"watchedFiles"`
	IsTailing        bool           `json:"isTailing"`
	WatchedFilesList []string       `json:"watchedFilesList"`
}

func severityToTurkish(severity string) string {
	s := strings.TrimSpace(strings.ToLower(severity))
	switch s {
	case "critical", "kritik":
		return "kritik"
	case "high", "yüksek":
		return "yüksek"
	case "medium", "orta":
		return "orta"
	case "low", "düşük":
		return "düşük"
	default:
		if s != "" {
			return severity
		}
		return "düşük"
	}
}

func NewHandler(ruleManager *rules.Manager) *Handler {
	h := &Handler{
		ruleManager:   ruleManager,
		analyzer:      analyzer.NewAnalyzer(ruleManager),
		tailer:        tailer.NewTailer(ruleManager),
		alerts:        make([]AlertResponse, 0),
		wsConnections: make(map[*websocket.Conn]struct{}),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
	go h.collectAlerts()
	return h
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, LoginResponse{Success: false, Error: "Geçersiz istek"})
		return
	}
	if req.Username != authUsername || req.Password != authPassword {
		c.JSON(http.StatusUnauthorized, LoginResponse{Success: false, Error: "Kullanıcı adı veya şifre hatalı"})
		return
	}
	c.JSON(http.StatusOK, LoginResponse{Success: true, Token: authToken})
}

func AuthToken() string {
	return authToken
}

func (h *Handler) collectAlerts() {
	for alert := range h.tailer.Alerts() {
		summary := parser.ParseLogLineToSummary(alert.Line)
		if summary == "" {
			summary = alert.Line
		}
		alertResp := AlertResponse{
			Timestamp:    alert.Timestamp,
			Source:       alert.Source,
			LogFile:      alert.LogFile,
			Line:         alert.Line,
			Summary:      summary,
			MatchedRules: alert.MatchedRules,
			Severity:     severityToTurkish(alert.Severity),
		}

		h.mu.Lock()
		h.alerts = append(h.alerts, alertResp)
		if len(h.alerts) > 1000 {
			h.alerts = h.alerts[len(h.alerts)-1000:]
		}
		h.mu.Unlock()

		h.broadcastAlert(alertResp)
	}
}

func (h *Handler) broadcastAlert(alert AlertResponse) {
	h.wsMu.RLock()
	conns := make([]*websocket.Conn, 0, len(h.wsConnections))
	for conn := range h.wsConnections {
		conns = append(conns, conn)
	}
	h.wsMu.RUnlock()

	for _, conn := range conns {
		if err := conn.WriteJSON(alert); err != nil {
			h.wsMu.Lock()
			delete(h.wsConnections, conn)
			h.wsMu.Unlock()
			conn.Close()
		}
	}
}

func (h *Handler) GetRules(c *gin.Context) {
	rules := h.ruleManager.GetRules()
	c.JSON(http.StatusOK, rules)
}

func (h *Handler) GetLogFiles(c *gin.Context) {
	files := h.ruleManager.GetLogFiles()
	c.JSON(http.StatusOK, files)
}

func (h *Handler) AnalyzeFiles(c *gin.Context) {
	var req AnalyzeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Files) == 0 {
		logFiles := h.ruleManager.GetEnabledLogFiles()
		for _, file := range logFiles {
			req.Files = append(req.Files, file.Path)
		}
	}
	entries, err := h.analyzer.AnalyzeMultipleFiles(req.Files)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var alerts []AlertResponse
	for _, entry := range entries {
		summary := entry.Summary
		if summary == "" {
			summary = entry.Line
		}
		alerts = append(alerts, AlertResponse{
			Timestamp:    parseTime(entry.Timestamp),
			Source:       entry.Source,
			LogFile:      entry.LogFile,
			Line:         entry.Line,
			Summary:      summary,
			MatchedRules: entry.MatchedRules,
			Severity:     severityToTurkish(entry.Severity),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"entries": alerts,
		"count":   len(alerts),
	})
}

func (h *Handler) StartTailing(c *gin.Context) {
	var req TailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.Files) == 0 {
		logFiles := h.ruleManager.GetEnabledLogFiles()
		for _, file := range logFiles {
			req.Files = append(req.Files, file.Path)
		}
	}
	var started []string
	var failed []string

	for _, filePath := range req.Files {
		if err := h.tailer.StartWatching(filePath); err != nil {
			failed = append(failed, filePath)
		} else {
			started = append(started, filePath)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"started": started,
		"failed":  failed,
	})
}

func (h *Handler) StopTailing(c *gin.Context) {
	var req TailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	for _, filePath := range req.Files {
		h.tailer.StopWatching(filePath)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Tailing stopped"})
}

func (h *Handler) GetAlerts(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	limit := 100
	if len(h.alerts) < limit {
		limit = len(h.alerts)
	}

	alerts := h.alerts
	if len(alerts) > limit {
		alerts = alerts[len(alerts)-limit:]
	}

	c.JSON(http.StatusOK, alerts)
}

func (h *Handler) WebSocketAlerts(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer conn.Close()

	h.wsMu.Lock()
	h.wsConnections[conn] = struct{}{}
	h.wsMu.Unlock()
	defer func() {
		h.wsMu.Lock()
		delete(h.wsConnections, conn)
		h.wsMu.Unlock()
	}()

	h.mu.RLock()
	recentAlerts := h.alerts
	if len(recentAlerts) > 50 {
		recentAlerts = recentAlerts[len(recentAlerts)-50:]
	}
	h.mu.RUnlock()

	for _, alert := range recentAlerts {
		if err := conn.WriteJSON(alert); err != nil {
			return
		}
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (h *Handler) GetStats(c *gin.Context) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	severityCount := make(map[string]int)
	for _, alert := range h.alerts {
		severityCount[severityToTurkish(alert.Severity)]++
	}

	watchedFiles := h.tailer.GetWatchedFiles()
	isTailing := len(watchedFiles) > 0

	stats := StatsResponse{
		TotalAlerts:      len(h.alerts),
		SeverityCount:    severityCount,
		ActiveRules:      len(h.ruleManager.GetEnabledRules()),
		WatchedFiles:     len(watchedFiles),
		IsTailing:        isTailing,
		WatchedFilesList: watchedFiles,
	}

	c.JSON(http.StatusOK, stats)
}

func parseTime(timeStr string) time.Time {
	formats := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"Jan 2 15:04:05",
		time.RFC3339Nano,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t
		}
	}

	return time.Now()
}
