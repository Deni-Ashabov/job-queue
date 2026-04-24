package tests

import (
	"encoding/json"
	"fmt"
	"job-queue/internal/api/http/dto"
	"job-queue/internal/models"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
)

const (
	host = "localhost:8082"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	envPath := filepath.Join(wd, "..", ".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println(".env not found at", envPath)
	} else {
		_ = godotenv.Load(envPath)
	}

	os.Exit(m.Run())
}

func TestURLShortener_HappyPath(t *testing.T) {
	godotenv.Load(".env")

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())

	e.POST("/job").
		WithJSON(dto.SaveRequest{
			Queue: "emails",
			Payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
		}).
		WithBasicAuth(
			os.Getenv("HTTP_SERVER_USER"),
			os.Getenv("HTTP_SERVER_PASS"),
		).
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("job_status")
}

func TestURLShortener_SaveJob(t *testing.T) {
	testCases := []struct {
		name         string
		queue        string
		payload      json.RawMessage
		error        string
		expectDelete bool
		status       int
	}{
		{
			name:  "Success",
			queue: "emails",
			payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
			status: http.StatusOK,
		},
		{
			name:  "Wrong queue",
			queue: "emals",
			payload: json.RawMessage(`{
				"to": "user1",
				"subject": "Hello",
				"body": "Some content"
			}`),
			error:  "field Queue must be one of allowed values (emails, payment, notification)",
			status: http.StatusBadRequest,
		},
		{
			name:    "Invalid json",
			queue:   "emails",
			payload: json.RawMessage(`{"queue":"emails","payload":invalid}`),
			error:   "failed to decode request",
			status:  http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			var resp *httpexpect.Object

			if tc.name == "Invalid json" {
				resp = e.POST("/job").
					WithHeader("Content-Type", "application/json").
					WithBytes([]byte(tc.payload)).
					WithBasicAuth(
						os.Getenv("HTTP_SERVER_USER"),
						os.Getenv("HTTP_SERVER_PASS"),
					).
					Expect().
					Status(tc.status).
					JSON().Object()
			} else {
				resp = e.POST("/job").
					WithJSON(dto.SaveRequest{
						Queue:   models.QueueType(tc.queue),
						Payload: tc.payload,
					}).
					WithBasicAuth(
						os.Getenv("HTTP_SERVER_USER"),
						os.Getenv("HTTP_SERVER_PASS"),
					).
					Expect().
					Status(tc.status).
					JSON().Object()
			}

			if tc.error != "" {
				resp.NotContainsKey("job_status")

				resp.Value("error").String().IsEqual(tc.error)

				return
			}
		})
	}
}
