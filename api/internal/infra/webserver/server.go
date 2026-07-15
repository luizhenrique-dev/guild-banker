package webserver

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"

	"github.com/luizhenrique-dev/guild-banker/api/config"
	"github.com/luizhenrique-dev/guild-banker/api/internal/fixed_expense"
	"github.com/luizhenrique-dev/guild-banker/api/internal/guild"
	"github.com/luizhenrique-dev/guild-banker/api/internal/importer"
	"github.com/luizhenrique-dev/guild-banker/api/internal/transaction"
)

type Server struct {
	router       *gin.Engine
	httpServer   *http.Server
	cfg          *config.Config
	guild        *guild.Handler
	fixedExpense *fixedexpense.Handler
	transaction  *transaction.Handler
	importer     *importer.Handler
}

func NewServer(cfg *config.Config, db *sqlx.DB) *Server {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PATCH", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	guildStorage := guild.NewStorage(db)
	guildService := guild.NewService(guildStorage)
	guildHandler := guild.NewHandler(guildService)
	fixedExpenseStorage := fixedexpense.NewStorage(db)
	fixedExpenseService := fixedexpense.NewService(fixedExpenseStorage)
	fixedExpenseHandler := fixedexpense.NewHandler(fixedExpenseService)
	transactionStorage := transaction.NewStorage(db)
	transactionService := transaction.NewService(transactionStorage)
	transactionHandler := transaction.NewHandler(transactionService)
	importerStorage := importer.NewStorage(db)
	importerService := importer.NewService(importerStorage, importer.NewC6CSVParser())
	importerHandler := importer.NewHandler(importerService)

	server := &Server{
		router:       router,
		cfg:          cfg,
		guild:        guildHandler,
		fixedExpense: fixedExpenseHandler,
		transaction:  transactionHandler,
		importer:     importerHandler,
	}

	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.WebServerPort),
		Handler: server.router,
	}

	server.registerRoutes()

	return server
}

func (s *Server) Start() error {
	slog.Info("starting webserver...", "port", s.cfg.WebServerPort)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("starting graceful shutdown")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return err
	}

	slog.Info("webserver shutdown complete")

	return nil
}
