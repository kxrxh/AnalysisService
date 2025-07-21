package server

import (
	"fmt"

	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/database"
	"csort.ru/analysis-service/internal/handlers"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/services"

	"github.com/bytedance/sonic"
	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app *fiber.App
	db  *database.DB
}

type Route struct {
	Method  string
	Path    string
	Handler fiber.Handler
}

// New creates a new Fiber application with routes configured.
func New(cfg *config.Config) (*Server, error) {
	db, err := database.New(&database.DatabaseConfig{
		Host:            cfg.DB.Host,
		Port:            cfg.DB.Port,
		User:            cfg.DB.User,
		Password:        cfg.DB.Password,
		Name:            cfg.DB.Name,
		MaxConns:        cfg.DB.MaxConns,
		MinConns:        cfg.DB.MinConns,
		MaxConnLifetime: cfg.DB.MaxConnLifetime,
		MaxConnIdleTime: cfg.DB.MaxConnIdleTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database service: %w", err)
	}

	app := fiber.New(fiber.Config{
		JSONEncoder: sonic.Marshal,
		JSONDecoder: sonic.Unmarshal,
	})
	// Add CORS middleware first to handle OPTIONS requests
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173,http://localhost:3000,http://localhost:8081",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Telegram-User-ID",
		AllowCredentials: true,
	}))

	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger: &logger.Logger,
		Fields: []string{
			"status",
			"method",
			"path",
			"ip",
			"latency",
		},
	}))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, err interface{}) {
			logger.Logger.Error().
				Msg(err.(error).Error())
		},
	}))

	app.Use(middleware.Fmt())

	// Initialize services
	analysisService := services.NewAnalysisService(database.NewQueries(db.Pool))

	// Initialize handlers
	analysisHandler := handlers.NewAnalysisHandler(analysisService)

	handlers := &Handlers{
		AnalysisHandler: analysisHandler,
	}

	// Define and register routes
	routes := defineRoutes(handlers)

	// Create the /api/v1 group
	api := app.Group("/api/v1")

	registerRoutes(api, routes)

	server := &Server{
		app: app,
		db:  db,
	}

	return server, nil
}

// Start starts the Fiber application.
func (s *Server) Start(port string) error {
	return s.app.Listen(fmt.Sprintf(":%s", port))
}

// Shutdown gracefully shuts down the Fiber application.
func (s *Server) Shutdown() error {
	// Close database connections
	if s.db != nil {
		s.db.Close()
	}
	return s.app.Shutdown()
}

func registerRoutes(api fiber.Router, routes []Route) {
	// Register all routes and build the public routes map
	for _, route := range routes {
		api.Add(route.Method, route.Path, route.Handler)
	}
}
