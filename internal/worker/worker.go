package worker

import (
	"context"
	"log"
	"sync"

	"github.com/chessgoddess/chesslens/internal/analysis"
	"github.com/chessgoddess/chesslens/internal/queue"
	"github.com/chessgoddess/chesslens/internal/repository"
)

type Worker struct {
	id            string
	queue         *queue.Queue
	analysisSvc   *analysis.Service
	sessionRepo   *repository.AnalysisSessionRepository
	gameRepo      *repository.GameRepository
	snapshotRepo  *repository.SnapshotRepository
	concurrency   int
	stopCh        chan struct{}
	wg            sync.WaitGroup
}

func New(id string, q *queue.Queue, analysisSvc *analysis.Service, sessionRepo *repository.AnalysisSessionRepository, gameRepo *repository.GameRepository, snapshotRepo *repository.SnapshotRepository, concurrency int) *Worker {
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
	log.Printf("Worker %s starting with %d concurrency", w.id, w.concurrency)

	for i := 0; i < w.concurrency; i++ {
		w.wg.Add(1)
		go w.processLoop(i)
	}
}

func (w *Worker) Stop() {
	log.Printf("Worker %s stopping", w.id)
	close(w.stopCh)
	w.wg.Wait()
	log.Printf("Worker %s stopped", w.id)
}

func (w *Worker) processLoop(workerID int) {
	defer w.wg.Done()

	for {
		select {
		case <-w.stopCh:
			return
		default:
			job, err := w.queue.Dequeue(queue.JobTypeAnalysis, 5)
			if err != nil {
				continue
			}

			log.Printf("Worker %s-%d processing job %s", w.id, workerID, job.ID)

			if err := w.handleJob(job); err != nil {
				log.Printf("Worker %s-%d failed job %s: %v", w.id, workerID, job.ID, err)
			} else {
				log.Printf("Worker %s-%d completed job %s", w.id, workerID, job.ID)
			}
		}
	}
}

func (w *Worker) handleJob(job *queue.Job) error {
	switch job.Type {
	case queue.JobTypeAnalysis:
		return w.handleAnalysis(job)
	case queue.JobTypeSnapshot:
		return w.handleSnapshot(job)
	default:
		return nil
	}
}

func (w *Worker) handleAnalysis(job *queue.Job) error {
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

	if err := w.analysisSvc.AnalyzeGame(ctx, session, game.PGN); err != nil {
		w.sessionRepo.UpdateStatus(ctx, sessionID, "failed")
		return err
	}

	return nil
}

func (w *Worker) handleSnapshot(job *queue.Job) error {
	sessionID, _ := job.Payload["session_id"].(string)
	userID, _ := job.Payload["user_id"].(string)

	ctx := context.Background()

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

	if err := w.snapshotRepo.Create(ctx, sessionID, userID, snapshotData); err != nil {
		return err
	}

	return nil
}
