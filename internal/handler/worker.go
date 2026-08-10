package handler

import (
	"context"
	"encoding/json"
	"pdf-box-aws/internal/domain"

	"github.com/aws/aws-lambda-go/events"
)

type Worker struct {
	repo FileRepository
}

func NewWorker(repo FileRepository) *Worker {
	return &Worker{
		repo: repo,
	}
}

func (w *Worker) HandleRequest(ctx context.Context, event events.SQSEvent) error {
	var failures []events.SQSBatchItemFailure
	for _, msg := range event.Records {
		var s3Event events.S3Event
		json.Unmarshal([]byte(msg.Body), &s3Event)
		for _, r := range s3Event.Records {
			r.S3.Object.Key
			r.S3.Object.Size
			r.EventName
			if err := w.repo.UpdateStatus(ctx, r.key, string(domain.StatusReady)); err != nil {
				failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: r.MessageId})
			}
		}
	}
	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}
