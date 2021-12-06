package azure

import (
	"azure-face-api-registrator-golang-kube/src/model"
	"context"
	"os"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
)

/* 前処理として、person-groupの初期化をする必要があるため face-detect-script を使用してください
*  実行コマンド
*  make init-person-group
*  make person-list
 */
func TestGetFaceCoordinate(t *testing.T) {
	apiKey := "xxxx"
	endpoint := "xxxx"
	faceContext := context.Background()

	client := face.NewClient(endpoint)
	client.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)

	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	ir, _ := GetFaceCoordinate(faceContext, client, image)

	if ir.Height != 73 {
		t.Errorf("incorrect height %v", ir.Height)
	}
	if ir.Left != 124 {
		t.Errorf("incorrect Left %v", ir.Left)
	}
	if ir.Top != 82 {
		t.Errorf("incorrect Top %v", ir.Top)
	}
	if ir.Width != 73 {
		t.Errorf("incorrect Top %v", ir.Width)
	}
}

func TestRegistorFaceId(t *testing.T) {
	apiKey := "xxxx"
	endpoint := "xxxx"
	faceContext := context.Background()

	personGroupPersonClient := face.NewPersonGroupPersonClient(endpoint)
	personGroupPersonClient.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)

	personGroupID := "mercury"
	var guestID float64 = 1

	var ir model.ImageRectangle

	ir.Height = 73
	ir.Left = 124
	ir.Top = 82
	ir.Width = 73

	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	personID, err := RegistorFaceId(faceContext, personGroupPersonClient, personGroupID, guestID)
	if err != nil {
		t.Errorf("something settings wrong")
	}

	t.Logf("faceid: %v", personID.String())

	RegistorPersonImage(faceContext, personGroupID, personID, image, personGroupPersonClient, ir)

}

/*
* 結合テスト
 */
func TestSetPerson(t *testing.T) {
	apiKey := "xxxx"
	endpoint := "xxxx"

	var setPersonParam model.SetPersonParam
	setPersonParam.Ctx = context.Background()
	setPersonParam.Client = face.NewClient(endpoint)
	setPersonParam.Client.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgc = face.NewPersonGroupClient(endpoint)
	setPersonParam.Pgc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgpc = face.NewPersonGroupPersonClient(endpoint)
	setPersonParam.Pgpc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.PersonGroupID = "mercury"

	image, err := os.Open("../../face/initial.jpeg")
	if err != nil {
		panic(err)
	}
	defer image.Close()

	setPersonParam.Img = image

	SetPerson(setPersonParam)
}
