package sentinel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

func processLLMReview(ctx context.Context, reviewRequest ReviewRequest, store Store) error {
	id := uuid.New().String()

	log.Printf("received new LLM review request, ID: %s", id)

	// TODO allow LLM reviewer to handle multiple tool choice options
	if len(reviewRequest.ToolRequests) != 1 {
		return fmt.Errorf("invalid number of tool choices provided for LLM review")
	}

	toolChoice := reviewRequest.ToolRequests[0]

	// Call the LLM to evaluate the tool_choice
	llmReasoning, decision, err := callLLMForReview(ctx, toolChoice, store)
	if err != nil {
		return fmt.Errorf("error calling LLM for review: %w", err)
	}

	resultID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("error parsing UUID: %w", err)
	}

	// Prepare the response
	response := ReviewResult{
		Id:              resultID,
		Decision:        decision,
		ReviewRequestId: *reviewRequest.Id,
		CreatedAt:       time.Now().Unix(),
		Toolrequest:     &toolChoice,
		Reasoning:       llmReasoning,
	}

	// Store the completed LLM review
	completedLLMReviews.Store(id, response)

	return nil
}

func processHumanReview(_ context.Context, hub *Hub, review Review, _ Store) error {
	t := time.Now().Unix()
	// Add the review request to the human review queue
	hub.ReviewChan <- review

	log.Printf("received new review request ID %s via API.", review.Id)

	// Create a channel for this review request
	responseChan := make(chan ReviewResult)
	reviewChannels.Store(review.Id, responseChan)

	// Start a goroutine to wait for the response
	go func() {
		select {
		case response := <-responseChan:
			// Store the completed review
			completedHumanReviews.Store(response.Id, response)
			reviewChannels.Delete(response.Id)
			log.Printf("review ID %s completed with decision: %s.", response.Id, response.Decision)
		case <-time.After(reviewTimeout):

			reviewStatus := ReviewStatus{
				Id:        review.Id,
				Status:    Timeout,
				CreatedAt: t,
			}

			// Timeout occurred
			completedHumanReviews.Store(review.Id, reviewStatus)
			reviewChannels.Delete(review.Id)
			log.Printf("review ID %s timed out.", review.Id)
		}
	}()

	return nil
}
