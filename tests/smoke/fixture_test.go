package smoke_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamwang/PersonalContentPlatform/internal/app"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/database"
	publication "github.com/ksamwang/PersonalContentPlatform/internal/publication/application"
)

type smokeFixture struct {
	client      apiClient
	db          *pgxpool.Pool
	workspaceID uuid.UUID
	userID      uuid.UUID
	workspace   string
	objectRoot  string
	blobID      uuid.UUID
}

func newSmokeFixture(t *testing.T) *smokeFixture {
	t.Helper()
	if os.Getenv("PCP_INTEGRATION_TEST") != "1" {
		t.Skip("set PCP_INTEGRATION_TEST=1 to run database smoke tests")
	}
	objectRoot := t.TempDir()
	t.Setenv("OBJECT_STORAGE_PROVIDER", "filesystem")
	t.Setenv("OBJECT_STORAGE_BASE_PATH", objectRoot)
	t.Setenv("AI_BASE_URL", "")
	t.Setenv("AI_API_KEY", "")
	t.Setenv("AI_MODEL", "")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	db, err := database.Open(context.Background(), cfg.DatabaseURL, 4)
	if err != nil {
		t.Fatal(err)
	}
	workspaceID, userID := uuid.New(), uuid.New()
	workspaceSlug := "smoke-" + workspaceID.String()
	if _, err = db.Exec(context.Background(), `INSERT INTO workspaces(id,slug,name) VALUES($1,$2,'API Smoke Test')`, workspaceID, workspaceSlug); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err = db.Exec(context.Background(), `INSERT INTO users(id,email,display_name,status) VALUES($1,$2,'Smoke Test','active')`, userID, fmt.Sprintf("smoke-%s@example.invalid", userID)); err != nil {
		db.Close()
		t.Fatal(err)
	}
	if _, err = db.Exec(context.Background(), `INSERT INTO memberships(workspace_id,user_id,role) VALUES($1,$2,'owner')`, workspaceID, userID); err != nil {
		db.Close()
		t.Fatal(err)
	}
	rawToken := make([]byte, 32)
	if _, err = rand.Read(rawToken); err != nil {
		t.Fatal(err)
	}
	tokenHash := sha256.Sum256(rawToken)
	if _, err = db.Exec(context.Background(), `INSERT INTO sessions(id,user_id,token_hash,expires_at) VALUES($1,$2,$3,$4)`, uuid.New(), userID, tokenHash[:], time.Now().Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	application, err := app.New(db, cfg)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(application.Router())
	workerContext, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan error, 1)
	go func() { workerDone <- publication.NewProcessor(db, "smoke-test").Run(workerContext) }()
	fixture := &smokeFixture{
		client: apiClient{t: t, base: server.URL, cookie: &http.Cookie{
			Name: cfg.SessionCookieName, Value: base64.RawURLEncoding.EncodeToString(rawToken),
		}},
		db: db, workspaceID: workspaceID, userID: userID, workspace: workspaceSlug, objectRoot: objectRoot,
	}
	t.Cleanup(func() {
		stopWorker()
		<-workerDone
		server.Close()
		cleanupSmokeData(t, db, workspaceID, userID, fixture.blobID)
		db.Close()
	})
	return fixture
}

func cleanupSmokeData(t *testing.T, db *pgxpool.Pool, workspaceID, userID, blobID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		t.Errorf("begin smoke cleanup: %v", err)
		return
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Publications do not cascade from their target or content records, so they
	// must be removed before deleting the workspace-owned object graph.
	if _, err = tx.Exec(ctx, `DELETE FROM publications WHERE workspace_id=$1`, workspaceID); err != nil {
		t.Errorf("clean publications: %v", err)
		return
	}
	if _, err = tx.Exec(ctx, `DELETE FROM workspaces WHERE id=$1`, workspaceID); err != nil {
		t.Errorf("clean workspace: %v", err)
		return
	}
	if _, err = tx.Exec(ctx, `DELETE FROM users WHERE id=$1`, userID); err != nil {
		t.Errorf("clean user: %v", err)
		return
	}
	if blobID != uuid.Nil {
		if _, err = tx.Exec(ctx, `
			DELETE FROM blobs
			WHERE id=$1
			  AND NOT EXISTS (SELECT 1 FROM assets WHERE blob_id=$1)
			  AND NOT EXISTS (SELECT 1 FROM asset_variants WHERE blob_id=$1)
		`, blobID); err != nil {
			t.Errorf("clean blob: %v", err)
			return
		}
	}
	if err = tx.Commit(ctx); err != nil {
		t.Errorf("commit smoke cleanup: %v", err)
	}
}
