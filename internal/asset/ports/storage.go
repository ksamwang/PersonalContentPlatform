package ports

import (
	"context"
	"io"
	"time"
)

type ObjectInfo struct {
	Size        int64
	ContentType string
}
type Storage interface {
	Put(context.Context, string, io.Reader, int64, string) error
	Open(context.Context, string) (io.ReadCloser, error)
	Stat(context.Context, string) (ObjectInfo, error)
	PresignPut(context.Context, string, string, time.Duration) (string, error)
}
