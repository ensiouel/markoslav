package service

import (
	"context"
	"github.com/fogleman/gg"
	"image"
	"markoslav/internal/model"
)

type ImageService interface {
	Draw(ctx context.Context, caption model.Caption, img image.Image) (image.Image, error)
}

type imageService struct {
	fontPath string
}

func NewImageService(fontPath string) ImageService {
	return &imageService{fontPath: fontPath}
}

func (service *imageService) Draw(_ context.Context, caption model.Caption, img image.Image) (image.Image, error) {
	c := gg.NewContextForImage(img)

	width := float64(c.Width())
	height := float64(c.Height())
	padding := 20.0

	if err := c.LoadFontFace(service.fontPath, width*0.08); err != nil {
		return nil, err
	}

	c.FontHeight()

	c.SetRGB(0, 0, 0)
	c.DrawStringWrapped(caption.Text,
		width/2, height-padding, 0.5, 1, width, 1.3, gg.AlignCenter,
	)

	c.SetRGB(1, 1, 1)
	c.DrawStringWrapped(caption.Text,
		width/2, height-padding-padding/6, 0.5, 1, width, 1.3, gg.AlignCenter,
	)

	return c.Image(), nil
}
