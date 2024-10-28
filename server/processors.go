package sentinel

import (
	"context"
	"fmt"
	"log"
	"time"
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
	if supervisorId == nil {
		return fmt.Errorf("supervisor ID is required but was not provided to processReview")
	}

	supervisor, err := p.store.GetSupervisor(ctx, *supervisorId)
	if err != nil {
		return fmt.Errorf("error getting supervisor: %w", err)
	}

	switch supervisor.Type {
	case HumanSupervisor:
		return p.processHumanReview(ctx, supervisionRequest)
	case ClientSupervisor:
		return nil
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
