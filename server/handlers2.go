package sentinel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func apiCreateNewChatHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var payload ChatCompletion
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	jsonRequest, err := validateAndDecodeRequest(payload.RequestData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Request: %s", err.Error()), "")
		return
	}

	jsonResponse, err := validateAndDecodeResponse(payload.ResponseData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Response: %s", err.Error()), "")
		return
	}

	id, err := store.CreateChatRequest(ctx, jsonRequest, jsonResponse, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error creating chat request: %s", err.Error()), "")
		return
	}

	respondJSON(w, id, http.StatusOK)
}

// validateAndDecodeRequest handles the decoding and validation of the chat completion request
func validateAndDecodeRequest(encodedData string) ([]byte, error) {
	decodedRequest, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionRequest
	if err = json.Unmarshal(decodedRequest, &v); err != nil {
		return nil, fmt.Errorf("invalid request format: %w", err)
	}

	return json.Marshal(v)
}

// validateAndDecodeResponse handles the decoding and validation of the chat completion response
func validateAndDecodeResponse(encodedData string) ([]byte, error) {
	decodedResponse, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionResponse
	if err = json.Unmarshal(decodedResponse, &v); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	return json.Marshal(v)
}
