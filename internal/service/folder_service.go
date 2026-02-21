package service

import (
	"context"
	"time"

	"github.com/kiry163/claw-pliers/internal/database"
	"github.com/kiry163/claw-pliers/internal/logger"
	"github.com/kiry163/claw-pliers/internal/utils"

	"github.com/rs/zerolog"
)

type FolderService struct {
	db     *database.DB
	logger *zerolog.Logger
}

func NewFolderService(db *database.DB) *FolderService {
	l := logger.Get()
	return &FolderService{
		db:     db,
		logger: l,
	}
}

type FolderMetadata struct {
	FolderID  string
	Name      string
	ParentID  *string
	CreatedBy string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *FolderService) CreateFolder(ctx context.Context, name, parentID, createdBy string) (FolderMetadata, error) {
	folderID := utils.GenerateFolderID()
	record := &database.Folder{
		FolderID:  folderID,
		Name:      name,
		ParentID:  nil,
		CreatedBy: createdBy,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if parentID != "" {
		record.ParentID = &parentID
	}

	if err := s.db.CreateFolder(record); err != nil {
		s.logger.Error().Err(err).Str("name", name).Msg("failed to create folder")
		return FolderMetadata{}, err
	}

	s.logger.Info().
		Str("folder_id", folderID).
		Str("name", name).
		Msg("folder created successfully")

	return FolderMetadata{
		FolderID:  record.FolderID,
		Name:      record.Name,
		ParentID:  record.ParentID,
		CreatedBy: record.CreatedBy,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (s *FolderService) GetFolder(ctx context.Context, folderID string) (FolderMetadata, error) {
	record, err := s.db.GetFolder(folderID)
	if err != nil {
		s.logger.Error().Err(err).Str("folder_id", folderID).Msg("failed to get folder")
		return FolderMetadata{}, err
	}

	return FolderMetadata{
		FolderID:  record.FolderID,
		Name:      record.Name,
		ParentID:  record.ParentID,
		CreatedBy: record.CreatedBy,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (s *FolderService) ListFolders(ctx context.Context, parentID *string) ([]FolderMetadata, error) {
	records, err := s.db.ListFolders(parentID)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list folders")
		return nil, err
	}

	items := make([]FolderMetadata, 0, len(records))
	for _, r := range records {
		items = append(items, FolderMetadata{
			FolderID:  r.FolderID,
			Name:      r.Name,
			ParentID:  r.ParentID,
			CreatedBy: r.CreatedBy,
			CreatedAt: r.CreatedAt,
		})
	}

	return items, nil
}

func (s *FolderService) DeleteFolder(ctx context.Context, folderID string) error {
	if err := s.db.DeleteFolder(folderID); err != nil {
		s.logger.Error().Err(err).Str("folder_id", folderID).Msg("failed to delete folder")
		return err
	}

	s.logger.Info().Str("folder_id", folderID).Msg("folder deleted successfully")
	return nil
}

func (s *FolderService) RenameFolder(ctx context.Context, folderID, newName string) error {
	if err := s.db.UpdateFolder(folderID, newName); err != nil {
		s.logger.Error().Err(err).Str("folder_id", folderID).Str("new_name", newName).Msg("failed to rename folder")
		return err
	}

	s.logger.Info().Str("folder_id", folderID).Str("new_name", newName).Msg("folder renamed successfully")
	return nil
}

func (s *FolderService) MoveFolder(ctx context.Context, folderID string, parentID *string) error {
	if err := s.db.MoveFolder(folderID, parentID); err != nil {
		s.logger.Error().Err(err).Str("folder_id", folderID).Msg("failed to move folder")
		return err
	}

	s.logger.Info().Str("folder_id", folderID).Msg("folder moved successfully")
	return nil
}

func (s *FolderService) GetFolderByPath(ctx context.Context, path string) (FolderMetadata, error) {
	record, err := s.db.GetFolderByPath(path)
	if err != nil {
		s.logger.Error().Err(err).Str("path", path).Msg("failed to get folder by path")
		return FolderMetadata{}, err
	}

	return FolderMetadata{
		FolderID:  record.FolderID,
		Name:      record.Name,
		ParentID:  record.ParentID,
		CreatedBy: record.CreatedBy,
		CreatedAt: record.CreatedAt,
		UpdatedAt: record.UpdatedAt,
	}, nil
}

func (s *FolderService) GetFolderPath(ctx context.Context, folderID string) (string, error) {
	return s.db.GetFolderPath(folderID)
}

func (s *FolderService) GetFolderItemCount(ctx context.Context, folderID string) (int64, int64, error) {
	return s.db.GetFolderItemCount(folderID)
}

func (s *FolderService) GenerateFolderID() string {
	return utils.GenerateFolderID()
}
