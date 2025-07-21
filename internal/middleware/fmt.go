package middleware

import (
	"net/http"
	"strings"

	"csort.ru/analysis-service/internal/logger"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v2"
)

// SuccessResponse defines the structure for successful API responses.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result"`
}

// ErrorInfo defines the structure for detailed error information.
type ErrorInfo struct {
	Code    int         `json:"code"`              // HTTP Status Code
	Message string      `json:"message"`           // User-friendly message
	Details interface{} `json:"details,omitempty"` // Optional technical details
	Path    string      `json:"path"`              // Path of the request
}

// ErrorResponse defines the structure for error API responses.
type ErrorResponse struct {
	Success bool      `json:"success"`
	Error   ErrorInfo `json:"error"`
}

// Fmt creates a middleware that standardizes JSON API responses.
// It wraps successful responses and formats errors consistently.
func Fmt() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Chain execution to the next handler. If an error occurs in the chain,
		// handle it. Otherwise, format the response set by the handler.
		if err := c.Next(); err != nil {
			return handleNextError(c, err)
		}
		return formatResponse(c)
	}
}

var formatLogger = logger.GetLogger("middleware.response_formatter")

// handleNextError handles errors returned directly from the c.Next() call.
// These are typically errors from other middleware or Fiber itself.
func handleNextError(c *fiber.Ctx, err error) error {
	statusCode := fiber.StatusInternalServerError // Default
	message := "Internal Server Error"
	var details interface{}

	if e, ok := err.(*fiber.Error); ok {
		// Use Fiber's specific error code and message
		statusCode = e.Code
		message = e.Message
		if statusCode == fiber.StatusNotFound {
			formatLogger.Debug().Err(e).Int("status", statusCode).Str("path", c.Path()).Msg("Fiber 404 error caught")
		} else {
			formatLogger.Warn().Err(e).Int("status", statusCode).Str("path", c.Path()).Msg("Fiber error caught")
		}
	} else {
		// For other unexpected errors, log them as critical
		details = err.Error()
		formatLogger.Error().Err(err).Str("path", c.Path()).Msg("Unhandled error from c.Next()")
	}

	response := ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    statusCode,
			Message: message,
			Details: details,
			Path:    c.Path(),
		},
	}

	c.Status(statusCode).Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	// Prevent Fiber from sending its default error body alongside our JSON
	c.Response().SetBody(nil)
	return c.JSON(response)
}

// formatResponse checks the response set by a handler and formats it if it's JSON.
func formatResponse(c *fiber.Ctx) error {
	statusCode := c.Response().StatusCode()
	contentType := string(c.Response().Header.ContentType())
	body := c.Response().Body()

	// Skip formatting for non-JSON responses (e.g., file downloads, HTML)
	if !strings.Contains(contentType, fiber.MIMEApplicationJSON) {
		formatLogger.Debug().Str("contentType", contentType).Str("path", c.Path()).Int("status", statusCode).Msg("Skipping formatting for non-JSON content")
		return nil // Pass through original response
	}

	// Clear the original body and set the correct content type for our formatted response
	c.Response().SetBody(nil)
	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	// Format based on status code range
	if statusCode >= http.StatusBadRequest {
		return formatErrorBody(c, statusCode, body)
	}

	if statusCode >= http.StatusOK && statusCode < http.StatusMultipleChoices {
		return formatSuccessBody(c, statusCode, body)
	}

	// For other status codes (e.g., 3xx redirects), pass through without formatting.
	formatLogger.Debug().Int("status", statusCode).Str("path", c.Path()).Msg("Non-success/error status, skipping formatting")
	if len(body) > 0 {
		return c.Send(body) // Send original body if it exists
	}
	return nil
}

// formatErrorBody creates a standardized ErrorResponse.
func formatErrorBody(c *fiber.Ctx, statusCode int, body []byte) error {
	var handlerErrorData fiber.Map
	errorMessage := ""
	var errorDetails interface{}

	// Try to parse the handler's original error body for custom messages/details
	if len(body) > 0 && sonic.Unmarshal(body, &handlerErrorData) == nil {
		if msg, ok := handlerErrorData["error"].(string); ok {
			errorMessage = msg
		}
		errorDetails = handlerErrorData["details"] // Preserve details if provided
	}

	// If no message was parsed, use the default HTTP status text
	if errorMessage == "" {
		errorMessage = http.StatusText(statusCode)
		if errorMessage == "" {
			errorMessage = "Unknown Error" // Fallback for unrecognized codes
		}
		// If we failed to parse but had an original body, use it as details
		if len(body) > 0 && errorDetails == nil {
			errorDetails = string(body)
		}
	}

	response := ErrorResponse{
		Success: false,
		Error: ErrorInfo{
			Code:    statusCode,
			Message: errorMessage,
			Details: errorDetails,
			Path:    c.Path(),
		},
	}
	return c.Status(statusCode).JSON(response)
}

// formatSuccessBody creates a standardized SuccessResponse.
func formatSuccessBody(c *fiber.Ctx, statusCode int, body []byte) error {
	// HTTP 204 No Content must have an empty body.
	if statusCode == http.StatusNoContent {
		formatLogger.Debug().Int("status", statusCode).Str("path", c.Path()).Msg("Handling 204 No Content")
		return c.SendStatus(http.StatusNoContent)
	}

	var data interface{}
	if len(body) > 0 {
		// Assume the original body is the JSON data payload
		if err := sonic.Unmarshal(body, &data); err != nil {
			formatLogger.Error().Err(err).Str("body", string(body)).Str("path", c.Path()).Msg("Failed to unmarshal success body, returning error")
			// Return a formatted internal error if the handler returned invalid JSON
			return formatErrorBody(c, http.StatusInternalServerError, []byte(`{"error":"Failed to process server response", "details":"Handler returned invalid JSON"}`))
		}
	} else {
		// If body is empty for a 2xx response (not 204), represent data as null
		data = nil
	}

	response := SuccessResponse{
		Success: true,
		Result:  data,
	}

	return c.Status(statusCode).JSON(response)
}
