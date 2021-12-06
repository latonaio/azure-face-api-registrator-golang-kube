package openimage

import (
	"os"

	"github.com/latonaio/golang-logging-library/logger"
)

func OpenImage(imagePath string) *os.File {
	logging := logger.NewLogger()
	img, err := os.Open(imagePath)
	if err != nil {
		panic(err)
	}
	logging.Info("open image")

	defer img.Close()
	return img
}
