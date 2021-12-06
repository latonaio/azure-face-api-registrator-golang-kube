package main

import (
	"azure-face-api-registrator-golang-kube/src/azure"
	"azure-face-api-registrator-golang-kube/src/model"
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
	"github.com/Azure/go-autorest/autorest"
	"github.com/latonaio/golang-logging-library/logger"
	rabbitmq "github.com/latonaio/rabbitmq-golang-client"
)

// Azure API に必要な鍵とエンドポイントの設定
var apiKey = os.Getenv("API_ACCESS_KEY")
var endpoint = os.Getenv("API_ENDPOINT")
var rabbitmqURL = os.Getenv("RABBITMQ_URL")
var queueOrigin = os.Getenv("QUEUE_ORIGIN")
var queueTo = os.Getenv("QUEUE_TO")
var logging = logger.NewLogger()

func main() {

	var setPersonParam model.SetPersonParam
	setPersonParam.Ctx = context.Background()
	setPersonParam.Client = face.NewClient(endpoint)
	setPersonParam.Client.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgc = face.NewPersonGroupClient(endpoint)
	setPersonParam.Pgc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.Pgpc = face.NewPersonGroupPersonClient(endpoint)
	setPersonParam.Pgpc.Authorizer = autorest.NewCognitiveServicesAuthorizer(apiKey)
	setPersonParam.PersonGroupID = os.Getenv("PERSON_GROUP_ID")

	// rabbitmq へ接続
	rabbitmqClient, err := rabbitmq.NewRabbitmqClient(
		rabbitmqURL,
		[]string{queueOrigin},
		[]string{queueTo},
	)
	if err != nil {
		logging.Error("can't connect rabbitmq")
		return
	}

	defer rabbitmqClient.Close()

	iter, err := rabbitmqClient.Iterator()
	if err != nil {
		logging.Error("not working iterator")
		return
	}

	logging.Info("start azure-face-api registrator")

	for message := range iter {

		registorImage(message, setPersonParam, rabbitmqClient)
	}

}

func registorImage(message rabbitmq.RabbitmqMessage, setPersonParam model.SetPersonParam, rabbitmqClient *rabbitmq.RabbitmqClient) {
	// 渡されるデータの例
	// key:face_image_path
	// value:/var/lib/aion/Data/ui-backend-for-omotebako/1638003921862.jpg

	// key:output_data_path
	// value:/var/lib/aion/Data/ui-backend-for-omotebako

	// key:guest_id
	// value:1

	for index, value := range message.Data() {
		if index == "guest_id" {
			setPersonParam.GuestID = value.(float64)
		}
		if index == "face_image_path" {
			setPersonParam.ImgPath = value.(string)
		}
	}

	// 検出対象となる顔画像を開く
	img, err := os.Open(setPersonParam.ImgPath)
	if err != nil {
		panic(err)
	}
	logging.Info("open image")
	defer img.Close()

	setPersonParam.Img = img

	// 画像を Azure API に登録
	result, personID := azure.SetPerson(setPersonParam)
	message.Success()

	payload := map[string]interface{}{
		"result":        result,
		"filepath":      setPersonParam.ImgPath,
		"guest_id":      setPersonParam.GuestID,
		"face_id_azure": personID,
	}
	if err := rabbitmqClient.Send(queueTo, payload); err != nil {
		log.Printf("error: %v", err)
	}
	logging.Info("send to %v", queueTo)

}
