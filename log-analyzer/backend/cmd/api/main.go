package main

import (
	"log"
	"os"

	"log-analyzer/backend/cmd/api/handlers"
	"log-analyzer/backend/internal/rules"

	"github.com/gin-gonic/gin"
)

func main() {
	configPath := "config/rules.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	ruleManager, err := rules.NewManager(configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	handler := handlers.NewHandler(ruleManager)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})
	api := r.Group("/api")
	api.Use(func(c *gin.Context) {
		if c.Request.URL.Path == "/api/login" && c.Request.Method == "POST" {
			c.Next()
			return
		}
		token := ""
		if auth := c.GetHeader("Authorization"); len(auth) > 7 && auth[:7] == "Bearer " {
			token = auth[7:]
		}
		if token == "" && c.Request.URL.Path == "/api/tail/ws" {
			token = c.Query("token")
		}
		if token != handlers.AuthToken() {
			c.AbortWithStatusJSON(401, gin.H{"error": "Yetkisiz"})
			return
		}
		c.Next()
	})
	api.POST("/login", handler.Login)
	api.GET("/rules", handler.GetRules)
	api.GET("/logfiles", handler.GetLogFiles)
	api.POST("/analyze", handler.AnalyzeFiles)
	api.POST("/tail/start", handler.StartTailing)
	api.POST("/tail/stop", handler.StopTailing)
	api.GET("/tail/alerts", handler.GetAlerts)
	api.GET("/tail/ws", handler.WebSocketAlerts)
	api.GET("/stats", handler.GetStats)
	r.Static("/assets", "./frontend/dist/assets")
	r.StaticFile("/", "./frontend/dist/index.html")
	r.NoRoute(func(c *gin.Context) {
		c.File("./frontend/dist/index.html")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
