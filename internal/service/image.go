package service

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/webp"
)

type ServiceImage interface {
	GetDimensions(io.Reader) (int, int, error)
}

type serviceImage struct {
}

func NewServiceImage() ServiceImage {
	return &serviceImage{}
}

func (s *serviceImage) GetDimensions(r io.Reader) (int, int, error) {
	imageConfig, _, err := image.DecodeConfig(r)
	if err != nil {
		return 0, 0, fmt.Errorf("sImage - failed to decode image: %w", err)
	}

	return imageConfig.Width, imageConfig.Height, nil
}
