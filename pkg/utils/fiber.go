package utils

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func ParseParamWithType[T any](c *fiber.Ctx, paramName string, defaultValue ...T) (T, error) {
	var zero T
	param := c.Params(paramName)

	if param == "" {
		if len(defaultValue) > 0 {
			return defaultValue[0], nil
		}
		return zero, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Missing %s parameter", paramName)})
	}

	// Handle type conversion based on T
	var result any
	var err error

	switch any(zero).(type) {
	case string:
		result = param
	case int:
		result, err = strconv.Atoi(param)
	case int32:
		var parsedVal int64
		parsedVal, err = strconv.ParseInt(param, 10, 32)
		if err == nil {
			result = int32(parsedVal)
		}
	case int64:
		result, err = strconv.ParseInt(param, 10, 64)
	case float32:
		result, err = strconv.ParseFloat(param, 32)
	case float64:
		result, err = strconv.ParseFloat(param, 64)
	case bool:
		result, err = strconv.ParseBool(param)
	default:
		return zero, fmt.Errorf("unsupported type for parameter %s", paramName)
	}

	if err != nil {
		return zero, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": fmt.Sprintf("Invalid %s parameter: %v", paramName, err)})
	}

	return result.(T), nil
}
