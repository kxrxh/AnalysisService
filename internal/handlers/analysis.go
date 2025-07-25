package handlers

import (
	"fmt"
	"strconv"
	"time"

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

	// Validate required fields
	if product == "" {
		analysisHandlerLog.Error().Msg("Product field is missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "product field is required"})
	}
	if userID == "" {
		analysisHandlerLog.Error().Msg("UserID field is missing")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "userID field is required"})
	}

	analysisHandlerLog.Info().Str("product", product).Str("userID", userID).Msg("Creating analysis")

	fileHeader, err := c.FormFile("files")
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to get file")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}

	// Validate file
	if fileHeader == nil {
		analysisHandlerLog.Error().Msg("File header is nil")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file is required"})
	}
	if fileHeader.Size == 0 {
		analysisHandlerLog.Error().Msg("File is empty")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "file cannot be empty"})
	}

	analysisHandlerLog.Info().
		Str("filename", fileHeader.Filename).
		Int64("filesize", fileHeader.Size).
		Str("content_type", fileHeader.Header.Get("Content-Type")).
		Msg("File details")

	file, err := fileHeader.Open()
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to open file")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to open file"})
	}
	defer file.Close()

	// Call the updated ProxyAnalysisAPICall method
	status, headers, body, err := h.service.ProxyAnalysisAPICall(c.Context(), product, userID, fileHeader.Filename, file)
	if err != nil {
		analysisHandlerLog.Error().Err(err).Msg("Failed to contact analysis API")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": "failed to contact analysis API"})
	}

	analysisHandlerLog.Info().
		Int("status", status).
		Str("response_headers", fmt.Sprintf("%v", headers)).
		Msg("Analysis API response")

	switch status {
	case fiber.StatusOK:
		// Try to parse JSON response
		var resp struct {
			Response string `json:"Response"`
		}
		if err := sonic.Unmarshal(body, &resp); err != nil {
			analysisHandlerLog.Error().
				Err(err).
				Str("body", string(body)).
				Msg("Failed to unmarshal success response from analysis API")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "invalid response format from analysis API"})
		}

		analysisHandlerLog.Info().Str("analysisID", resp.Response).Msg("Analysis created successfully")

		// Wait 2-5 seconds to allow analysis to be added to DB
		time.Sleep(3 * time.Second)

		analysis, err := h.service.GetAnalysisByID(c.Context(), resp.Response)
		if err != nil {
			analysisHandlerLog.Error().
				Err(err).
				Str("analysisID", resp.Response).
				Msg("Failed to fetch analysis")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to fetch analysis"})
		}

		return c.JSON(analysis)

	case fiber.StatusBadRequest:
		// Try to parse JSON, but fallback to plain text
		var resp struct {
			Response string `json:"Response"`
		}
		errorMsg := string(body)
		if err := sonic.Unmarshal(body, &resp); err == nil && resp.Response != "" {
			errorMsg = resp.Response
		}

		analysisHandlerLog.Error().
			Int("status", status).
			Str("response", errorMsg).
			Msg("Bad request from analysis API")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": errorMsg})

	case fiber.StatusInternalServerError:
		// Handle 500 errors specifically
		errorMsg := string(body)
		var resp struct {
			Response string `json:"Response"`
		}
		if err := sonic.Unmarshal(body, &resp); err == nil && resp.Response != "" {
			errorMsg = resp.Response
		}

		analysisHandlerLog.Error().
			Int("status", status).
			Str("response", errorMsg).
			Msg("Internal server error from analysis API")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("analysis API internal error: %s", errorMsg)})

	default:
		// For other errors, the response might be plain text
		errorMsg := string(body)
		var resp struct {
			Response string `json:"Response"`
		}
		if err := sonic.Unmarshal(body, &resp); err == nil && resp.Response != "" {
			errorMsg = resp.Response
		}

		analysisHandlerLog.Error().
			Int("status", status).
			Str("response", errorMsg).
			Msg("Unexpected response from analysis API")
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"error": fmt.Sprintf("analysis API error (status %d): %s", status, errorMsg)})
	}
}
