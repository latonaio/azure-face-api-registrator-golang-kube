# azure-face-api-registrator-golang-kube
## 概要  
1枚の画像を Azure Face API(Detect) にかけ、返り値として、画像に映っているすべての人物の顔の位置座標、性別・年齢等の情報を取得します。   
Azure Face API の仕様により、顔の位置座標を形成する長方形の面積が最も広い顔が先頭に来ます。    
この仕様を利用して、その先頭の顔の 位置座標を保持します。  
次に、この位置座標を使って、その位置座標の通り、当該画像を切り取ります。  
続いて、Azure Face API(Person Group _ Person - Create) に1つのレコードを登録します。  
そして、Azure Face API(Person Group _ Person - Add Face) に当該切り取られた画像を入力して、当該レコードに画像を更新します。        
最後に、Azure Face API の AI の学習内容に更新をかけるために、Azure Face API(Train) を実行します。 
  
参考：Azure Face API の Person Group は、Azure Face API ユーザ のインスタンス毎に独立した顔情報の維持管理の単位です。  

## azure-face-api-registrator-golang-kube を使用したエッジコンピューティングアーキテクチャの一例  
azure-face-api-registrator-golang-kube は、以下の黄色い枠の部分のリソースです。  

![フローチャート図](doc/omotebako_architecture_20211104.drawio.png)  

## 動作環境  

* OS: LinuxOS    
* CPU: ARM/AMD/Intel   
* Kubernetes    
* Golang Runtime

## 前提条件    
Azure Face API サービス に アクセスキー、エンドポイント、Person Group を登録します。  
登録されたエンドポイント、アクセスキー、Person Group を、本リポジトリ内の main.go に記載してください。  
```  
var apiKey = os.Getenv("API_ACCESS_KEY")
var endpoint = os.Getenv("API_ENDPOINT")
setPersonParam.PersonGroupID = os.Getenv("PERSON_GROUP_ID")
```  

## Azure Face API の Version 指定    
azure-face-api の version を指定します。  
本レポジトリの main.go では、下記のように記載されています。  
```
"github.com/Azure/azure-sdk-for-go/services/cognitiveservices/v1.0/face"
```
## Azure Face API(Detect) の テスト実行  
Azure Face API(Detect) の テスト実行 をするときは、本レポジトリのface下に任意の顔画像を配置してください。  
Azure FAce API 登録されているエンドポイントを、事前に学習させます。下記の手順で学習させることができます。  
```
# shellディレクトリ内のrecreate-group.shを実行します。シェル内のENDPOINT, SUBSCRIPTION_KEY, PERSON_GROUP_IDは使用するFaceAPIのエンドポイントに応じて書き換えて下さい。
$ bash recreate-group.sh
# 上記のコマンド実行するとPerson_idが出力されるので、train.shの3行目のPERSON_IDの値を置換しシェルを実行して下さい。
$ bash train.sh
```
* SQLにface_id_azure (TEXT), guest_id (INT) カラムを持つguestテーブルを作成しておきます。  
* `shell/setup-env.sh`　は、face-api-config.jsonと.envを作成するためのシェルスクリプトです。   



## Azure  の Version  
Azure Face API で使用する azure-sdk-for-go のバージョン指定は、go.mod の中にあります。  
```
	github.com/Azure/azure-sdk-for-go v59.3.0+incompatible
	github.com/Azure/go-autorest/autorest v0.11.22
```

## Recognition API の Version
Azure Face API で使用する Recognition API のバージョン指定は、関連ソースコードとともに、xxxxxx の中にあります。
```
RECOGNITION_MODEL="recognition_XX"
```

## I/O
#### Input-1
入力データ1のJSONフォーマットは、main.go にある通り、次の様式です。  
```
	// key:face_image_path
	// value:/var/lib/aion/Data/ui-backend-for-omotebako/1638003921862.jpg

	// key:output_data_path
	// value:/var/lib/aion/Data/ui-backend-for-omotebako

	// key:guest_id
	// value:1
```
1. 入力データのファイルパス(output_data_path)    
前工程のマイクロサービスから渡されたJSONメッセージファイルのパス          
2. 顧客ID(guest_id)      
(エッジ)アプリケーションの顧客ID       
3. 顔画像のパス(face_image_path)        
入力顔画像のパス  

#### Input-2
入力データ2として、Azure Face API(Detect)への入力は、Azure FaceClient を用いて、主として src/azure/getFaceCoordinate.go の次のソースコードにより行われます。  
本レポジトリの src/azure/getFaceCoordinate.go の例では、画像に映っているすべての人物の顔の位置座標(X軸/Y軸)を、Azure Face API から取得するという記述になっています。  

```
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
```
#### Input-3
入力データ3として、Azure Face API(Person Group _ Person - Create)、ならびに、Azure Face API(Person Group _ Person - Add Face)への入力は、Azure FaceClient を用いて、主として src/azure/getFaceCoordinate.go の次のソースコードにより行われます。  

```
func RegistorPersonImage(ctx context.Context, personGroupId string, personId *uuid.UUID, img *os.File, lflc face.PersonGroupPersonClient, ir model.ImageRectangle) {
	var targetFace = []int32{ir.Left, ir.Top, ir.Width, ir.Height}
	_, err := lflc.AddFaceFromStream(ctx, personGroupId, *personId, img, "", targetFace, face.Detection01)
	if err != nil {
		logging.Error("not work AddFaceFromStream")
		panic(err)
	}
	logging.Info("registor image to azure api")
}
```            
#### Input-4
入力データ4として、Azure Face API(Train)への入力は、Azure FaceClient を用いて、主として src/azure/getFaceCoordinate.go の次のソースコードにより行われます。  
```
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
```

#### Output-1  
出力データ1のJSONフォーマットは、src/azure/getFaceCoordinate_test.go にある通り、次の様式です。（一部抜粋）  
```
personGroupID := "mercury"
	var guestID float64 = 1

	var ir model.ImageRectangle

	ir.Height = 73
	ir.Left = 124
	ir.Top = 82
	ir.Width = 73
```

#### Output-2  
出力データ2のJSONフォーマットは、main.go にある通り、次の様式です。  
```
{
		"result":        result,
		"filepath":      setPersonParam.ImgPath,
		"guest_id":      setPersonParam.GuestID,
		"face_id_azure": personID,
	}
```
1. ゲストID(guest_id)    
(エッジ)アプリケーションの顧客ID        
2. 顔画像ファイルのパス(filepath)      
顔画像ファイルのパス      
3. AzureFaceID(face_id_azure)      
AzureFaceAPIのFaceID    


## Getting Started  
1. 下記コマンドでDockerイメージを作成します。  
```
make docker-build
```
2. aion-service-definitions/services.ymlに設定を記載し、AionCore経由でKubernetesコンテナを起動します。  
services.ymlへの記載例：   
```
  azure-face-api-registrator-python-kube:
    startup: yes
    always: yes
    scale: 1
    env:
      RABBITMQ_URL: amqp://guest:guest@rabbitmq:5672/xxxxxxxx
      QUEUE_ORIGIN: azure-face-api-registrator-python-kube-queue
      QUEUE_TO: register-face-to-guest-table-kube-queue
```
## Flowchart
![フローチャート図](doc/flowchart.png)

### goの拡張機能を入れる
VSCodeでコマンドパレット(Ctrl + Shift + P)を開いてGO:Install/Update Toolsで検索
すべてにチェックを入れてインストールする。

### 各種セーブ時に設定されるようにする
```
cd /home/your_work_directory/
mkdir .vscode
touch .vscode/setting.json
```

.vscode の設定
```json
{
    "editor.formatOnSave": true,
    "[go]": {
        "editor.tabSize": 4,
        "editor.insertSpaces": false,
        "editor.formatOnSave": false,
        "editor.formatOnPaste": false,
        "editor.formatOnType": false
      },
      "files.autoSave": "afterDelay",
      "files.autoSaveDelay": 1000,
      "go.formatTool": "goreturns",
      "go.formatFlags": [
        "-w"
      ],
      "go.lintTool": "golint",
      "go.lintOnSave": "package",
      "go.vetOnSave": "package",
      "go.buildOnSave": "package",
      "go.testOnSave": false,
      "go.gocodeAutoBuild": true,
      "go.installDependenciesWhenBuilding": true
}
```

### misc
[Faceクライアントライブラリ ドキュメント](https://docs.microsoft.com/ja-jp/azure/cognitive-services/face/quickstarts/client-libraries?tabs=visual-studio&pivots=programming-language-go)

