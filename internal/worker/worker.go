// Worker processes background jobs from the Redis queue (analysis, snapshots).
package worker

import (
	"context"
	"log/slog"
	"sync"

	"github.com/chessgoddess/chesslens/internal/repository"
	"github.com/chessgoddess/chesslens/internal/service"
)

type Worker struct {
	id           string
	queue        JobQueue
	analysisSvc  *service.AnalysisService
	sessionRepo  *repository.AnalysisSessionRepository
	gameRepo     *repository.GameRepository
	snapshotRepo *repository.SnapshotRepository
	concurrency  int
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

func New(
	id string,
	q JobQueue,
	analysisSvc *service.AnalysisService,
	sessionRepo *repository.AnalysisSessionRepository,
	gameRepo *repository.GameRepository,
	snapshotRepo *repository.SnapshotRepository,
	concurrency int,
) *Worker {
	return &Worker{
		id:           id,
		queue:        q,
		analysisSvc:  analysisSvc,
		sessionRepo:  sessionRepo,
		gameRepo:     gameRepo,
		snapshotRepo: snapshotRepo,
		concurrency:  concurrency,
		stopCh:       make(chan struct{}),
	}
}

func (w *Worker) Start() {
	slog.Info("worker starting", "id", w.id, "concurrency", w.concurrency)
	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(i)
	}
}

func (w *Worker) Stop() {
	slog.Info("worker stopping", "id", w.id)
	close(w.stopCh)
	w.wg.Wait()
	slog.Info("worker stopped", "id", w.id)
}

func (w *Worker) processLoop(workerID int) {
	defer w.wg.Done()

	jobTypes := []JobType{JobTypeAnalysis, JobTypeSnapshot, JobTypeAI}

	for {
		select {
		case <-w.stopCh:
			return
		default:
			if w.queue == nil {
				return
			}

			for _, jt := range jobTypes {
				job, err := w.queue.Dequeue(jt)
				if err != nil {
					continue
				}

				slog.Info("processing job", "worker", w.id, "worker_id", workerID, "type", jt, "job", job.ID)

				if err := w.handleJob(job); err != nil {
					slog.Error("job failed", "worker", w.id, "job", job.ID, "error", err)
				} else {
					slog.Info("job completed", "worker", w.id, "job", job.ID)
				}
			}
		}
	}
}

func (w *Worker) handleJob(job *Job) error {
	switch job.Type {
	case JobTypeAnalysis:
		return w.handleAnalysis(job)
	case JobTypeSnapshot:
		return w.handleSnapshot(job)
	default:
		return nil
	}
}

func (w *Worker) handleAnalysis(job *Job) error {
	sessionID, _ := job.Payload["session_id"].(string)
	gameID, _ := job.Payload["game_id"].(string)

	ctx := context.Background()

	session, err := w.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	game, err := w.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		return err
	}

	if err := w.sessionRepo.UpdateStatus(ctx, sessionID, "running"); err != nil {
		return err
	}

	if w.analysisSvc == nil {
		w.sessionRepo.UpdateStatus(ctx, sessionID, "failed")
		return nil
	}

	if err := w.analysisSvc.AnalyzeGame(ctx, session, game.PGN); err != nil {
		w.sessionRepo.UpdateStatus(ctx, sessionID, "failed")
		return err
	}

	return nil
}

func (w *Worker) handleSnapshot(job *Job) error {
	sessionID, _ := job.Payload["session_id"].(string)
	userID, _ := job.Payload["user_id"].(string)

	ctx := context.Background()

	if w.analysisSvc == nil {
		return nil
	}

	moves, err := w.analysisSvc.GetMovesBySessionID(ctx, sessionID)
	if err != nil {
		return err
	}

	session, err := w.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	snapshotData := map[string]interface{}{
		"session": session,
		"moves":   moves,
	}

	return w.snapshotRepo.Create(ctx, sessionID, userID, snapshotData)
}
