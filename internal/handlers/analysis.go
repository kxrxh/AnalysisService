package handlers

import (
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/models"
	"csort.ru/analysis-service/internal/services"
	"csort.ru/analysis-service/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"strconv"
)

var analysisHandlerLog = logger.GetLogger("handlers.analysis")

type AnalysisHandler struct {
	service *services.AnalysisService
}

func NewAnalysisHandler(service *services.AnalysisService) *AnalysisHandler {
	return &AnalysisHandler{
		service: service,
	}
}

func (h *AnalysisHandler) GetAnalyses(c *fiber.Ctx) error {
	// Extract Telegram-User-ID from header
	userIDStr := c.Get("Telegram-User-ID")
	if userIDStr == "" {
		analysisHandlerLog.Error().Msg("Telegram-User-ID header is missing")
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Telegram-User-ID header is required"})
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Str("userIDStr", userIDStr).Msg("Invalid Telegram-User-ID format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid Telegram-User-ID format"})
	}

	var params models.GetAnalysesPaginatedRequest
	if err := c.QueryParser(&params); err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Error parsing query params")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid query params"})
	}

	paginatedResponse, err := h.service.GetAnalyses(c.Context(), userID, params)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Int64("userID", userID).Msg("Error getting analyses")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(paginatedResponse)
}

func (h *AnalysisHandler) GetAnalysisByID(c *fiber.Ctx) error {
	id, err := utils.ParseParamWithType[string](c, "id")
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to parse ID parameter")
		return err
	}

	analysis, err := h.service.GetAnalysisByID(c.Context(), id)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Str("analysisID", id).Msg("Error getting analysis by id")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(analysis)
}

func (h *AnalysisHandler) GetAnalysisObjects(c *fiber.Ctx) error {
	id, err := utils.ParseParamWithType[string](c, "id")
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to parse ID parameter")
		return err
	}

	analysisHandlerLog.Info().Str("id_analysis", id).Msg("Fetching objects for analysis")

	objects, err := h.service.GetObjectsByAnalysisID(c.Context(), id)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Str("analysisID", id).Msg("Error getting analysis objects by id")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "internal server error"})
	}

	return c.JSON(objects)
}
