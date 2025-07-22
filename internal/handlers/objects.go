package handlers

import (
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/services"
	"github.com/gofiber/fiber/v2"
)

var objectsHandlerLog = logger.GetLogger("handlers.objects")

type ObjectsHandler struct {
	service *services.ObjectsService
}

func NewObjectsHandler(service *services.ObjectsService) *ObjectsHandler {
	return &ObjectsHandler{
		service: service,
	}
}

type GetObjectsRequest struct {
	Objects []int32 `json:"objects"`
}

func (h *ObjectsHandler) GetObjects(c *fiber.Ctx) error {
	request := GetObjectsRequest{}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request body"})
	}

	objects, err := h.service.GetObjects(c.Context(), request.Objects)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to get objects"})
	}

	return c.JSON(objects)
}
