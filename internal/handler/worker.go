package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"pdf-box-aws/internal/domain"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

type Worker struct {
	repo    FileRepository
	storage S3service
}

func NewWorker(repo FileRepository, storage S3service) *Worker {
	return &Worker{
		repo:    repo,
		storage: storage,
	}
}

var errPermanent = errors.New("permanent failure, do not retry")

func (w *Worker) HandleRequest(ctx context.Context, event events.SQSEvent) (events.SQSEventResponse, error) {
	var failures []events.SQSBatchItemFailure

	for _, msg := range event.Records {
		if err := w.processMessage(ctx, msg); err != nil {
			if errors.Is(err, errPermanent) {
				slog.ErrorContext(ctx, "dropping message", "messageId", msg.MessageId, "error", err)
				continue // delete from queue, do not retry
			}
			slog.ErrorContext(ctx, "retryable failure", "messageId", msg.MessageId, "error", err)
			failures = append(failures, events.SQSBatchItemFailure{ItemIdentifier: msg.MessageId})
		}
	}

	return events.SQSEventResponse{BatchItemFailures: failures}, nil
}

func (w *Worker) processMessage(ctx context.Context, msg events.SQSMessage) error {
	var s3Event events.S3Event
	if err := json.Unmarshal([]byte(msg.Body), &s3Event); err != nil {
		return fmt.Errorf("%w: unmarshal body: %v", errPermanent, err)
	}

	attempts, _ := strconv.Atoi(msg.Attributes["ApproximateReceiveCount"])

	for _, r := range s3Event.Records {
		if !strings.HasPrefix(r.EventName, "ObjectCreated:") {
			continue
		}

		userID, fileID, err := domain.ParseS3Key(r.S3.Object.Key)
		if err != nil {
			return fmt.Errorf("%w: bad key %q: %v", errPermanent, r.S3.Object.Key, err)
		}

		file, err := w.repo.Get(ctx, userID, fileID)
		if errors.Is(err, domain.ErrNotFound) {
			if attempts < 3 {
				return fmt.Errorf("record not visible yet, retrying")
			}
			if err := w.storage.DeleteFile(ctx, r.S3.Object.Key); err != nil {
				slog.ErrorContext(ctx, "orphan cleanup failed", "key", r.S3.Object.Key)
			}
			return fmt.Errorf("%w: no record for %s", errPermanent, r.S3.Object.Key)
		}
		if err != nil {
			return fmt.Errorf("dynamo get: %w", err)
		}

		if r.S3.Object.Size > domain.MaxFileSize || file.Mime != domain.MimeTypePDF || file.Size != r.S3.Object.Size {
			_ = w.repo.UpdateStatus(ctx, userID, fileID, domain.StatusRejected)
			if err := w.storage.DeleteFile(ctx, r.S3.Object.Key); err != nil {
				slog.ErrorContext(ctx, "failed to delete oversized object", "key", r.S3.Object.Key)
			}
			return nil
		}

		err = w.repo.UpdateStatus(ctx, userID, fileID, domain.StatusUploaded)
		if errors.Is(err, domain.ErrInvalidStatus) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("mark ready: %w", err)
		}
	}

	return nil
}
