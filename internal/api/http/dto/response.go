package dto

import (
	"encoding/json"
	"fmt"
	"job-queue/internal/models"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type GetResponse struct {
	Response
	JobID       int64            `json:"id"`
	Queue       models.QueueType `json:"queue"`
	Payload     json.RawMessage  `json:"payload"`
	JobStatus   string           `json:"job_status"`
	AvailableAt time.Time        `json:"available_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

type SaveResponse struct {
	Response
	JobID       int64            `json:"id"`
	Queue       models.QueueType `json:"queue"`
	JobStatus   string           `json:"job_status"`
	AvailableAt time.Time        `json:"available_at"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func OK() Response {
	return Response{
		Status: StatusOK,
	}
}

func Error(msg string) Response {
	return Response{
		Status: StatusError,
		Error:  msg,
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string

	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is a required field", err.Field()))
		case "json":
			errMsgs = append(errMsgs,
				fmt.Sprintf("field %s must be valid JSON", err.Field()),
			)
		case "oneof":
			errMsgs = append(errMsgs,
				fmt.Sprintf(
					"field %s must be one of allowed values (emails, payment, notification)",
					err.Field(),
				),
			)
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is not valid", err.Field()))
		}
	}

	return Response{
		Status: StatusError,
		Error:  strings.Join(errMsgs, ", "),
	}
}
