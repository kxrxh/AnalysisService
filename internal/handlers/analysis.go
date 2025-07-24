package handlers

import (
	"strconv"

	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/models"
	"csort.ru/analysis-service/internal/services"
	"csort.ru/analysis-service/pkg/utils"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
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

func (h *AnalysisHandler) CreateAnalysis(c *fiber.Ctx) error {
	product := c.FormValue("product")
	userID := c.FormValue("userID")
	fileHeader, err := c.FormFile("files")
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to get file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	file, err := fileHeader.Open()
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to open file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer file.Close()

	status, _, body, err := h.service.ProxyAnalysisAPICall(c.Context(), product, userID, fileHeader.Filename, file)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to contact analysis API")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to contact analysis API"})
	}

	var resp struct {
		Response string `json:"Response"`
	}

	if err := sonic.Unmarshal(body, &resp); err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to unmarshal response from analysis API")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid response from analysis API"})
	}

	switch status {
	case fiber.StatusOK:
		analysis, err := h.service.GetAnalysisByID(c.Context(), resp.Response)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch analysis"})
		}
		return c.JSON(analysis)
	case fiber.StatusBadRequest:
		analysisHandlerLog.Error().Int("status", status).Msg("Bad request from analysis API")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": resp.Response})
	default:
		analysisHandlerLog.Error().Int("status", status).Msg("Unexpected response from analysis API")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "unexpected response from analysis API"})
	}
}
