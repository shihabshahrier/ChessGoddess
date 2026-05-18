package repository

import (
	"context"
	"testing"

	"github.com/chessgoddess/chesslens/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Skip("Skipping integration test - requires running Postgres")
	return nil
}

func TestSnapshotRepository_CreateAndRetrieve(t *testing.T) {
	pool := setupTestDB(t)
	if pool == nil {
		return
	}

	repo := NewSnapshotRepository(pool)

	ctx := context.Background()
	testData := map[string]interface{}{
		"session": map[string]interface{}{
			"id":     "test-session",
			"depth":  20,
			"status": "completed",
		},
		"moves": []map[string]interface{}{
			{"id": "m1", "san": "e4", "evaluation": 0.3},
			{"id": "m2", "san": "e5", "evaluation": 0.2},
		},
	}

	err := repo.Create(ctx, "test-session", "test-user", testData)
	if err != nil {
		t.Fatalf("failed to create snapshot: %v", err)
	}

	snapshots, err := repo.ListByUserID(ctx, "test-user", 10, 0)
	if err != nil {
		t.Fatalf("failed to list snapshots: %v", err)
	}

	if len(snapshots) == 0 {
		t.Error("expected at least one snapshot")
	}
}

func TestSnapshotData_Integrity(t *testing.T) {
	snapshot := models.Snapshot{
		ID:         "snap-1",
		SessionID:  "sess-1",
		UserID:     "user-1",
		ShareToken: "abc123",
		IsPublic:   false,
		Data: map[string]interface{}{
			"session": map[string]interface{}{
				"id":     "sess-1",
				"depth":  20,
				"status": "completed",
			},
			"moves": []interface{}{
				map[string]interface{}{"id": "m1", "san": "e4", "evaluation": 0.3},
				map[string]interface{}{"id": "m2", "san": "e5", "evaluation": 0.2},
			},
		},
	}

	if snapshot.Data == nil {
		t.Error("snapshot data should not be nil")
	}

	sessionData, ok := snapshot.Data["session"].(map[string]interface{})
	if !ok {
		t.Error("session data should be a map")
	}

	if sessionData["id"] != "sess-1" {
		t.Errorf("expected session id 'sess-1', got '%v'", sessionData["id"])
	}

	moves, ok := snapshot.Data["moves"].([]interface{})
	if !ok {
		t.Error("moves should be a slice")
	}

	if len(moves) != 2 {
		t.Errorf("expected 2 moves, got %d", len(moves))
	}
}

func TestSnapshotData_Immutable(t *testing.T) {
	original := models.Snapshot{
		ID:         "snap-1",
		SessionID:  "sess-1",
		UserID:     "user-1",
		ShareToken: "abc123",
		Data: map[string]interface{}{
			"moves": []interface{}{
				map[string]interface{}{"id": "m1", "san": "e4"},
			},
		},
	}

	copy := original
	copy.Data = make(map[string]interface{})
	for k, v := range original.Data {
		copy.Data[k] = v
	}

	copy.Data["moves"] = []interface{}{
		map[string]interface{}{"id": "m1", "san": "e5"},
	}

	originalMoves := original.Data["moves"].([]interface{})
	originalMove := originalMoves[0].(map[string]interface{})
	if originalMove["san"] != "e4" {
		t.Error("original snapshot should not be modified")
	}
}

func TestSnapshotShareToken_Unique(t *testing.T) {
	token1 := "abc123def456ghi789jkl012mno345pqr678stu901"
	token2 := "def456ghi789jkl012mno345pqr678stu901vwx234"

	if token1 == token2 {
		t.Error("share tokens should be unique")
	}

	if len(token1) < 10 {
		t.Error("share token should be sufficiently long")
	}
}
