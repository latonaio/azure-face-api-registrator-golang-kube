package model

import (
	"context"
	"image"
	"os"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/cognitiveservices/face"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

type ImageRectangle struct {
	Left   int32
	Top    int32
	Width  int32
	Height int32
}

type SetPersonParam struct {
	Ctx           context.Context
	Client        face.Client                  // 顔の検出、類似の検索、および検証の例に使用されるクライアント
	Pgpc          face.PersonGroupPersonClient // PersonGroupにPersonを追加する際に使用されるクライアント
	Pgc           face.PersonGroupClient       // PersonGroupに使用されるクライアント
	Img           *os.File
	ImgPath       string
	PersonGroupID string
	GuestID       float64
}
