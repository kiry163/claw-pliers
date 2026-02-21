package service

import (
	"context"
	"io"
	"time"

	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/file"
	"github.com/kiry163/claw-pliers/internal/logger"
	"github.com/kiry163/claw-pliers/internal/utils"

	"github.com/rs/zerolog"
)

type FileService struct {
	db      *database.DB
	storage file.Storage
	logger  *zerolog.Logger
}

func NewFileService(db *database.DB, storage file.Storage) *FileService {
	l := logger.Get()
	return &FileService{
		db:      db,
		storage: storage,
		logger:  l,
	}
}

type FileMetadata struct {
	FileID       string
	OriginalName string
	ObjectKey    string
	Size         int64
	MimeType     string
	FolderID     *string
	CreatedBy    string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ListFilesResult struct {
	Total int64
	Items []FileMetadata
}

func (s *FileService) CreateFile(ctx context.Context, reader io.Reader, size int64, fileID, originalName, folderID, createdBy string) (FileMetadata, error) {
	saveResult, err := s.storage.Save(ctx, reader, size, fileID, originalName)
	if err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to save file to storage")
		return FileMetadata{}, err
	}

	record := &database.File{
		FileID:       fileID,
		OriginalName: originalName,
		ObjectKey:    saveResult.ObjectKey,
		Size:         size,
		MimeType:     saveResult.MimeType,
		CreatedBy:    createdBy,
		CreatedAt:    time.Now().UTC(),
		UpdatedAt:    time.Now().UTC(),
	}

	if folderID != "" {
		record.FolderID = &folderID
	}

	if err := s.db.CreateFile(record); err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to create file record")
		return FileMetadata{}, err
	}

	s.logger.Info().
		Str("file_id", fileID).
		Str("original_name", originalName).
		Int64("size", size).
		Msg("file created successfully")

	return FileMetadata{
		FileID:       record.FileID,
		OriginalName: record.OriginalName,
		ObjectKey:    record.ObjectKey,
		Size:         record.Size,
		MimeType:     record.MimeType,
		FolderID:     record.FolderID,
		CreatedBy:    record.CreatedBy,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}, nil
}

func (s *FileService) GetFile(ctx context.Context, fileID string) (FileMetadata, error) {
	record, err := s.db.GetFile(fileID)
	if err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to get file")
		return FileMetadata{}, err
	}

	return FileMetadata{
		FileID:       record.FileID,
		OriginalName: record.OriginalName,
		ObjectKey:    record.ObjectKey,
		Size:         record.Size,
		MimeType:     record.MimeType,
		FolderID:     record.FolderID,
		CreatedBy:    record.CreatedBy,
		CreatedAt:    record.CreatedAt,
		UpdatedAt:    record.UpdatedAt,
	}, nil
}

func (s *FileService) ListFiles(ctx context.Context, folderID *string, limit, offset int, order, keyword string) (ListFilesResult, error) {
	var records []database.File
	var total int64
	var err error

	if folderID != nil || folderID == nil {
		records, total, err = s.db.ListFilesByFolder(folderID, limit, offset, order, keyword)
	} else {
		records, total, err = s.db.ListFiles(limit, offset, order, keyword)
	}

	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list files")
		return ListFilesResult{}, err
	}

	items := make([]FileMetadata, 0, len(records))
	for _, r := range records {
		items = append(items, FileMetadata{
			FileID:       r.FileID,
			OriginalName: r.OriginalName,
			Size:         r.Size,
			MimeType:     r.MimeType,
			CreatedAt:    r.CreatedAt,
		})
	}

	return ListFilesResult{
		Total: total,
		Items: items,
	}, nil
}

func (s *FileService) DeleteFile(ctx context.Context, fileID string) error {
	record, err := s.db.DeleteFile(fileID)
	if err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to delete file record")
		return err
	}

	if err := s.storage.Delete(ctx, record.ObjectKey); err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to delete file from storage")
		return err
	}

	s.logger.Info().Str("file_id", fileID).Msg("file deleted successfully")
	return nil
}

func (s *FileService) GetFileContent(ctx context.Context, fileID string) (io.ReadCloser, FileMetadata, error) {
	record, err := s.db.GetFile(fileID)
	if err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to get file")
		return nil, FileMetadata{}, err
	}

	reader, _, err := s.storage.Get(ctx, record.ObjectKey, nil, nil)
	if err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to get file content")
		return nil, FileMetadata{}, err
	}

	metadata := FileMetadata{
		FileID:       record.FileID,
		OriginalName: record.OriginalName,
		Size:         record.Size,
		MimeType:     record.MimeType,
	}

	return reader, metadata, nil
}

func (s *FileService) GenerateFileID() string {
	return utils.GenerateFileID()
}

func (s *FileService) GenerateShareToken() string {
	return utils.GenerateShareToken()
}

func (s *FileService) CreateShareLink(ctx context.Context, fileID, createdBy, publicURL string) (string, string, error) {
	now := time.Now().UTC()
	expiresAt := now.Add(7 * 24 * time.Hour)
	token := s.GenerateShareToken()

	shareLink := &database.ShareLink{
		Token:     token,
		FileID:    fileID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
		CreatedBy: createdBy,
		Status:    "active",
	}

	if err := s.db.CreateShareLink(shareLink); err != nil {
		s.logger.Error().Err(err).Str("file_id", fileID).Msg("failed to create share link")
		return "", "", err
	}

	downloadLink := publicURL + "/s/" + token
	s.logger.Info().Str("file_id", fileID).Str("token", token).Msg("share link created")
	return token, downloadLink, nil
}

func (s *FileService) GetShareLink(ctx context.Context, token string) (database.ShareLink, error) {
	return s.db.GetShareLink(token)
}

func (s *FileService) GetActiveShareLink(ctx context.Context, fileID string, now time.Time) (database.ShareLink, error) {
	return s.db.GetActiveShareLink(fileID, now)
}
