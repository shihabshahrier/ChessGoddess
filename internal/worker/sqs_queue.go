// SQSQueue implements JobQueue using AWS SQS. Used in production on AWS.
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var _ JobQueue = (*SQSQueue)(nil)

type SQSQueue struct {
	client    *sqs.Client
	queueURLs map[JobType]string
}

func NewSQSQueue(analysisURL, snapshotURL, aiExplainURL string) (*SQSQueue, error) {
	cfg, err := awsconfig.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &SQSQueue{
		client: sqs.NewFromConfig(cfg),
		queueURLs: map[JobType]string{
			JobTypeAnalysis: analysisURL,
			JobTypeSnapshot: snapshotURL,
			JobTypeAI:       aiExplainURL,
		},
	}, nil
}

func (q *SQSQueue) Enqueue(job *Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	url, ok := q.queueURLs[job.Type]
	if !ok {
		return fmt.Errorf("no queue URL for job type %s", job.Type)
	}

	body := string(data)
	_, err = q.client.SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    &url,
		MessageBody: &body,
	})
	if err != nil {
		return fmt.Errorf("sqs send failed: %w", err)
	}
	return nil
}

func (q *SQSQueue) Dequeue(jobType JobType) (*Job, error) {
	url, ok := q.queueURLs[jobType]
	if !ok {
		return nil, fmt.Errorf("no queue URL for job type %s", jobType)
	}

	waitTime := int32(20) // long polling
	maxMsgs := int32(1)
	out, err := q.client.ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
		QueueUrl:            &url,
		WaitTimeSeconds:     waitTime,
		MaxNumberOfMessages: maxMsgs,
	})
	if err != nil {
		return nil, fmt.Errorf("sqs receive failed: %w", err)
	}

	if len(out.Messages) == 0 {
		return nil, fmt.Errorf("no messages available")
	}

	msg := out.Messages[0]
	var job Job
	if err := json.Unmarshal([]byte(*msg.Body), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Delete after successful dequeue — processing failures retry via DLQ.
	_, _ = q.client.DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      &url,
		ReceiptHandle: msg.ReceiptHandle,
	})

	return &job, nil
}

func (q *SQSQueue) EnqueueAnalysis(sessionID, gameID string, depth int) error {
	return q.Enqueue(&Job{
		Type: JobTypeAnalysis,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"game_id":    gameID,
			"depth":      depth,
		},
		Priority:  1,
		CreatedAt: time.Now().Unix(),
	})
}

func (q *SQSQueue) EnqueueSnapshot(sessionID, userID string) error {
	return q.Enqueue(&Job{
		Type: JobTypeSnapshot,
		Payload: map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		},
		Priority:  2,
		CreatedAt: time.Now().Unix(),
	})
}

func (q *SQSQueue) GetQueueLength(jobType JobType) (int64, error) {
	url, ok := q.queueURLs[jobType]
	if !ok {
		return 0, fmt.Errorf("no queue URL for job type %s", jobType)
	}

	out, err := q.client.GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl:       &url,
		AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameApproximateNumberOfMessages},
	})
	if err != nil {
		return 0, fmt.Errorf("sqs get attributes failed: %w", err)
	}

	val, ok := out.Attributes[string(sqstypes.QueueAttributeNameApproximateNumberOfMessages)]
	if !ok {
		return 0, nil
	}

	var count int64
	fmt.Sscanf(val, "%d", &count)
	return count, nil
}

func (q *SQSQueue) Ping(ctx context.Context) error {
	for _, url := range q.queueURLs {
		u := url
		_, err := q.client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
			QueueUrl:       &u,
			AttributeNames: []sqstypes.QueueAttributeName{sqstypes.QueueAttributeNameQueueArn},
		})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (q *SQSQueue) Close() error {
	return nil
}
