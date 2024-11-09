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
		interval:        2 * time.Second, // Configurable interval
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
	supervisorRequests, err := p.store.GetSupervisionRequestsForStatus(ctx, Pending)
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

	if supervisorId == uuid.Nil {
		return fmt.Errorf("supervisor ID is required but was not provided to processReview")
	}

	supervisor, err := p.store.GetSupervisor(ctx, supervisorId)
	if err != nil {
		return fmt.Errorf("error getting supervisor: %w", err)
	}

	switch supervisor.Type {
	case SupervisorTypeHumanSupervisor:
		return p.processHumanReview(ctx, supervisionRequest)
	case SupervisorTypeClientSupervisor:
		return p.processClientReview(ctx, supervisionRequest)
	case SupervisorTypeNoSupervisor:
		return p.processNoSupervisionReview(ctx, supervisionRequest)
	default:
		return fmt.Errorf("unknown supervisor type: %s", supervisor.Type)
	}
}

func (p *Processor) processHumanReview(ctx context.Context, supervisionRequest SupervisionRequest) error {
	// Send to supervisor channel for human processing
	select {
	case p.humanReviewChan <- supervisionRequest:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return fmt.Errorf("supervisor channel is full")
	}
}

func (p *Processor) processNoSupervisionReview(ctx context.Context, supervisionRequest SupervisionRequest) error {
	log.Printf("Processing no supervision review for supervision request %s", *supervisionRequest.Id)
	// Update the supervision request status to completed as no processing is required.
	status := SupervisionStatus{
		Status:               Completed,
		CreatedAt:            time.Now(),
		SupervisionRequestId: supervisionRequest.Id,
	}

	if err := p.store.CreateSupervisionStatus(ctx, *supervisionRequest.Id, status); err != nil {
		return fmt.Errorf("error creating supervision status: %w", err)
	}

	return nil
}

func (p *Processor) processClientReview(_ context.Context, supervisionRequest SupervisionRequest) error {
	log.Printf("Processing client review for supervision request %s", *supervisionRequest.Id)
	return nil

}
