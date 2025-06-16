package server

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"ObservabilityServer/internal/database"
	"ObservabilityServer/internal/model"
)

type Server struct {
	port int

	db database.Service
}

func NewServer(config model.Config) *http.Server {
	newServer := &Server{
		port: config.Port,

		db: database.New(config.Database),
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", newServer.port),
		Handler:      newServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
