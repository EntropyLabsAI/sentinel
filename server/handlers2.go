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

	jsonRequest, requestMessages, err := validateAndDecodeRequest(payload.RequestData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Request: %s", err.Error()), "")
		return
	}

	// We need to get the new messages that were appended to the messages array by the client since the last time they
	// sent a request.
	// To do this, we load all of the current messages in msg for this run, then we iterate over the requestMessages
	// and compare the b64 encdoed JSON of the db message and the request message.
	// We handle the following cases:
	// If len(requestMessages) = len(dbMessages)
	// --> if len(requestMessages) == 0 error out
	// --> Check if all of the messages are the same, if so do nothing(?)
	// --> Most likely, the last message in the request no longer matches the last message in the db
	// --> We need to invalidate the messages in the database and add the new message
	// If len(requestMessages) > len(dbMessages)
	// --> Ensure everything up to requestMessages[len(dbMessages)-1] is the same, else error out
	// --> Add the new messages to the database
	// If len(dbMessages) > len(requestMessages)
	// --> Ensure everything up to dbMessages[len(requestMessages)-1] is the same, else error out
	// --> Invalidate the messages in the database after dbMessages[len(requestMessages)-1]
	newRequestMessages, err := newMessagesInRequest(ctx, requestMessages, runId, store)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error adding new messages to DB: %s", err.Error()), "")
		return
	}

	jsonResponse, response, err := validateAndDecodeResponse(payload.ResponseData)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Response: %s", err.Error()), "")
		return
	}

	// Parse out the choices into SentinelChoice objects
	choices, err := convertChoices(ctx, response.Choices, runId, store)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error converting choices: %s", err.Error()), "")
		return
	}

	id, err := store.CreateChatRequest(ctx, runId, jsonRequest, jsonResponse, choices, "openai", newRequestMessages)
	if err != nil {
		sendErrorResponse(w, http.StatusBadRequest, fmt.Sprintf("Error creating chat request: %s", err.Error()), "")
		return
	}

	// Extract all IDs from the created chat structure
	chatIds := extractChatIds(*id, choices)

	respondJSON(w, chatIds, http.StatusOK)
}

func newMessagesInRequest(
	ctx context.Context,
	requestMessages []openai.ChatCompletionMessage,
	runId uuid.UUID,
	store Store,
) ([]SentinelMessage, error) {
	fmt.Printf("[DEBUG] Starting newMessagesInRequest for runId: %s\n", runId)
	fmt.Printf("[DEBUG] Number of request messages: %d\n", len(requestMessages))

	toAdd := make([]SentinelMessage, 0)

	dbMessages, err := store.GetMessagesForRun(ctx, runId, false)
	if err != nil {
		fmt.Printf("[ERROR] Failed to get messages for run: %v\n", err)
		return nil, fmt.Errorf("error getting messages for run: %w", err)
	}
	fmt.Printf("[DEBUG] Number of DB messages: %d\n", len(dbMessages))

	type dbMsg struct {
		Id  uuid.UUID
		Msg openai.ChatCompletionMessage
	}

	var backupMessages []openai.ChatCompletionMessage
	var dbChatMessages []dbMsg
	for i, message := range dbMessages {
		fmt.Printf("[DEBUG] Processing DB message %d\n", i)
		var v openai.ChatCompletionMessage
		decoded, err := base64.StdEncoding.DecodeString(*message.Data)
		if err != nil {
			fmt.Printf("[ERROR] Failed to decode message %d: %v\n", i, err)
			return nil, fmt.Errorf("error decoding message: %w", err)
		}

		if err := json.Unmarshal(decoded, &v); err != nil {
			fmt.Printf("[ERROR] Failed to unmarshal message %d: %v\n", i, err)
			return nil, fmt.Errorf("error unmarshalling message: %w", err)
		}

		dbChatMessages = append(dbChatMessages, dbMsg{
			Id:  *message.Id,
			Msg: v,
		})
		backupMessages = append(backupMessages, v)
		fmt.Printf("[DEBUG] Successfully processed DB message %d with role: %s\n", i, v.Role)
	}

	switch {
	case len(requestMessages) == len(dbChatMessages):
		fmt.Printf("[DEBUG] Case: Equal lengths (request=%d, db=%d)\n", len(requestMessages), len(dbChatMessages))
		if len(requestMessages) == 0 {
			fmt.Printf("[ERROR] No messages found in request\n")
			return nil, fmt.Errorf("no messages parsed out of LLM request JSON")
		}

		invalidatedLastMessage := false
		for i, message := range requestMessages {
			fmt.Printf("[DEBUG] Comparing message %d\n", i)
			equal, err := deepEqual(message, dbChatMessages[i].Msg)
			if err != nil {
				fmt.Printf("[ERROR] Failed to compare messages at index %d: %v\n", i, err)
				return nil, fmt.Errorf("error comparing messages: %w", err)
			}
			if !equal {
				fmt.Printf("[DEBUG] Message mismatch at index %d\n", i)
				if i != len(requestMessages)-1 {
					fmt.Printf("[ERROR] Mismatch occurred before last message at index %d\n", i)
					return nil, fmt.Errorf("message content mismatch at index %d", i)
				}

				fmt.Printf("[DEBUG] Invalidating last message with ID: %s\n", dbChatMessages[i].Id)
				err = invalidateMessage(ctx, store, dbChatMessages[i].Id)
				if err != nil {
					fmt.Printf("[ERROR] Failed to invalidate message: %v\n", err)
					return nil, fmt.Errorf("error invalidating message: %w", err)
				}

				invalidatedLastMessage = true
			}
		}

		if !invalidatedLastMessage {
			fmt.Printf("[DEBUG] All messages are identical\n")
			return nil, fmt.Errorf("all messages in the db and request are the same, strange")
		}
	case len(requestMessages) > len(dbChatMessages):
		fmt.Printf("[DEBUG] Case: More request messages than DB (request=%d, db=%d)\n", len(requestMessages), len(dbChatMessages))

		for i, message := range requestMessages {
			if i < len(dbChatMessages) {
				fmt.Printf("[DEBUG] Comparing existing message %d\n", i)
				equal, err := deepEqual(message, dbChatMessages[i].Msg)
				if err != nil {
					fmt.Printf("[ERROR] Failed to compare messages at index %d: %v\n", i, err)
					return nil, fmt.Errorf("error comparing messages: %w", err)
				}
				if !equal {
					fmt.Printf("[ERROR] Message mismatch at index %d\n", i)
					// JSON stringify both arrays for debugging
					wanted, _ := json.MarshalIndent(requestMessages, "", "  ")
					got, _ := json.MarshalIndent(backupMessages, "", "  ")
					fmt.Printf("[DEBUG] Wanted:\n%+v\n, Got:\n%+v\n", string(wanted), string(got))
					return nil, fmt.Errorf("message content mismatch at index %d: wanted:\n%+v\n, got:\n%+v\n", i, string(wanted), string(got))
				}
			} else {
				break
			}
		}

		fmt.Printf("[DEBUG] Processing %d new messages\n", len(requestMessages)-len(dbChatMessages))
		for i, message := range requestMessages[len(dbChatMessages):] {
			fmt.Printf("[DEBUG] Converting new message %d\n", i)
			sentinelMsg, err := convertMessage(ctx, message, runId, store)
			if err != nil {
				fmt.Printf("[ERROR] Failed to convert message %d: %v\n", i, err)
				return nil, fmt.Errorf("error converting message: %w", err)
			}
			toAdd = append(toAdd, sentinelMsg)
		}

	case len(dbChatMessages) > len(requestMessages):
		fmt.Printf("[DEBUG] Case: More DB messages than request (request=%d, db=%d)\n", len(requestMessages), len(dbChatMessages))

		for i, message := range requestMessages {
			fmt.Printf("[DEBUG] Comparing message %d\n", i)
			if i < len(dbChatMessages) {
				equal, err := deepEqual(message, dbChatMessages[i].Msg)
				if err != nil {
					fmt.Printf("[ERROR] Failed to compare messages at index %d: %v\n", i, err)
					return nil, fmt.Errorf("error comparing messages: %w", err)
				}
				if !equal {
					fmt.Printf("[ERROR] Message mismatch at index %d\n", i)
					return nil, fmt.Errorf("message content mismatch at index %d", i)
				}
			}
		}

		fmt.Printf("[DEBUG] Invalidating %d messages\n", len(dbChatMessages)-len(requestMessages))
		for i, message := range dbChatMessages[len(requestMessages):] {
			fmt.Printf("[DEBUG] Invalidating message %d with ID: %s\n", i, message.Id)
			err = invalidateMessage(ctx, store, message.Id)
			if err != nil {
				fmt.Printf("[ERROR] Failed to invalidate message %d: %v\n", i, err)
				return nil, fmt.Errorf("error invalidating message: %w", err)
			}
		}
	}

	fmt.Printf("[DEBUG] Completed successfully. Number of new messages to add: %d\n", len(toAdd))
	return toAdd, nil
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

func deepEqual(a, b openai.ChatCompletionMessage) (bool, error) {
	jsonA, err := json.Marshal(a)
	if err != nil {
		return false, fmt.Errorf("error marshalling message A: %w", err)
	}

	jsonB, err := json.Marshal(b)
	if err != nil {
		return false, fmt.Errorf("error marshalling message B: %w", err)
	}

	if string(jsonA) != string(jsonB) {
		return false, nil
	}

	return true, nil
}

// validateAndDecodeRequest handles the decoding and validation of the chat completion request
// It splits out the messages and converts them to SentinelMessage objects
func validateAndDecodeRequest(encodedData string) ([]byte, []openai.ChatCompletionMessage, error) {
	decodedRequest, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionRequest
	if err = json.Unmarshal(decodedRequest, &v); err != nil {
		return nil, nil, fmt.Errorf("invalid request format: %w", err)
	}

	marshaledRequest, err := json.Marshal(v)
	if err != nil {
		return nil, nil, fmt.Errorf("error marshalling request: %w", err)
	}

	return marshaledRequest, v.Messages, nil
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

func convertChoices(ctx context.Context, choices []openai.ChatCompletionChoice, runId uuid.UUID, store ToolStore) ([]SentinelChoice, error) {
	var result []SentinelChoice
	for _, choice := range choices {
		message, err := convertMessage(ctx, choice.Message, runId, store)
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

func convertMessage(ctx context.Context, message openai.ChatCompletionMessage, runId uuid.UUID, store ToolStore) (SentinelMessage, error) {
	toolCalls, err := convertToolCalls(ctx, message.ToolCalls, runId, store)
	if err != nil {
		return SentinelMessage{}, fmt.Errorf("error converting tool calls: %w", err)
	}

	// If the message has an image in it, it will look like this:
	// {Role:user Content: Refusal: MultiContent:[{Type:image_url Text: ImageURL:0xc000220320}] Name: FunctionCall:<nil> ToolCalls:[] ToolCallID:}
	// We need to convert this to a SentinelMessage with a type of ImageURL
	// and the content being the image URL

	var msgType MessageType
	var msgContent string
	if message.MultiContent != nil {
		for _, content := range message.MultiContent {
			if content.Type == "image_url" {
				msgType = ImageUrl
				msgContent = string(content.ImageURL.URL)
			}
		}
	} else {
		msgType = Text
		msgContent = message.Content
	}

	originalMessageJSON, err := json.Marshal(message)
	if err != nil {
		return SentinelMessage{}, fmt.Errorf("error marshalling original message: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(originalMessageJSON)

	id := uuid.New()

	sMsg := SentinelMessage{
		Id:        &id,
		Role:      SentinelMessageRole(message.Role),
		ToolCalls: &toolCalls,
		Type:      &msgType,
		Content:   msgContent,
		Data:      &b64,
	}

	// fmt.Printf("\n\nMessage ID: %+v\n", message.ToolCallID)
	// fmt.Printf("Message content: %+v\n", message.Content)
	// fmt.Printf("Sentinel message: %+v\n", sMsg.Content)

	return sMsg, nil
}

func convertToolCalls(ctx context.Context, toolCalls []openai.ToolCall, runId uuid.UUID, store ToolStore) ([]SentinelToolCall, error) {
	var result []SentinelToolCall
	for _, toolCall := range toolCalls {
		toolCall, err := convertToolCall(ctx, toolCall, runId, store)
		if err != nil {
			return nil, fmt.Errorf("error converting tool call: %w", err)
		}
		if toolCall != nil {
			result = append(result, *toolCall)
		}
	}
	return result, nil
}

func convertToolCall(ctx context.Context, toolCall openai.ToolCall, runId uuid.UUID, store ToolStore) (*SentinelToolCall, error) {
	// Get this from the DB
	tool, err := store.GetToolFromNameAndRunId(ctx, toolCall.Function.Name, runId)
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	id := uuid.New().String()

	return &SentinelToolCall{
		Id:        id,
		ToolId:    tool.Id.String(),
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
			MessageId:   choice.Message.Id.String(),
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

// GetRunMessagesHandler gets the messages for a run
func apiGetRunMessagesHandler(w http.ResponseWriter, r *http.Request, runId uuid.UUID, store Store) {
	ctx := r.Context()

	messages, err := store.GetMessagesForRun(ctx, runId, true)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting messages for run", err.Error())
		return
	}

	respondJSON(w, messages, http.StatusOK)
}

func apiGetToolCallStateHandler(w http.ResponseWriter, r *http.Request, toolCallId uuid.UUID, store Store) {
	ctx := r.Context()

	// First verify the run exists
	toolCall, err := store.GetToolCall(ctx, toolCallId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool call", err.Error())
		return
	}
	if toolCall == nil {
		sendErrorResponse(w, http.StatusNotFound, "Run not found", "")
		return
	}

	execution := RunExecution{
		Chains:   make([]ChainExecutionState, 0),
		Toolcall: *toolCall,
	}

	toolId, err := uuid.Parse(toolCall.ToolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error parsing tool id", err.Error())
		return
	}

	// Get all chains for this tool
	chains, err := store.GetSupervisorChains(ctx, toolId)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting chains", err.Error())
		return
	}

	for _, chain := range chains {
		// Get the chain execution from the chain ID + tool call ID
		chainExecutionId, err := store.GetChainExecutionFromChainAndToolCall(ctx, chain.ChainId, toolCallId)
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

	status, err := getToolCallStatus(ctx, toolCallId, store)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "error getting tool call status", err.Error())
		return
	}

	execution.Status = status

	respondJSON(w, execution, http.StatusOK)
}
