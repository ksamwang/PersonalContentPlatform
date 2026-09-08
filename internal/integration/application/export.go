package application

import (
	"context"

	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/domain"
	"github.com/ksamwang/PersonalContentPlatform/internal/integration/ports"
)

type ExportService struct {
	repository ports.Repository
}

func NewExportService(repository ports.Repository) *ExportService {
	return &ExportService{repository: repository}
}

func (s *ExportService) Build(ctx context.Context, workspaceID, actorID uuid.UUID) (domain.Manifest, error) {
	runID, err := s.repository.StartExport(ctx, workspaceID, actorID)
	if err != nil {
		return domain.Manifest{}, err
	}
	manifest, buildErr := s.repository.BuildManifest(ctx, workspaceID)
	if completeErr := s.repository.CompleteExport(ctx, runID, manifest, buildErr); completeErr != nil && buildErr == nil {
		return domain.Manifest{}, completeErr
	}
	return manifest, buildErr
}
