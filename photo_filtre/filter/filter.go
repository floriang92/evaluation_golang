package filter

import (
	"github.com/disintegration/imaging"
)

type Filter interface {
	Process(srcPath, dstPath string) error
}

type Grayscale struct{}

type Blur struct{}

func (f *Grayscale) Process(srcPath, destPath string) error {
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}
	grayscaleImage := imaging.Grayscale(srcImage)
	err = imaging.Save(grayscaleImage, destPath)
	if err != nil {
		return err
	}

	return nil
}

func (f *Blur) Process(srcPath, destPath string) error {
	srcImage, err := imaging.Open(srcPath)
	if err != nil {
		return err
	}
	blurredImage := imaging.Blur(srcImage, 5.0)

	err = imaging.Save(blurredImage, destPath)
	if err != nil {
		return err
	}
	return nil
}
