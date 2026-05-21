package worker

import "context"

// JobQueue abstracts the queue transport. Redis for local dev, SQS for AWS.
type JobQueue interface {
	Enqueue(job *Job) error
	Dequeue(jobType JobType) (*Job, error)
	EnqueueAnalysis(sessionID, gameID string, depth int) error
	EnqueueSnapshot(sessionID, userID string) error
	GetQueueLength(jobType JobType) (int64, error)
	Ping(ctx context.Context) error
	Close() error
}
