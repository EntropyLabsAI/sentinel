package asteroid

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
)

type AsteroidConverter interface {
	ToAsteroidMessages(requestData, responseData []byte) ([]AsteroidMessage, error)
	ToAsteroidChoices(responseData []byte) ([]AsteroidChoice, error)
	ValidateB64EncodedRequest(encodedData string) ([]byte, error)
	ValidateB64EncodedResponse(encodedData string) ([]byte, error)
}

type OpenAIConverter struct {
	store ToolStore
}

func (c *OpenAIConverter) ToAsteroidMessages(
	ctx context.Context,
	requestData, responseData []byte,
	runId uuid.UUID,
) ([]AsteroidMessage, error) {
	var chatRequest openai.ChatCompletionRequest
	if err := json.Unmarshal(requestData, &chatRequest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat request: %w", err)
	}

	var chatResponse openai.ChatCompletionResponse
	if err := json.Unmarshal(responseData, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat response: %w", err)
	}

	openaiMessages := chatRequest.Messages

	asteroidMsgs := make([]AsteroidMessage, 0)
	for _, msg := range openaiMessages {
		converted, err := c.ConvertMessage(ctx, msg, runId)
		if err != nil {
			return nil, fmt.Errorf("failed to convert message: %w", err)
		}

		asteroidMsgs = append(asteroidMsgs, converted)
	}

	// TODO support multiple choices
	firstChoiceMessage := chatResponse.Choices[0].Message
	converted, err := c.ConvertMessage(ctx, firstChoiceMessage, runId)
	if err != nil {
		return nil, fmt.Errorf("failed to convert message: %w", err)
	}

	asteroidMsgs = append(asteroidMsgs, converted)

	return asteroidMsgs, nil
}

func (c *OpenAIConverter) ToAsteroidChoices(
	ctx context.Context,
	responseData []byte,
	runId uuid.UUID,
) ([]AsteroidChoice, error) {
	var chatResponse openai.ChatCompletionResponse
	if err := json.Unmarshal(responseData, &chatResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal chat response: %w", err)
	}

	choices, err := c.ConvertChoices(ctx, chatResponse.Choices, runId)
	if err != nil {
		return nil, fmt.Errorf("failed to convert choices: %w", err)
	}

	return choices, nil
}

func (c *OpenAIConverter) ConvertChoices(
	ctx context.Context,
	choices []openai.ChatCompletionChoice,
	runId uuid.UUID,
) ([]AsteroidChoice, error) {
	var result []AsteroidChoice
	for _, choice := range choices {
		message, err := c.ConvertMessage(ctx, choice.Message, runId)
		if err != nil {
			return nil, fmt.Errorf("error converting message: %w", err)
		}

		id := uuid.New().String()
		result = append(result, AsteroidChoice{
			AsteroidId:   id,
			Index:        choice.Index,
			Message:      message,
			FinishReason: AsteroidChoiceFinishReason(choice.FinishReason),
		})
	}

	return result, nil
}

func (c *OpenAIConverter) ConvertMessage(
	ctx context.Context,
	message openai.ChatCompletionMessage,
	runId uuid.UUID,
) (AsteroidMessage, error) {
	toolCalls, err := c.ConvertToolCalls(ctx, message.ToolCalls, runId)
	if err != nil {
		return AsteroidMessage{}, fmt.Errorf("error converting tool calls: %w", err)
	}

	// If the message has an image in it, it will look like this:
	// {Role:user Content: Refusal: MultiContent:[{Type:image_url Text: ImageURL:0xc000220320}] Name: FunctionCall:<nil> ToolCalls:[] ToolCallID:}
	// We need to convert this to a AsteroidMessage with a type of ImageURL
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
		return AsteroidMessage{}, fmt.Errorf("error marshalling original message: %w", err)
	}
	b64 := base64.StdEncoding.EncodeToString(originalMessageJSON)

	id := uuid.New()

	sMsg := AsteroidMessage{
		Id:        &id,
		Role:      AsteroidMessageRole(message.Role),
		ToolCalls: &toolCalls,
		Type:      &msgType,
		Content:   msgContent,
		Data:      &b64,
	}

	return sMsg, nil
}

func (c *OpenAIConverter) ConvertToolCalls(
	ctx context.Context,
	toolCalls []openai.ToolCall,
	runId uuid.UUID,
) ([]AsteroidToolCall, error) {
	var result []AsteroidToolCall
	for _, toolCall := range toolCalls {
		toolCall, err := c.ConvertToolCall(ctx, toolCall, runId)
		if err != nil {
			return nil, fmt.Errorf("error converting tool call: %w", err)
		}
		if toolCall != nil {
			result = append(result, *toolCall)
		}
	}
	return result, nil
}

func (c *OpenAIConverter) ConvertToolCall(
	ctx context.Context,
	toolCall openai.ToolCall,
	runId uuid.UUID,
) (*AsteroidToolCall, error) {
	tool, err := c.store.GetToolFromNameAndRunId(ctx, toolCall.Function.Name, runId)
	if err != nil {
		return nil, fmt.Errorf("error getting tool: %w", err)
	}
	if tool == nil {
		return nil, fmt.Errorf("tool not found: %s", toolCall.Function.Name)
	}

	id := uuid.New()

	return &AsteroidToolCall{
		CallId:    &toolCall.ID,
		Id:        id,
		ToolId:    *tool.Id,
		Name:      &toolCall.Function.Name,
		Arguments: &toolCall.Function.Arguments,
	}, nil
}

func (c *OpenAIConverter) ValidateB64EncodedRequest(encodedData string) ([]byte, error) {
	decodedRequest, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionRequest
	if err = json.Unmarshal(decodedRequest, &v); err != nil {
		return nil, fmt.Errorf("invalid request format: %w", err)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshalling request: %w", err)
	}

	return b, nil
}

func (c *OpenAIConverter) ValidateB64EncodedResponse(encodedData string) ([]byte, error) {
	decodedResponse, err := base64.StdEncoding.DecodeString(encodedData)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 format: %w", err)
	}

	var v openai.ChatCompletionResponse
	if err = json.Unmarshal(decodedResponse, &v); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("error marshalling response: %w", err)
	}

	return b, nil
}
