package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"stockmind-go/internal/client"
	"stockmind-go/internal/config"
	"stockmind-go/internal/handler"
	"stockmind-go/internal/service"
	"stockmind-go/internal/store"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := store.NewSQLiteStore(cfg.SQLite.Path)
	if err != nil {
		log.Fatalf("Failed to init SQLite: %v", err)
	}

	claudeClient := client.NewClaudeClient(cfg.Claude)
	dataClient := client.NewDataClient(cfg.DataService)
	chatSvc := service.NewChatService(claudeClient, dataClient, db)
	h := handler.NewHandler(chatSvc, db)

	r := gin.Default()
	r.Use(handler.CORSMiddleware())
	h.RegisterRoutes(r)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("StockMind Go server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
