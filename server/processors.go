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
	humanReviewChan chan SupervisionRequest
	interval        time.Duration
}

func NewProcessor(store Store, humanReviewChan chan SupervisionRequest) *Processor {
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
			if err := p.processPendingSupervisionRequests(ctx); err != nil {
				log.Printf("Error processing pending reviews: %v", err)
			}
		}
	}
}

func (p *Processor) processPendingSupervisionRequests(ctx context.Context) error {
	supervisorRequests, err := p.store.GetPendingSupervisionRequests(ctx)
	if err != nil {
		return fmt.Errorf("error getting pending reviews: %w", err)
	}

	for _, supervisorRequest := range supervisorRequests {
		if err := p.processReview(ctx, supervisorRequest); err != nil {
			log.Printf("Error processing supervisor %s: %v", *supervisorRequest.Id, err)
			continue
		}
	}

	return nil
}

func (p *Processor) processReview(ctx context.Context, supervisionRequest SupervisionRequest) error {
	supervisorId := supervisionRequest.SupervisorId
	if supervisorId == nil {
		return fmt.Errorf("supervisor ID is required but was not provided to processReview")
	}

	supervisor, err := p.store.GetSupervisor(ctx, *supervisorId)
	if err != nil {
		return fmt.Errorf("error getting supervisor: %w", err)
	}

	switch supervisor.Type {
	case Human:
		return p.processHumanReview(ctx, supervisionRequest)
	case Llm:
		return p.processLLMReview(ctx, supervisionRequest)
	case Code:
		return p.processCodeReview(ctx, supervisionRequest)
	default:
		return fmt.Errorf("unknown supervisor type: %s", supervisor.Type)
	}
}

func (p *Processor) processHumanReview(ctx context.Context, supervisionRequest SupervisionRequest) error {
	// Send to supervisor channel for human processing
	select {
	case p.humanReviewChan <- supervisionRequest:
		log.Printf("Sent SupervisionRequest.RequestId %s to human supervisor channel", *supervisionRequest.Id)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("supervisor channel is full")
	}
}

func (p *Processor) processCodeReview(_ context.Context, supervisionRequest SupervisionRequest) error {
	// Implement automated code supervisor processing
	log.Printf("Code supervisor processing not implemented yet for supervisor %s", supervisionRequest.Id)
	return nil
}

func (p *Processor) processLLMReview(ctx context.Context, supervisionRequest SupervisionRequest) error {
	id := uuid.New().String()

	if supervisionRequest.Id == nil || *supervisionRequest.Id == uuid.Nil {
		return fmt.Errorf("can't process LLM supervisor, supervisor request ID is required")
	}

	log.Printf("received new LLM supervisor request, ID: %s", *supervisionRequest.Id)

	// TODO allow LLM reviewer to handle multiple tool choice options
	if len(supervisionRequest.ToolRequests) != 1 {
		return fmt.Errorf("invalid number of tool choices provided for LLM supervisor")
	}

	toolChoice := supervisionRequest.ToolRequests[0]

	// Call the LLM to evaluate the tool_choice
	llmReasoning, decision, err := callLLMForReview(ctx, toolChoice, p.store)
	if err != nil {
		return fmt.Errorf("error calling LLM for supervisor: %w", err)
	}

	resultID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("error parsing UUID: %w", err)
	}

	// Prepare the response
	response := SupervisionResult{
		Id:                   resultID,
		Decision:             decision,
		SupervisionRequestId: *supervisionRequest.Id,
		CreatedAt:            time.Now(),
		Toolrequest:          &toolChoice,
		Reasoning:            llmReasoning,
	}

	// Update the database with the new supervisor result and then add a reviewrequest_status entry
	err = p.store.CreateSupervisionResult(ctx, response)
	if err != nil {
		return fmt.Errorf("error creating supervisor result: %w", err)
	}

	return nil
}
