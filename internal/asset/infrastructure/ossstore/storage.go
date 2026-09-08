package ossstore

import (
	"context"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/ksamwang/PersonalContentPlatform/internal/asset/ports"
	"io"
	"strconv"
	"time"
)

type Storage struct{ bucket *oss.Bucket }

func New(endpoint, bucketName, key, secret string) (*Storage, error) {
	client, err := oss.New(endpoint, key, secret)
	if err != nil {
		return nil, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, err
	}
	return &Storage{bucket: bucket}, nil
}
func (s *Storage) Put(_ context.Context, key string, body io.Reader, _ int64, mime string) error {
	return s.bucket.PutObject(key, body, oss.ContentType(mime))
}
func (s *Storage) Open(_ context.Context, key string) (io.ReadCloser, error) {
	return s.bucket.GetObject(key)
}
func (s *Storage) Stat(_ context.Context, key string) (ports.ObjectInfo, error) {
	headers, err := s.bucket.GetObjectMeta(key)
	if err != nil {
		return ports.ObjectInfo{}, err
	}
	size, err := strconv.ParseInt(headers.Get("Content-Length"), 10, 64)
	if err != nil {
		return ports.ObjectInfo{}, err
	}
	return ports.ObjectInfo{Size: size, ContentType: headers.Get("Content-Type")}, nil
}
func (s *Storage) PresignPut(_ context.Context, key, mime string, ttl time.Duration) (string, error) {
	return s.bucket.SignURL(key, oss.HTTPPut, int64(ttl.Seconds()), oss.ContentType(mime))
}
func (s *Storage) Check(_ context.Context) error {
	_, err := s.bucket.ListObjects(oss.MaxKeys(1))
	return err
}
