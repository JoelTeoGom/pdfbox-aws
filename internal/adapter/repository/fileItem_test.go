package repository

import (
	"testing"
	"time"

	"pdf-box-aws/internal/domain"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

var (
	createdAt      = time.Date(2026, 8, 20, 10, 0, 0, 0, time.UTC)
	trashedAt      = time.Date(2026, 8, 20, 12, 0, 0, 0, time.UTC)
	trashExpiresAt = trashedAt.Add(domain.DefaultTrashRetention)
)

func newFile(status domain.Status) *domain.File {
	return &domain.File{
		ID:        "01K3ZQF5W0000000000000000",
		OwnerID:   "user-1",
		S3Key:     domain.S3KeyFor("user-1", "01K3ZQF5W0000000000000000"),
		Filename:  "invoice.pdf",
		Size:      1024,
		Mime:      domain.MimeTypePDF,
		Status:    status,
		CreatedAt: createdAt,
	}
}

func TestRoundTripKeepsTrashFields(t *testing.T) {
	original := newFile(domain.StatusTrashed)
	original.TrashedAt = trashedAt
	original.TrashExpiresAt = trashExpiresAt

	restored := FromDomainFile(original).ToDomainFile()

	if restored.ID != original.ID || restored.OwnerID != original.OwnerID {
		t.Fatalf("identity lost: got %s/%s", restored.OwnerID, restored.ID)
	}
	if !restored.CreatedAt.Equal(original.CreatedAt) {
		t.Errorf("createdAt: got %s, want %s", restored.CreatedAt, original.CreatedAt)
	}
	if !restored.TrashedAt.Equal(trashedAt) {
		t.Errorf("trashedAt: got %s, want %s", restored.TrashedAt, trashedAt)
	}
	if !restored.TrashExpiresAt.Equal(trashExpiresAt) {
		t.Errorf("trashExpiresAt: got %s, want %s", restored.TrashExpiresAt, trashExpiresAt)
	}
	if restored.Status != domain.StatusTrashed {
		t.Errorf("status: got %s, want %s", restored.Status, domain.StatusTrashed)
	}
}

func TestRoundTripLeavesUnsetTimestampsZero(t *testing.T) {
	restored := FromDomainFile(newFile(domain.StatusUploaded)).ToDomainFile()

	if !restored.TrashedAt.IsZero() {
		t.Errorf("trashedAt should be zero, got %s", restored.TrashedAt)
	}
	if !restored.TrashExpiresAt.IsZero() {
		t.Errorf("trashExpiresAt should be zero, got %s", restored.TrashExpiresAt)
	}
}

func TestSweepIndexIsSparse(t *testing.T) {
	tests := []struct {
		name             string
		file             *domain.File
		wantInIndex      bool
		wantPartitionKey string
		wantDueAt        int64
	}{
		{
			name:             "pending is due from creation",
			file:             newFile(domain.StatusPending),
			wantInIndex:      true,
			wantPartitionKey: "SWEEP#pending",
			wantDueAt:        createdAt.Unix(),
		},
		{
			name:        "uploaded stays out of the index",
			file:        newFile(domain.StatusUploaded),
			wantInIndex: false,
		},
		{
			name: "trashed is due when its retention runs out",
			file: func() *domain.File {
				file := newFile(domain.StatusTrashed)
				file.TrashedAt = trashedAt
				file.TrashExpiresAt = trashExpiresAt
				return file
			}(),
			wantInIndex:      true,
			wantPartitionKey: "SWEEP#trashed",
			wantDueAt:        trashExpiresAt.Unix(),
		},
		{
			name:             "rejected waits for its object to be removed",
			file:             newFile(domain.StatusRejected),
			wantInIndex:      true,
			wantPartitionKey: "SWEEP#rejected",
			wantDueAt:        createdAt.Unix(),
		},
		{
			name:             "deleted without a confirmed object is retried",
			file:             newFile(domain.StatusDeleted),
			wantInIndex:      true,
			wantPartitionKey: "SWEEP#deleted",
			wantDueAt:        createdAt.Unix(),
		},
		{
			name: "deleted with the object gone leaves the index",
			file: func() *domain.File {
				file := newFile(domain.StatusDeleted)
				file.S3Deleted = true
				return file
			}(),
			wantInIndex: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := attributevalue.MarshalMap(FromDomainFile(tc.file))
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			_, hasPartitionKey := item["gsiPK"]
			_, hasSortKey := item["gsiSK"]

			if !tc.wantInIndex {
				if hasPartitionKey || hasSortKey {
					t.Fatalf("item should not be in the index, got gsiPK=%v gsiSK=%v", item["gsiPK"], item["gsiSK"])
				}
				return
			}

			if !hasPartitionKey || !hasSortKey {
				t.Fatalf("item should be in the index, got gsiPK=%v gsiSK=%v", item["gsiPK"], item["gsiSK"])
			}

			partitionKey, dueAt := sweepKeysFor(tc.file)
			if partitionKey != tc.wantPartitionKey {
				t.Errorf("partition key: got %s, want %s", partitionKey, tc.wantPartitionKey)
			}
			if dueAt != tc.wantDueAt {
				t.Errorf("due at: got %d, want %d", dueAt, tc.wantDueAt)
			}
		})
	}
}

func TestTrashedWithoutRetentionStillReachesTheIndex(t *testing.T) {
	file := newFile(domain.StatusTrashed)
	file.TrashedAt = trashedAt

	item, err := attributevalue.MarshalMap(FromDomainFile(file))
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if _, hasSortKey := item["gsiSK"]; !hasSortKey {
		t.Fatal("gsiSK missing, the item would be invisible to the sweeper")
	}
}
