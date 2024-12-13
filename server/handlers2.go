package sentinel

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
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

	converter := OpenAIConverter{store}

	jsonRequest, err := converter.ValidateB64EncodedRequest(payload.RequestData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Request: %s", err.Error()), "")
		return
	}

	jsonResponse, err := converter.ValidateB64EncodedResponse(payload.ResponseData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Response: %s", err.Error()), "")
		return
	}

	// Parse out the choices into SentinelChoice objects
	sentinelChoices, err := converter.ToSentinelChoices(ctx, jsonResponse, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error converting choices: %s", err.Error()), "")
		return
	}

	id, err := store.CreateChatRequest(
		ctx,
		runId,
		jsonRequest,
		jsonResponse,
		sentinelChoices,
		"openai",
		[]SentinelMessage{},
	)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error creating chat request: %s", err.Error()), "")
		return
	}

	// Extract all IDs from the created chat structure
	chatIds := extractChatIds(*id, sentinelChoices)

	respondJSON(w, chatIds, http.StatusOK)
}

func invalidateMessage(ctx context.Context, store Store, id uuid.UUID) error {
	msg, err := store.GetMessage(ctx, id)
	if err != nil {
		return fmt.Errorf("error getting message: %w", err)
	}

	if msg.Role == SentinelMessageRoleSentinel {
		return nil
	}

	msg.Role = SentinelMessageRoleSentinel

	err = store.UpdateMessage(ctx, id, *msg)
	if err != nil {
		return fmt.Errorf("error updating message: %w", err)
	}

	fmt.Printf("[DEBUG] Successfully invalidated message with ID: %s\n", id)

	return nil
}

func extractChatIds(chatId uuid.UUID, choices []SentinelChoice) ChatIds {
	result := ChatIds{
		ChatId:    chatId,
		ChoiceIds: make([]ChoiceIds, 0, len(choices)),
	}

	for _, choice := range choices {
		choiceIds := ChoiceIds{
			ChoiceId:    choice.SentinelId,
			MessageId:   choice.Message.Id.String(),
			ToolCallIds: make([]ToolCallIds, 0),
		}

		if choice.Message.ToolCalls != nil {
			for _, toolCall := range *choice.Message.ToolCalls {
				id := toolCall.Id.String()
				toolId := toolCall.ToolId.String()
				choiceIds.ToolCallIds = append(choiceIds.ToolCallIds, ToolCallIds{
					ToolCallId: &id,
					ToolId:     &toolId,
				})
			}
		}

		result.ChoiceIds = append(result.ChoiceIds, choiceIds)
	}

	return result
}

// GetRunMessagesHandler gets the messages for a run
func apiGetRunMessagesHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	requestData, responseData, err := store.GetLatestChat(ctx, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting messages for run", err.Error())
		return
	}

	converter := OpenAIConverter{store}

	sentinelMsgs, err := converter.ToSentinelMessages(ctx, requestData, responseData, runId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error converting messages", err.Error())
		return
	}

	respondJSON(w, sentinelMsgs, http.StatusOK)
}

func apiGetToolCallStateHandler(w http.ResponseWriter, r *http.Request, toolCallId string, store Store) {
	ctx := r.Context()

	// First verify the run exists by using the toolCallId (provided by OpenAI) to get our ToolCall object
	// which will have our Sentinel-generated UUID (Id)
	toolCall, err := store.GetToolCallFromCallId(ctx, toolCallId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool call", err.Error())
		return
	}
	if toolCall == nil {
		sendErrorResponse(w, http.StatusNotFound, "Tool call was not found", "")
		return
	}

	execution := RunExecution{
		Chains:   make([]ChainExecutionState, 0),
		Toolcall: *toolCall,
	}

	// Get all chains for this tool
	chains, err := store.GetSupervisorChains(ctx, toolCall.ToolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting chains", err.Error())
		return
	}

	for _, chain := range chains {
		// Get the chain execution from the chain ID + tool call ID
		chainExecutionId, err := store.GetChainExecutionFromChainAndToolCall(ctx, chain.ChainId, toolCall.Id)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error getting chain execution", err.Error())
			return
		}

		ceState, err := store.GetChainExecutionState(ctx, *chainExecutionId)
		if err != nil {
			sendErrorResponse(w, http.StatusInternalServerError, "error getting chain execution state", err.Error())
			return
		}

		execution.Chains = append(execution.Chains, *ceState)
	}

	status, err := getToolCallStatus(ctx, toolCall.Id, store)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool call status", err.Error())
		return
	}

	execution.Status = status

	respondJSON(w, execution, http.StatusOK)
}
