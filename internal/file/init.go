package file

import (
	"context"
	"io"

	"github.com/kiry163/claw-pliers/internal/config"
	"github.com/kiry163/claw-pliers/internal/database"
)

var (
	Database    *database.DB
	FileStorage Storage
)

type StubStorage struct{}

func (s *StubStorage) Save(ctx context.Context, reader io.Reader, size int64, fileID, originalName string) (SaveResult, error) {
	return SaveResult{ObjectKey: fileID, Size: size, MimeType: "application/octet-stream"}, nil
}

type stubReader struct{}

func (r *stubReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (r *stubReader) Close() error {
	return nil
}

func (s *StubStorage) Get(ctx context.Context, objectKey string, rangeStart, rangeEnd *int64) (io.ReadCloser, ObjectInfo, error) {
	return &stubReader{}, ObjectInfo{}, nil
}

func (s *StubStorage) Stat(ctx context.Context, objectKey string) (ObjectInfo, error) {
	return ObjectInfo{}, nil
}

func (s *StubStorage) Delete(ctx context.Context, objectKey string) error {
	return nil
}

func Init(cfg config.Config) error {
	var err error

	Database, err = database.Open(database.Config{Path: cfg.Database.Path})
	if err != nil {
		return err
	}

	storage, err := NewMinioStorage(context.Background(), cfg.Minio)
	if err != nil {
		FileStorage = &StubStorage{}
	} else {
		FileStorage = storage
	}

	return nil
}
