package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Image processing commands",
}

var imageConvertCmd = &cobra.Command{
	Use:   "convert <input> <output>",
	Short: "Convert image format",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Image convert - This is a stub")
		return nil
	},
}

var imageOCRCommand = &cobra.Command{
	Use:   "ocr <image>",
	Short: "OCR text recognition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Image OCR - This is a stub")
		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageConvertCmd)
	imageCmd.AddCommand(imageOCRCommand)
}
