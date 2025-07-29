package api

import (
	"log"
	"net/http"

	"github.com/Ashmit-Kumar/Auto-Ship/autoship-server/internal/models"
	"github.com/gofiber/fiber/v2"
)

type DeploymentStatusRequest struct {
	ID        string `json:"id" validate:"required"`
	Status    string `json:"status" validate:"required,oneof=success error"`
	Message   string `json:"message,omitempty"`
	Subdomain string `json:"subdomain,omitempty"`
}

func DeploymentStatusHandler(c *fiber.Ctx) error {
	var req DeploymentStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Optional: Update MongoDB record or log it
	log.Printf("Received deployment status: %+v\n", req)

	// You could also persist this status if needed:
	// err := models.UpdateDeploymentStatus(req.ID, req.Status, req.Message, req.Subdomain)

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"message": "Status received",
	})
}

// func UpdateDeploymentStatus(id, status, message, subdomain string) error {
	// Update the deployment status in the database
	// return nil
//}