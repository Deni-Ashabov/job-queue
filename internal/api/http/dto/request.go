package dto

import (
	"encoding/json"
	"job-queue/internal/models"
)

type SaveRequest struct {
	Queue   models.QueueType `json:"queue" validate:"oneof=emails payment notification"`
	Payload json.RawMessage  `json:"payload" validate:"required,json"`
}
