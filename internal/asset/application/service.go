package application

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/id"
	"io"
	"mime"
	"path/filepath"
	"strings"
	"time"
)

const MaxUploadSize int64 = 100 << 20

type Service struct {
	repo    ports.Repository
	storage ports.StorageResolver
}

func New(repo ports.Repository, storage ports.StorageResolver) *Service {
	return &Service{repo: repo, storage: storage}
}
func (s *Service) Prepare(ctx context.Context, workspaceID, userID uuid.UUID, filename, mimeType string, size int64) (domain.UploadPlan, error) {
	filename = filepath.Base(strings.TrimSpace(filename))
	if filename == "." || filename == "" || size < 1 || size > MaxUploadSize {
		return domain.UploadPlan{}, fmt.Errorf("invalid file or size")
	}
	if mimeType == "" {
		mimeType = mime.TypeByExtension(filepath.Ext(filename))
	}
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	intent := domain.UploadIntent{ID: id.New(), WorkspaceID: workspaceID, StorageKey: fmt.Sprintf("original/%s/%s", workspaceID, id.New()), Filename: filename, MIME: mimeType, ExpectedSize: size, ExpiresAt: time.Now().Add(15 * time.Minute)}
	storage, profileID, err := s.storage.Resolve(ctx, workspaceID, nil)
	if err != nil {
		return domain.UploadPlan{}, err
	}
	intent.StorageProfileID = profileID
	if err := s.repo.CreateIntent(ctx, intent, userID); err != nil {
		return domain.UploadPlan{}, err
	}
	url, err := storage.PresignPut(ctx, intent.StorageKey, mimeType, 15*time.Minute)
	if err != nil {
		return domain.UploadPlan{}, err
	}
	if url == "" {
		url = fmt.Sprintf("/v1/workspaces/%s/assets/uploads/%s", workspaceID, intent.ID)
	}
	return domain.UploadPlan{UploadID: intent.ID, Method: "PUT", URL: url, Headers: map[string]string{"Content-Type": mimeType}, ExpiresAt: intent.ExpiresAt}, nil
}
func (s *Service) Upload(ctx context.Context, workspaceID, intentID uuid.UUID, body io.Reader, size int64) error {
	intent, err := s.repo.Intent(ctx, workspaceID, intentID)
	if err != nil {
		return err
	}
	if size != intent.ExpectedSize {
		return fmt.Errorf("upload size does not match intent")
	}
	storage, _, err := s.storage.Resolve(ctx, workspaceID, intent.StorageProfileID)
	if err != nil {
		return err
	}
	return storage.Put(ctx, intent.StorageKey, body, size, intent.MIME)
}
func (s *Service) Complete(ctx context.Context, workspaceID, userID, intentID uuid.UUID) (domain.Asset, error) {
	intent, err := s.repo.Intent(ctx, workspaceID, intentID)
	if err != nil {
		return domain.Asset{}, err
	}
	storage, _, err := s.storage.Resolve(ctx, workspaceID, intent.StorageProfileID)
	if err != nil {
		return domain.Asset{}, err
	}
	info, err := storage.Stat(ctx, intent.StorageKey)
	if err != nil {
		return domain.Asset{}, err
	}
	if info.Size != intent.ExpectedSize {
		return domain.Asset{}, fmt.Errorf("stored object size mismatch")
	}
	reader, err := storage.Open(ctx, intent.StorageKey)
	if err != nil {
		return domain.Asset{}, err
	}
	defer reader.Close()
	hash := sha256.New()
	if _, err = io.Copy(hash, reader); err != nil {
		return domain.Asset{}, err
	}
	return s.repo.Complete(ctx, intent, userID, hex.EncodeToString(hash.Sum(nil)), info.Size)
}
func (s *Service) List(ctx context.Context, workspaceID uuid.UUID, limit int) ([]domain.Asset, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	return s.repo.List(ctx, workspaceID, limit)
}
func (s *Service) Get(ctx context.Context, workspaceID, assetID uuid.UUID) (domain.Asset, error) {
	return s.repo.Get(ctx, workspaceID, assetID)
}
func (s *Service) Usages(ctx context.Context, workspaceID, assetID uuid.UUID) ([]domain.Usage, error) {
	return s.repo.Usages(ctx, workspaceID, assetID)
}
func (s *Service) Archive(ctx context.Context, workspaceID, assetID uuid.UUID) error {
	return s.repo.Archive(ctx, workspaceID, assetID)
}
func (s *Service) Replace(ctx context.Context, workspaceID, assetID, replacementID uuid.UUID) (domain.Asset, error) {
	if assetID == replacementID {
		return domain.Asset{}, fmt.Errorf("replacement asset must be different")
	}
	return s.repo.Replace(ctx, workspaceID, assetID, replacementID)
}
func (s *Service) OpenPublic(ctx context.Context, assetID uuid.UUID) (io.ReadCloser, domain.StoredObject, error) {
	object, err := s.repo.Object(ctx, assetID)
	if err != nil {
		return nil, domain.StoredObject{}, err
	}
	storage, _, err := s.storage.Resolve(ctx, object.WorkspaceID, object.StorageProfileID)
	if err != nil {
		return nil, domain.StoredObject{}, err
	}
	reader, err := storage.Open(ctx, object.StorageKey)
	return reader, object, err
}
