package image

import (
	"github.com/kiry163/claw-pliers/internal/config"
)

var cfg *config.Config

func Init(imageCfg config.Config) error {
	cfg = &imageCfg
	return nil
}

func GetConfig() *config.Config {
	return cfg
}

type ImageResult struct {
	OutputPath string
	Size       int64
}

func Convert(inputPath, outputPath string, quality int) error {
	return nil
}

func Compress(inputPath, outputPath string, maxSize string, quality int) error {
	return nil
}

func Resize(inputPath, outputPath string, width, height string, fit string) error {
	return nil
}

func Rotate(inputPath, outputPath string, degrees int, flip, flop bool) error {
	return nil
}

func Watermark(inputPath, outputPath, logoPath, text string, opacity float64, fontSize int) error {
	return nil
}

func OCR(inputPath string) (string, error) {
	return "", nil
}

func Recognize(inputPath, prompt string) (string, error) {
	return "", nil
}

func Generate(prompt, outputPath, model, size string) error {
	return nil
}
