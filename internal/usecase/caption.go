package usecase

import (
	"context"
	"github.com/google/uuid"
	"image"
	"markoslav/internal/dto"
	"markoslav/internal/model"
	"markoslav/internal/service"
	"markoslav/pkg/filter"
)

type CaptionUsecase interface {
	Create(ctx context.Context, request dto.CreateCaption) (model.Caption, error)

	Approve(ctx context.Context, captionID uuid.UUID) error
	Reject(ctx context.Context, captionID uuid.UUID) error

	Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error)

	DrawRandom(ctx context.Context, img image.Image) (image.Image, error)
}

type captionUsecase struct {
	captionService service.CaptionService
	imageService   service.ImageService
}

func NewCaptionUsecase(captionService service.CaptionService, imageService service.ImageService) CaptionUsecase {
	return &captionUsecase{captionService: captionService, imageService: imageService}
}

func (usecase *captionUsecase) Create(ctx context.Context, request dto.CreateCaption) (model.Caption, error) {
	return usecase.captionService.Create(ctx, request)
}

func (usecase *captionUsecase) Approve(ctx context.Context, captionID uuid.UUID) error {
	err := usecase.captionService.Update(ctx, dto.UpdateCaption{
		ID:       captionID,
		Approved: true,
	})
	if err != nil {
		return err
	}

	return nil
}

func (usecase *captionUsecase) Reject(ctx context.Context, captionID uuid.UUID) error {
	err := usecase.captionService.Delete(ctx, captionID)
	if err != nil {
		return err
	}

	return nil
}

func (usecase *captionUsecase) Select(ctx context.Context, count int, offset int, options filter.Options) ([]model.Caption, error) {
	return usecase.captionService.Select(ctx, count, offset, options)
}

func (usecase *captionUsecase) DrawRandom(ctx context.Context, img image.Image) (image.Image, error) {
	caption, err := usecase.captionService.GetRandom(ctx)
	if err != nil {
		return nil, err
	}

	return usecase.imageService.Draw(ctx, caption, img)
}
