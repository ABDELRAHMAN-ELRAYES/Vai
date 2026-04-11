package jobs

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
)

// Job can be scheduled.
type Job interface {
	Name() string
	Run(ctx context.Context) error
}

type jobEntry struct {
	job      Job
	interval time.Duration
}

// Manages background jobs
type Scheduler struct {
	logger *zap.SugaredLogger
	jobs   []jobEntry
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewScheduler(logger *zap.SugaredLogger) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Adds a job to the scheduler
func (scheduler *Scheduler) Register(job Job, interval time.Duration) {
	scheduler.jobs = append(scheduler.jobs, jobEntry{job: job, interval: interval})
}

// Run all jobs
func (scheduler *Scheduler) Start() {
	scheduler.logger.Infow("Starting background job scheduler", "job_count", len(scheduler.jobs))

	for _, entry := range scheduler.jobs {
		scheduler.wg.Add(1)
		go scheduler.runJob(entry)
	}
}

// Stop all background jobs
func (scheduler *Scheduler) Stop() {
	scheduler.logger.Info("Stopping background job scheduler...")
	scheduler.cancel()
	scheduler.wg.Wait()
	scheduler.logger.Info("Background job scheduler stopped.")
}

func (scheduler *Scheduler) runJob(entry jobEntry) {
	defer scheduler.wg.Done()

	ticker := time.NewTicker(entry.interval)
	defer ticker.Stop()

	scheduler.logger.Infow("Job started", "job", entry.job.Name(), "interval", entry.interval.String())

	if err := entry.job.Run(scheduler.ctx); err != nil {
		scheduler.logger.Errorw("Job failed", "job", entry.job.Name(), "error", err)
	}

	for {
		select {
		case <-ticker.C:
			scheduler.logger.Debugw("Running scheduled job", "job", entry.job.Name())
			if err := entry.job.Run(scheduler.ctx); err != nil {
				scheduler.logger.Errorw("Job failed", "job", entry.job.Name(), "error", err)
			}
		case <-scheduler.ctx.Done():
			scheduler.logger.Infow("Job stopped", "job", entry.job.Name())
			return
		}
	}
}
