package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/domain"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
)

type imageRecipe struct {
	name     string
	maxWidth int
}

var imageRecipes = []imageRecipe{{"thumbnail", 480}, {"content-1280", 1280}}

func (s *Service) deriveImages(ctx context.Context, intent domain.UploadIntent, asset domain.Asset) error {
	if asset.MediaType != "image" {
		return nil
	}
	storage, profileID, err := s.storage.Resolve(ctx, asset.WorkspaceID, intent.StorageProfileID)
	if err != nil {
		return err
	}
	reader, err := storage.Open(ctx, intent.StorageKey)
	if err != nil {
		return err
	}
	source, _, err := image.Decode(reader)
	_ = reader.Close()
	if err != nil {
		return nil
	}
	bounds := source.Bounds()
	originalWidth, originalHeight := bounds.Dx(), bounds.Dy()
	for _, recipe := range imageRecipes {
		width := originalWidth
		if width > recipe.maxWidth {
			width = recipe.maxWidth
		}
		height := originalHeight * width / originalWidth
		if height < 1 {
			height = 1
		}
		target := image.NewRGBA(image.Rect(0, 0, width, height))
		draw.CatmullRom.Scale(target, target.Bounds(), source, bounds, draw.Over, nil)
		var output bytes.Buffer
		if err = jpeg.Encode(&output, target, &jpeg.Options{Quality: 86}); err != nil {
			return err
		}
		key := fmt.Sprintf("variants/%s/%s/%s.jpg", asset.WorkspaceID, asset.ID, recipe.name)
		if err = storage.Put(ctx, key, bytes.NewReader(output.Bytes()), int64(output.Len()), "image/jpeg"); err != nil {
			return err
		}
		hash := sha256.Sum256(output.Bytes())
		if err = s.repo.SaveVariant(ctx, asset.WorkspaceID, asset.ID, profileID, recipe.name, key, hex.EncodeToString(hash[:]), int64(output.Len()), width, height, originalWidth, originalHeight); err != nil {
			return err
		}
	}
	return nil
}
