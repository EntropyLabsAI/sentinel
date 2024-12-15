package asteroid

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/google/uuid"
)

type AnthropicConverter struct {
	store ToolStore
}

func (c *AnthropicConverter) ToAsteroidMessages(
	ctx context.Context,
	requestData, responseData []byte,
	runId uuid.UUID,
) ([]AsteroidMessage, error) {
	var messageRequest struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}
	if err := json.Unmarshal(requestData, &messageRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message request: %w", err)
	}

	var messageResponse anthropic.Message
	if err := json.Unmarshal(responseData, &messageResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message response: %w", err)
	}

	asteroidMsgs := make([]AsteroidMessage, 0)

	// Convert request messages
	for _, msg := range messageRequest.Messages {
		id := uuid.New()
		var msgType MessageType
		var msgContent string
		var b64 string

		// Try to unmarshal as string first
		var strContent string
		if err := json.Unmarshal(msg.Content, &strContent); err != nil {
			// If not a string, try to unmarshal as array of content blocks
			var arrayContent []struct {
				Type   string `json:"type"`
				Text   string `json:"text,omitempty"`
				Source struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type,omitempty"`
					Data      string `json:"data,omitempty"`
				} `json:"source,omitempty"`
			}
			if err := json.Unmarshal(msg.Content, &arrayContent); err != nil {
				return nil, fmt.Errorf("content must be either string or valid content array: %w", err)
			}

			// Process array content
			for _, content := range arrayContent {
				switch content.Type {
				case "image":
					msgType = ImageUrl
					msgContent = content.Source.Data
				case "text":
					msgType = Text
					msgContent = content.Text
				}
			}
		} else {
			// Handle simple string content
			msgType = Text
			msgContent = strContent
		}

		b64 = base64.StdEncoding.EncodeToString([]byte(msgContent))

		converted := AsteroidMessage{
			Id:        &id,
			Role:      AsteroidMessageRole(msg.Role),
			Type:      &msgType,
			Content:   msgContent,
			Data:      &b64,
			ToolCalls: &[]AsteroidToolCall{},
		}
		asteroidMsgs = append(asteroidMsgs, converted)
	}

	// Convert response message
	converted, err := c.convertMessage(ctx, messageResponse, runId)
	if err != nil {
		return nil, fmt.Errorf("failed to convert response message: %w", err)
	}
	asteroidMsgs = append(asteroidMsgs, converted)

	return asteroidMsgs, nil
}

func (c *AnthropicConverter) ToAsteroidChoices(
	ctx context.Context,
	responseData []byte,
	runId uuid.UUID,
) ([]AsteroidChoice, error) {
	var messageResponse anthropic.Message
	if err := json.Unmarshal(responseData, &messageResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message response: %w", err)
	}

	message, err := c.convertMessage(ctx, messageResponse, runId)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message: %w", err)
	}

	id := uuid.New().String()
	choice := AsteroidChoice{
		AsteroidId:   id,
		Index:        0,
		Message:      message,
		FinishReason: AsteroidChoiceFinishReason(messageResponse.StopReason),
	}

	return []AsteroidChoice{choice}, nil
}

func (c *AnthropicConverter) ValidateB64EncodedRequest(encodedData string) ([]byte, error) {
	decodedRequest, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var request struct {
		Model    string `json:"model"`
		Messages []struct {
			Role    string          `json:"role"`
			Content json.RawMessage `json:"content"`
		} `json:"messages"`
	}

	if err = json.Unmarshal(decodedRequest, &request); err != nil {
		return nil, fmt.Errorf("invalid request format: %w", err)
	}

	for _, msg := range request.Messages {
		var strContent string
		if err := json.Unmarshal(msg.Content, &strContent); err != nil {
			var arrayContent []struct {
				Type   string `json:"type"`
				Text   string `json:"text,omitempty"`
				Source struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type,omitempty"`
					Data      string `json:"data,omitempty"`
				} `json:"source,omitempty"`
			}
			if err := json.Unmarshal(msg.Content, &arrayContent); err != nil {
				return nil, fmt.Errorf("content must be either string or valid content array: %w", err)
			}
		}
	}

	b, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	return b, nil
}

func (c *AnthropicConverter) ValidateB64EncodedResponse(encodedData string) ([]byte, error) {
	decodedResponse, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var response anthropic.Message
	if err = json.Unmarshal(decodedResponse, &response); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	b, err := json.Marshal(response)
	if err != nil {
		return nil, fmt.Errorf("error marshalling response: %w", err)
	}

	return b, nil
}

// Helper method to convert an Anthropic Message to AsteroidMessage
func (c *AnthropicConverter) convertMessage(
	ctx context.Context,
	message anthropic.Message,
	runId uuid.UUID,
) (AsteroidMessage, error) {
	toolCalls := []AsteroidToolCall{}

	// Process content blocks to extract tool calls and text
	var msgContent string
	for _, block := range message.Content {
		switch block := block.AsUnion().(type) {
		case anthropic.TextBlock:
			msgContent += block.Text
		case anthropic.ToolUseBlock:
			toolCall, err := c.convertToolUseBlock(ctx, block, runId)
			if err != nil {
				return AsteroidMessage{}, fmt.Errorf("error converting tool call: %w", err)
			}
			toolCalls = append(toolCalls, *toolCall)
		}
	}

	originalMessageJSON, err := json.Marshal(message)
	if err != nil {
		return AsteroidMessage{}, fmt.Errorf("error marshalling original message: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(originalMessageJSON)

	id := uuid.New()
	msgType := Text

	return AsteroidMessage{
		Id:        &id,
		Role:      AsteroidMessageRole(message.Role),
		ToolCalls: &toolCalls,
		Type:      &msgType,
		Content:   msgContent,
		Data:      &b64,
	}, nil
}

// Helper method to convert an Anthropic MessageParam to AsteroidMessage
func (c *AnthropicConverter) convertMessageParam(
	ctx context.Context,
	message anthropic.MessageParam,
	runId uuid.UUID,
) (AsteroidMessage, error) {
	toolCalls := []AsteroidToolCall{}
	var msgContent string

	blocks := message.Content.Value

	// Process content blocks
	for _, block := range blocks {
		switch block := block.(type) {
		case anthropic.TextBlockParam:
			msgContent += block.Text.Value
		case anthropic.ToolUseBlockParam:
			args, ok := block.Input.Value.(string)
			if !ok {
				return AsteroidMessage{}, fmt.Errorf("error converting block.Input.Value to string")
			}

			toolCall := AsteroidToolCall{
				CallId:    &block.ID.Value,
				Id:        uuid.New(),
				Name:      &block.Name.Value,
				Arguments: &args,
			}
			toolCalls = append(toolCalls, toolCall)
		}
	}

	originalMessageJSON, err := json.Marshal(message)
	if err != nil {
		return AsteroidMessage{}, fmt.Errorf("error marshalling original message: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(originalMessageJSON)

	id := uuid.New()
	msgType := Text

	return AsteroidMessage{
		Id:        &id,
		Role:      AsteroidMessageRole(message.Role.Value),
		ToolCalls: &toolCalls,
		Type:      &msgType,
		Content:   msgContent,
		Data:      &b64,
	}, nil
}

func (c *AnthropicConverter) convertToolUseBlock(
	ctx context.Context,
	toolUse anthropic.ToolUseBlock,
	runId uuid.UUID,
) (*AsteroidToolCall, error) {
	tool, err := c.store.GetToolFromNameAndRunId(ctx, toolUse.Name, runId)
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolUse.Name)
	}

	args := string(toolUse.Input)

	return &AsteroidToolCall{
		CallId:    &toolUse.ID,
		Id:        uuid.New(),
		ToolId:    *tool.Id,
		Name:      &toolUse.Name,
		Arguments: &args,
	}, nil
}
