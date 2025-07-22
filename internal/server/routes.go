package server

import (
	"csort.ru/analysis-service/internal/handlers"
	"github.com/gofiber/fiber/v2"
)

// Handlers struct to hold all application handlers.
type Handlers struct {
	AnalysisHandler *handlers.AnalysisHandler
	ObjectsHandler  *handlers.ObjectsHandler
}

// defineRoutes creates a list of all application routes.
func defineRoutes(h *Handlers) []Route {
	return []Route{
		{Method: fiber.MethodGet, Path: "/health", Handler: healthCheckHandler},
		{Method: fiber.MethodGet, Path: "/analyses", Handler: h.AnalysisHandler.GetAnalyses},
		{Method: fiber.MethodGet, Path: "/analyses/:id", Handler: h.AnalysisHandler.GetAnalysisByID},
		{Method: fiber.MethodGet, Path: "/analyses/:id/objects", Handler: h.AnalysisHandler.GetAnalysisObjects},
		{Method: fiber.MethodPost, Path: "/objects", Handler: h.ObjectsHandler.GetObjects},
	}
}

// healthCheckHandler provides a simple health check endpoint.
func healthCheckHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "healthy",
		"service": "analysis-service",
		"version": "v1",
	})
}
