package sentinel

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

type Processor struct {
	store           Store
	humanReviewChan chan ReviewRequest
	interval        time.Duration
}

func NewProcessor(store Store, humanReviewChan chan ReviewRequest) *Processor {
	return &Processor{
		store:           store,
		humanReviewChan: humanReviewChan,
		interval:        10 * time.Second, // Configurable interval
	}
}

func (p *Processor) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := p.processPendingReviewRequests(ctx); err != nil {
				log.Printf("Error processing pending reviews: %v", err)
			}
		}
	}
}

func (p *Processor) processPendingReviewRequests(ctx context.Context) error {
	reviewRequests, err := p.store.GetPendingReviewRequests(ctx)
	if err != nil {
		return fmt.Errorf("error getting pending reviews: %w", err)
	}

	for _, reviewRequest := range reviewRequests {
		if err := p.processReview(ctx, reviewRequest); err != nil {
			log.Printf("Error processing review %s: %v", *reviewRequest.Id, err)
			continue
		}
	}

	return nil
}

func (p *Processor) processReview(ctx context.Context, reviewRequest ReviewRequest) error {
	// Get tool requests for this review
	toolRequests, err := p.store.GetReviewToolRequests(ctx, *reviewRequest.Id)
	if err != nil {
		return fmt.Errorf("error getting tool requests: %w", err)
	}

	// Process each tool request to determine the supervisor
	for _, toolRequest := range toolRequests {
		supervisor, err := p.store.GetSupervisorFromToolID(ctx, toolRequest.ToolId)
		if err != nil {
			return fmt.Errorf("error getting supervisor for tool %s: %w", toolRequest.ToolId, err)
		}
		if supervisor == nil {
			return fmt.Errorf("no supervisor found for tool %s", toolRequest.ToolId)
		}

		switch supervisor.Type {
		case Human:
			return p.processHumanReview(ctx, reviewRequest)
		case Llm:
			return p.processLLMReview(ctx, reviewRequest)
		case Code:
			return p.processCodeReview(ctx, reviewRequest)
		case All:
			// For "all" type, process with all available methods
			if err := p.processHumanReview(ctx, reviewRequest); err != nil {
				return err
			}
			if err := p.processLLMReview(ctx, reviewRequest); err != nil {
				return err
			}
			if err := p.processCodeReview(ctx, reviewRequest); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unknown supervisor type: %s", supervisor.Type)
		}
	}

	return nil
}

func (p *Processor) processHumanReview(ctx context.Context, reviewRequest ReviewRequest) error {
	// Send to review channel for human processing
	select {
	case p.humanReviewChan <- reviewRequest:
		log.Printf("Sent review %s to human review channel", *reviewRequest.Id)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("review channel is full")
	}
}

func (p *Processor) processCodeReview(_ context.Context, reviewRequest ReviewRequest) error {
	// Implement automated code review processing
	log.Printf("Code review processing not implemented yet for review %s", reviewRequest.Id)
	return nil
}

func (p *Processor) processLLMReview(ctx context.Context, reviewRequest ReviewRequest) error {
	id := uuid.New().String()

	if reviewRequest.Id == nil || *reviewRequest.Id == uuid.Nil {
		return fmt.Errorf("can't process LLM review, review request ID is required")
	}

	log.Printf("received new LLM review request, ID: %s", *reviewRequest.Id)

	// TODO allow LLM reviewer to handle multiple tool choice options
	if len(reviewRequest.ToolRequests) != 1 {
		return fmt.Errorf("invalid number of tool choices provided for LLM review")
	}

	toolChoice := reviewRequest.ToolRequests[0]

	// Call the LLM to evaluate the tool_choice
	llmReasoning, decision, err := callLLMForReview(ctx, toolChoice, p.store)
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
		CreatedAt:       time.Now(),
		Toolrequest:     &toolChoice,
		Reasoning:       llmReasoning,
	}

	// Update the database with the new review result and then add a reviewrequest_status entry
	err = p.store.CreateReviewResult(ctx, response)
	if err != nil {
		return fmt.Errorf("error creating review result: %w", err)
	}

	return nil
}
