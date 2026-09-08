package storagefactory

import (
	"context"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/filesystem"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/ossstore"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/infrastructure/s3store"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/ports"
	"github.com/ksamwang/PersonalContentPlatform/internal/platform/config"
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
