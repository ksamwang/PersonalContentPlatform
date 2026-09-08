package storagefactory

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/filesystem"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/ossstore"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/s3store"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
	"github.com/ksamwang/PersonalContentPlatform/internal/settings/domain"
	settingsports "github.com/ksamwang/PersonalContentPlatform/internal/settings/ports"
)

func New(ctx context.Context, c config.Config) (ports.Storage, error) {
	switch c.StorageProvider {
	case "filesystem":
		return filesystem.New(c.StorageBasePath)
	case "s3", "r2":
		return s3store.New(ctx, c.StorageEndpoint, c.StorageRegion, c.StorageBucket, c.StorageAccessKey, c.StorageSecretKey)
	case "oss":
		return ossstore.New(c.StorageEndpoint, c.StorageBucket, c.StorageAccessKey, c.StorageSecretKey)
	default:
		return nil, fmt.Errorf("unsupported OBJECT_STORAGE_PROVIDER %q", c.StorageProvider)
	}
}

type Resolver struct {
	repository settingsports.Repository
	fallback   ports.Storage
}

func NewResolver(repository settingsports.Repository, fallback ports.Storage) *Resolver {
	return &Resolver{repository: repository, fallback: fallback}
}
func (r *Resolver) Resolve(ctx context.Context, workspaceID uuid.UUID, profileID *uuid.UUID) (ports.Storage, *uuid.UUID, error) {
	var profile *domain.StorageProfile
	var err error
	if profileID != nil {
		profile, err = r.repository.StorageByID(ctx, workspaceID, *profileID)
	} else {
		profile, err = r.repository.ActiveStorage(ctx, workspaceID)
	}
	if err != nil || profile == nil {
		if profileID != nil {
			return nil, nil, fmt.Errorf("stored object provider is unavailable")
		}
		return r.fallback, nil, nil
	}
	storage, err := NewProfile(ctx, *profile)
	return storage, &profile.ID, err
}
func NewProfile(ctx context.Context, p domain.StorageProfile) (ports.Storage, error) {
	switch p.Provider {
	case "filesystem":
		return filesystem.New(p.BasePath)
	case "s3", "r2":
		return s3store.New(ctx, p.Endpoint, p.Region, p.Bucket, p.AccessKey, p.SecretKey)
	case "oss":
		return ossstore.New(p.Endpoint, p.Bucket, p.AccessKey, p.SecretKey)
	default:
		return nil, fmt.Errorf("unsupported storage provider %q", p.Provider)
	}
}
