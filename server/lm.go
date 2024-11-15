package sentinel

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

// getExplanationFromLLM calls the LLM to get an explanation and a danger score for a given text.
func getExplanationFromLLM(ctx context.Context, text string) (string, string, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are tasked with analysing some code and providing a summary for a technical reader and a danger score out of 3 choices. Please provide a succinct summary and finish with your evaluation of the code's potential danger score, out of 'harmless', 'risky' or 'dangerous'. Give your summary inside <summary></summary> tags and your score inside <score></score> tags. Start your response with <summary> and finish it with </score>. For example: <summary>The code is a simple implementation of a REST API using the Gin framework.</summary><score>harmless</score>",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: "<code>" + text + "</code>",
		},
	}

	response, err := getLLMResponse(ctx, messages, openai.GPT4oMini)
	if err != nil {
		return "", "", err
	}

	// Parse the LLM response to extract the summary and score
	summaryStart := "<summary>"
	summaryEnd := "</summary>"
	scoreStart := "<score>"
	scoreEnd := "</score>"

	summaryIndex := strings.Index(response, summaryStart)
	summaryEndIndex := strings.Index(response, summaryEnd)
	scoreIndex := strings.Index(response, scoreStart)
	scoreEndIndex := strings.Index(response, scoreEnd)

	if summaryIndex == -1 || summaryEndIndex == -1 || scoreIndex == -1 || scoreEndIndex == -1 {
		return "", "", fmt.Errorf("invalid response format")
	}

	summary := response[summaryIndex+len(summaryStart) : summaryEndIndex]
	score := response[scoreIndex+len(scoreStart) : scoreEndIndex]

	return summary, score, nil
}

// getLLMResponse is a helper function that interacts with the OpenAI API and returns the LLM response.
func getLLMResponse(ctx context.Context, messages []openai.ChatCompletionMessage, model string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model:    model,
			Messages: messages,
		},
	)

	if err != nil {
		return "", fmt.Errorf("error creating LLM chat completion: %v", err)
	}

	return resp.Choices[0].Message.Content, nil
}
