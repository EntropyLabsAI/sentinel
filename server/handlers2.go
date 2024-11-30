package sentinel

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func apiCreateNewChatCompletionRequestHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	// Get payload
	var payload CreateNewChatCompletionRequestBody
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// b64 decode
	decoded, err := base64.StdEncoding.DecodeString(payload.RequestData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid base64 format", err.Error())
		return
	}

	// Check that the request is valid
	var v openai.ChatCompletionRequest
	err = json.Unmarshal(decoded, &v)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	// Unmarshal into json
	jsonRequest, err := json.Marshal(v)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	id, err := store.CreateChatRequest(ctx, jsonRequest)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Error creating chat request", err.Error())
		return
	}

	respondJSON(w, id, http.StatusOK)
}

func apiCreateNewChatCompletionResponseHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	fmt.Printf("CreateNewChatCompletionResponseHandler: %v\n", runId)

	// Get payload
	var payload CreateNewChatCompletionResponseBody
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	// b64 decode
	decoded, err := base64.StdEncoding.DecodeString(payload.ResponseData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid base64 format", err.Error())
		return
	}

	// Check that the response is valid
	var v openai.ChatCompletionResponse
	err = json.Unmarshal(decoded, &v)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid response format", err.Error())
		return
	}

	jsonResponse, err := json.Marshal(v)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid response format", err.Error())
		return
	}

	id, err := store.CreateChatResponse(ctx, jsonResponse)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Error creating chat response", err.Error())
		return
	}

	fmt.Printf("%s\n", string(jsonResponse))

	respondJSON(w, id, http.StatusOK)
}
