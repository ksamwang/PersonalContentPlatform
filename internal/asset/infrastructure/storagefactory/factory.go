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
	"path/filepath"
	"strings"
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
	repository     settingsports.Repository
	fallback       ports.Storage
	filesystemRoot string
}

func NewResolver(repository settingsports.Repository, fallback ports.Storage, filesystemRoot string) *Resolver {
	return &Resolver{repository: repository, fallback: fallback, filesystemRoot: filesystemRoot}
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
	storage, err := NewProfileWithRoot(ctx, *profile, r.filesystemRoot)
	return storage, &profile.ID, err
}
func NewProfileWithRoot(ctx context.Context, p domain.StorageProfile, root string) (ports.Storage, error) {
	if p.Provider != "filesystem" {
		return NewProfile(ctx, p)
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	configured := strings.TrimSpace(p.BasePath)
	if configured == "" || configured == "." {
		configured = rootAbs
	} else {
		candidate, absErr := filepath.Abs(configured)
		if absErr != nil {
			return nil, absErr
		}
		if filepath.Clean(candidate) == filepath.Clean(rootAbs) {
			configured = rootAbs
		} else if !filepath.IsAbs(p.BasePath) {
			configured = filepath.Join(rootAbs, p.BasePath)
		} else {
			configured = candidate
		}
	}
	resolved, err := filepath.Abs(configured)
	if err != nil {
		return nil, err
	}
	relative, err := filepath.Rel(rootAbs, resolved)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, fmt.Errorf("filesystem path must stay within the configured storage root")
	}
	return filesystem.New(resolved)
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
