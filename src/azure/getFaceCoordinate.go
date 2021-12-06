package azure

import (
	"azure-face-api-registrator-golang-kube/src/model"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/gofrs/uuid"
	"github.com/latonaio/golang-logging-library/logger"
)

var logging = logger.NewLogger()

/*
* 画像から顔の位置を取得
 */
func GetFaceCoordinate(ctx context.Context, client face.Client, image *os.File) (model.ImageRectangle, error) {
	var imageRectangle model.ImageRectangle
	var err error

	// DetectWithStreamから取得するパラメータを設定
	attributes := []face.AttributeType{}
	returnFaceID := true
	returnRecognitionModel := false
	returnFaceLandmarks := false

	detectSingleFaces, dErr := client.DetectWithStream(ctx, image, &returnFaceID, &returnFaceLandmarks, attributes, face.Recognition01, &returnRecognitionModel, face.Detection01)
	if dErr != nil {
		logging.Error("failed getting rectangle")
		err = dErr
	}

	// Dereference *[]DetectedFace, in order to loop through it.
	dFaces := *detectSingleFaces.Value

	imageRectangle.Top = *dFaces[0].FaceRectangle.Top
	imageRectangle.Height = *dFaces[0].FaceRectangle.Height
	imageRectangle.Left = *dFaces[0].FaceRectangle.Left
	imageRectangle.Width = *dFaces[0].FaceRectangle.Width

	logging.Info("succeed getting rectangle")

	return imageRectangle, err

}

/*
* ゲストIDをAzure Face API に登録
* ゲストIDに割り当てられたIDを返却
 */
func RegistorFaceId(ctx context.Context, lflc face.PersonGroupPersonClient, personGroupID string, guestID float64) (*uuid.UUID, error) {
	var body face.NameAndUserDataContract

	str := strconv.FormatFloat(guestID, 'f', -1, 64)
	body.Name = &str

	person, err := lflc.Create(ctx, personGroupID, body)
	if err != nil {
		logging.Error("can't create person")
	}

	logging.Info("registor faceid: %v", person.PersonID.String())
	return person.PersonID, err
}

/*
* 顔画像をAzure Face API に登録
 */
func RegistorPersonImage(ctx context.Context, personGroupId string, personId *uuid.UUID, img *os.File, lflc face.PersonGroupPersonClient, ir model.ImageRectangle) {
	var targetFace = []int32{ir.Left, ir.Top, ir.Width, ir.Height}
	_, err := lflc.AddFaceFromStream(ctx, personGroupId, *personId, img, "", targetFace, face.Detection01)
	if err != nil {
		logging.Error("not work AddFaceFromStream")
		panic(err)
	}
	logging.Info("registor image to azure api")
}

/*
* Azure に顔画像をトレーニングしてもらう
 */
func TrainPersonImage(ctx context.Context, lflc face.PersonGroupClient, ir model.ImageRectangle, personGroupID string) (result string, err error) {
	lflc.Train(ctx, personGroupID)

	var trainState string

	for {
		isTrainStatus, _ := lflc.GetTrainingStatus(ctx, personGroupID)

		if isTrainStatus.Status == face.TrainingStatusTypeSucceeded {
			trainState = "success"
			break
		}
		if isTrainStatus.Status == face.TrainingStatusTypeFailed {
			trainState = "failed"
			break
		}

	}

	if trainState == "failed" {
		logging.Error("failed to train")
		err = fmt.Errorf("failed to train")
		return trainState, err
	}

	return trainState, err
}
