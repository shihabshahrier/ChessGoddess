// Queue implements a Redis-backed job queue for background processing.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type JobType string

const (
	JobTypeAnalysis JobType = "analysis"
	JobTypeSnapshot JobType = "snapshot"
	JobTypeAI       JobType = "ai_explanation"
)

type Job struct {
	ID        string                 `json:"id"`
	Type      JobType                `json:"type"`
	Payload   map[string]interface{} `json:"payload"`
	Priority  int                    `json:"priority"`
	CreatedAt int64                  `json:"created_at"`
}

type Queue struct {
	client *redis.Client
	ctx    context.Context
}

func NewQueue(redisURL string) (*Queue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Queue{
		client: client,
		ctx:    ctx,
	}, nil
}

func (q *Queue) Enqueue(job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	queueKey := fmt.Sprintf("queue:%s", job.Type)
	if err := q.client.LPush(q.ctx, queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue job: %w", err)
	}

	return nil
}

func (q *Queue) Dequeue(jobType JobType) (*Job, error) {
	queueKey := fmt.Sprintf("queue:%s", jobType)

	results, err := q.client.BRPop(q.ctx, 0, queueKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	if len(results) < 2 {
		return nil, fmt.Errorf("invalid dequeue result")
	}

	var job Job
	if err := json.Unmarshal([]byte(results[1]), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}

func (q *Queue) EnqueueAnalysis(sessionID, gameID string, depth int) error {
	job := &Job{
		Type: JobTypeAnalysis,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"game_id":    gameID,
			"depth":      depth,
		},
		Priority:  1,
		CreatedAt: time.Now().Unix(),
	}
	return q.Enqueue(job)
}

func (q *Queue) EnqueueSnapshot(sessionID, userID string) error {
	job := &Job{
		Type: JobTypeSnapshot,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		},
		Priority:  2,
		CreatedAt: time.Now().Unix(),
	}
	return q.Enqueue(job)
}

func (q *Queue) GetQueueLength(jobType JobType) (int64, error) {
	queueKey := fmt.Sprintf("queue:%s", jobType)
	return q.client.LLen(q.ctx, queueKey).Result()
}

func (q *Queue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *Queue) Close() error {
	return q.client.Close()
}
