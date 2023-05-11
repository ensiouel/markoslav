package service

import (
	"context"
	"github.com/google/uuid"
	"markoslav/internal/dto"
	"markoslav/internal/model"
	"markoslav/internal/storage"
	"markoslav/pkg/apperror"
	"markoslav/pkg/filter"
	"time"
)

type CaptionService interface {
	Create(ctx context.Context, request dto.CreateCaption) (model.Caption, error)

	GetByID(ctx context.Context, captionID uuid.UUID) (model.Caption, error)
	GetRandom(ctx context.Context) (model.Caption, error)

	Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error)

	Update(ctx context.Context, request dto.UpdateCaption) error

	Delete(ctx context.Context, captionID uuid.UUID) error
}

type captionService struct {
	storage storage.CaptionStorage
}

func NewCaptionService(storage storage.CaptionStorage) CaptionService {
	return &captionService{storage: storage}
}

func (service *captionService) Create(ctx context.Context, request dto.CreateCaption) (model.Caption, error) {
	exists, err := service.storage.ExistsByText(ctx, request.Text)
	if err != nil {
		return model.Caption{}, err
	}

	if exists {
		return model.Caption{}, apperror.AlreadyExists.WithMessage("caption already exists")
	}

	caption := model.Caption{
		ID:        uuid.New(),
		Text:      request.Text,
		AuthorID:  request.AuthorID,
		Approved:  false,
		CreatedAt: time.Now(),
	}
	err = service.storage.Create(ctx, caption)
	if err != nil {
		return model.Caption{}, err
	}

	return caption, nil
}

func (service *captionService) GetByID(ctx context.Context, captionID uuid.UUID) (model.Caption, error) {
	caption, err := service.storage.GetByID(ctx, captionID)
	if err != nil {
		return model.Caption{}, err
	}

	return caption, nil
}

func (service *captionService) GetRandom(ctx context.Context) (model.Caption, error) {
	caption, err := service.storage.GetRandom(ctx)
	if err != nil {
		return model.Caption{}, err
	}

	return caption, nil
}

func (service *captionService) Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error) {
	captions, err := service.storage.Select(ctx, count, offset, options)
	if err != nil {
		return []model.Caption{}, err
	}

	return captions, nil
}

func (service *captionService) Update(ctx context.Context, request dto.UpdateCaption) error {
	caption, err := service.GetByID(ctx, request.ID)
	if err != nil {
		return err
	}

	caption.Approved = request.Approved

	err = service.storage.Update(ctx, caption)
	if err != nil {
		return err
	}

	return nil
}

func (service *captionService) Delete(ctx context.Context, captionID uuid.UUID) error {
	err := service.storage.Delete(ctx, captionID)
	if err != nil {
		return err
	}

	return nil
}
