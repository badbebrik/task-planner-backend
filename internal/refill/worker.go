package refill

import (
	"context"
	"log"
	"task-planner/internal/schedule"
	"time"

	"task-planner/internal/goal"
)

type Worker struct {
	goalRepo  goal.RepositoryAggregator
	service   goal.Service
	scheduler schedule.Service
}

func NewWorker(repo goal.RepositoryAggregator, svc goal.Service, sch schedule.Service) *Worker {
	return &Worker{goalRepo: repo, service: svc, scheduler: sch}
}

func (w *Worker) Tick(ctx context.Context) {
	goals, err := w.goalRepo.ListActiveGoals(ctx)
	if err != nil {
		log.Printf("[Refill] list goals: %v", err)
		return
	}
	for _, g := range goals {
		added, err := w.service.AutoRefillTasks(ctx, g.ID)
		if err != nil {
			log.Printf("[Refill] goal %s: %v", g.ID, err)
		}
		if added > 0 {
			_, _ = w.scheduler.AutoScheduleForGoal(ctx, g.ID)
		}
		time.Sleep(2 * time.Second)
	}
}
