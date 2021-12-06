package azure

import (
	"azure-face-api-registrator-golang-kube/src/model"
)

func SetPerson(setPersonParam model.SetPersonParam) (string, string) {

	// 画像から顔の位置を取得
	faceRectangle, err := GetFaceCoordinate(setPersonParam.Ctx, setPersonParam.Client, setPersonParam.Img)
	if err != nil {
		panic(err)
	}

	// 直前の処理で画像が終端まで読み込まれたので、再度最初から読み込みできるようにする
	setPersonParam.Img.Seek(0, 0)

	// ゲストIDを Azure Face API に登録、対応する personID を取得
	personID, err := RegistorFaceId(setPersonParam.Ctx, setPersonParam.Pgpc, setPersonParam.PersonGroupID, setPersonParam.GuestID)
	if err != nil {
		panic(err)
	}

	// 顔を人物に割り当てる
	RegistorPersonImage(setPersonParam.Ctx, setPersonParam.PersonGroupID, personID, setPersonParam.Img, setPersonParam.Pgpc, faceRectangle)

	// PersonGroup をトレーニングする
	result, err := TrainPersonImage(setPersonParam.Ctx, setPersonParam.Pgc, faceRectangle, setPersonParam.PersonGroupID)
	if err != nil {
		panic(err)
	}

	return result, personID.String()
}
