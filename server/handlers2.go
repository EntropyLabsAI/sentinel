package sentinel

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

func apiGetToolCallHandler(w http.ResponseWriter, r *http.Request, id uuid.UUID, store Store) {
	toolCall, err := store.GetToolCall(r.Context(), id)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool call", err.Error())
		return
	}

	if toolCall == nil {
		sendErrorResponse(w, http.StatusNotFound, "tool call not found", "")
		return
	}

	respondJSON(w, toolCall, http.StatusOK)
}

func apiCreateNewChatHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	var payload SentinelChat
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid JSON format", err.Error())
		return
	}

	jsonRequest, err := validateAndDecodeRequest(payload.RequestData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Request: %s", err.Error()), "")
		return
	}

	jsonResponse, response, err := validateAndDecodeResponse(payload.ResponseData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Response: %s", err.Error()), "")
		return
	}

	// Parse out the choices into SentinelChoice objects
	choices, err := convertChoices(ctx, response.Choices, store)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error converting choices: %s", err.Error()), "")
		return
	}

	id, err := store.CreateChatRequest(ctx, runId, jsonRequest, jsonResponse, choices, "openai")
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error creating chat request: %s", err.Error()), "")
		return
	}

	// Extract all IDs from the created chat structure
	chatIds := extractChatIds(*id, choices)

	respondJSON(w, chatIds, http.StatusOK)
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
func validateAndDecodeResponse(encodedData string) ([]byte, *openai.ChatCompletionResponse, error) {
	decodedResponse, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionResponse
	if err = json.Unmarshal(decodedResponse, &v); err != nil {
		return nil, nil, fmt.Errorf("invalid response format: %w", err)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling response: %w", err)
	}

	return b, &v, nil
}

func convertChoices(ctx context.Context, choices []openai.ChatCompletionChoice, store ToolStore) ([]SentinelChoice, error) {
	var result []SentinelChoice
	for _, choice := range choices {
		message, err := convertMessage(ctx, choice.Message, store)
		if err != nil {
			return nil, fmt.Errorf("error converting message: %w", err)
		}

		id := uuid.New().String()
		result = append(result, SentinelChoice{
			SentinelId:   id,
			Index:        choice.Index,
			Message:      message,
			FinishReason: SentinelChoiceFinishReason(choice.FinishReason),
		})
	}

	return result, nil
}

func convertMessage(ctx context.Context, message openai.ChatCompletionMessage, store ToolStore) (SentinelMessage, error) {
	toolCalls, err := convertToolCalls(ctx, message.ToolCalls, store)
	if err != nil {
		return SentinelMessage{}, fmt.Errorf("error converting tool calls: %w", err)
	}

	// Hardcode to text for now
	t := Text

	id := uuid.New().String()

	return SentinelMessage{
		Id:        &id,
		Content:   message.Content,
		Role:      SentinelMessageRole(message.Role),
		ToolCalls: &toolCalls,
		Type:      &t,
	}, nil
}

func convertToolCalls(ctx context.Context, toolCalls []openai.ToolCall, store ToolStore) ([]SentinelToolCall, error) {
	var result []SentinelToolCall
	for _, toolCall := range toolCalls {
		toolCall, err := convertToolCall(ctx, toolCall, store)
		if err != nil {
			return nil, fmt.Errorf("error converting tool call: %w", err)
		}
		if toolCall != nil {
			result = append(result, *toolCall)
		}
	}
	return result, nil
}

func convertToolCall(ctx context.Context, toolCall openai.ToolCall, store ToolStore) (*SentinelToolCall, error) {
	// Get this from the DB
	fmt.Printf("Tool call name: %s\n", toolCall.Function.Name)
	tool, err := store.GetToolFromName(ctx, toolCall.Function.Name)
	if err != nil {
		fmt.Printf("Error getting tool: %s\n", err.Error())
		return nil, fmt.Errorf("error getting tool: %w", err)
	}
	if tool == nil {
		fmt.Printf("Tool not found: %s\n", toolCall.Function.Name)
		return nil, fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	id := uuid.New().String()

	return &SentinelToolCall{
		Id:        id,
		ToolId:    tool.Id.String(),
		Type:      SentinelToolCallType(toolCall.Type),
		Name:      &toolCall.Function.Name,
		Arguments: &toolCall.Function.Arguments,
	}, nil
}

func extractChatIds(chatId uuid.UUID, choices []SentinelChoice) ChatIds {
	result := ChatIds{
		ChatId:    chatId,
		ChoiceIds: make([]ChoiceIds, 0, len(choices)),
	}

	for _, choice := range choices {
		choiceIds := ChoiceIds{
			ChoiceId:    choice.SentinelId,
			MessageId:   *choice.Message.Id,
			ToolCallIds: make([]ToolCallIds, 0),
		}

		if choice.Message.ToolCalls != nil {
			for _, toolCall := range *choice.Message.ToolCalls {
				choiceIds.ToolCallIds = append(choiceIds.ToolCallIds, ToolCallIds{
					ToolCallId: &toolCall.Id,
					ToolId:     &toolCall.ToolId,
				})
			}
		}

		result.ChoiceIds = append(result.ChoiceIds, choiceIds)
	}

	return result
}
