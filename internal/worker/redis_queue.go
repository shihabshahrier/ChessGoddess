// RedisQueue implements JobQueue using Redis LPUSH/BRPOP. Used for local dev.
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

var _ JobQueue = (*RedisQueue)(nil)

type RedisQueue struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisQueue(redisURL string) (*RedisQueue, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}

	client := redis.NewClient(opts)

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &RedisQueue{
		client: client,
		ctx:    ctx,
	}, nil
}

func (q *RedisQueue) Enqueue(job *Job) error {
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

func (q *RedisQueue) Dequeue(jobType JobType) (*Job, error) {
	queueKey := fmt.Sprintf("queue:%s", jobType)

	results, err := q.client.BRPop(q.ctx, 5*time.Second, queueKey).Result()
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

func (q *RedisQueue) EnqueueAnalysis(sessionID, gameID string, depth int) error {
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

func (q *RedisQueue) EnqueueSnapshot(sessionID, userID string) error {
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

func (q *RedisQueue) GetQueueLength(jobType JobType) (int64, error) {
	queueKey := fmt.Sprintf("queue:%s", jobType)
	return q.client.LLen(q.ctx, queueKey).Result()
}

func (q *RedisQueue) Ping(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}

func (q *RedisQueue) Close() error {
	return q.client.Close()
}
