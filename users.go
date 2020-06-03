package main

import (
	"context"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"go.mongodb.org/mongo-driver/bson"
)

//GetIndex 使用者的dashboard介面
func GetIndex(ctx iris.Context) {
	session := sessions.Get(ctx)

	filter := bson.M{
		"uid": sessions.Get(ctx).Get("_id"),
	}

	result, err := APIQueryBase("characters", filter)

	if err != nil {
		session.SetFlash("msg", err)
	}

	ctx.ViewData("charAmount", len(result))

	if err := ctx.View("users/dashboard.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetEditUser 設定使用者的設定介面
func GetEditUser(ctx iris.Context) {
	if err := ctx.View("users/editUser.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PostUpdateUser 更新使用者設定
func PostUpdateUser(ctx iris.Context) {

}

//GetCharList 使用者角色列表
func GetCharList(ctx iris.Context) {
	session := sessions.Get(ctx)

	filter := bson.M{
		"uid": session.Get("_id"), //用使用者的_id去找characters collection的uid
	}

	result, err := APIQueryBase("characters", filter)

	if err != nil {
		// log.Fatal(err)
		session.SetFlash("msg", err)
	} else {
		if len(result) == 0 {
			ctx.ViewData("message", "目前您未有任何角色，請開始新增角色")
		} else {
			ctx.ViewData("characters", result)
			ctx.ViewData("raceList", map[int]string{1: "人類", 2: "精靈", 3: "巨人"})
			ctx.ViewData("weekbornList", map[int]string{
				7: "立春(Imbolic) 星期日",
				1: "春分(Alban Eiler) 星期一",
				2: "入夏(Beltane) 星期二",
				3: "立夏(Alban Heruin) 星期三",
				4: "秋收(Lughnasadh) 星期四",
				5: "秋收節(Alban Elved) 星期五",
				6: "山夏(Samhain) 星期六"})

			ctx.ViewData("serverList", AdminDB.servers)
		}
	}
	ctx.ViewData("message", session.GetFlashString("msg"))
	if err := ctx.View("users/charList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetNewChar 新增角色介面
func GetNewChar(ctx iris.Context) {
	ctx.ViewData("template", map[string]string{
		"banner": "新增",
		"method": "POST",
		"action": "/user/newchar",
		"button": "新增",
	})
	if err := ctx.View("users/CharForm.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PostNewChar 新增角色
func PostNewChar(ctx iris.Context) {
	session := sessions.Get(ctx)

	coll := DBSource.db.Collection("characters")
	/*
	   inputCharname
	   inputBirthday
	   inputWeekborn
	   inputRace
	   inputServer
	*/
	inputWeekborn, _ := strconv.ParseInt(ctx.PostValue("inputWeekborn"), 10, 32)
	inputRace, _ := strconv.ParseInt(ctx.PostValue("inputRace"), 10, 32)
	inputServer, _ := strconv.ParseInt(ctx.PostValue("inputServer"), 10, 32)
	insertData := bson.M{
		"uid":              session.Get("_id"),
		"characterId":      0,
		"name":             ctx.PostValue("inputCharname"),
		"birthday":         ctx.PostValue("inputBirthday"),
		"weekborn":         inputWeekborn,
		"race":             inputRace,
		"server":           inputServer,
		"imageUrl":         "",
		"enabled":          true,
		"shared":           false,
		"create_timestamp": primitive.Timestamp{T: uint32(time.Now().Unix())},
		"modify_timestamp": primitive.Timestamp{T: uint32(time.Now().Unix())},
	}

	//err := coll.FindOne(context.TODO(), filter).Decode(&result)
	insertResult, err := coll.InsertOne(context.TODO(), insertData)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Added a new character with objectID: ", insertResult.InsertedID)

	ctx.Redirect("/user/char")
}

//GetChar 瀏覽角色內容
//分成兩種進入方式
//1)/user/#### 此為已編入資料庫編號之角色(characterId)
//2)/user/u/123456789012345678901234 此為尚未編入資料庫編號之角色
func GetChar(ctx iris.Context) {
	// session := sessions.Get(ctx)

	// coll := DBSource.db.Collection("characters")

	// var result, query bson.M
	// if strings.HasPrefix(ctx.Path(), "/char/u/") {
	// 	id, _ := primitive.ObjectIDFromHex(ctx.Params().Get("uid"))
	// 	query = bson.M{
	// 		"uid": session.Get("_id"),
	// 		"_id": id, //未編入角色(24碼)
	// 	}
	// } else {
	// 	charid, _ := strconv.ParseInt(ctx.Params().Get("uid"), 10, 32)
	// 	query = bson.M{
	// 		"uid":         session.Get("_id"),
	// 		"characterId": charid,
	// 	}
	// }

	ctx.View("users/Char.html")

}

//GetEditChar 編輯角色基本資料
func GetEditChar(ctx iris.Context) {
	session := sessions.Get(ctx)

	var filter bson.M
	if strings.HasPrefix(ctx.Path(), "/char/u/") {
		id, _ := primitive.ObjectIDFromHex(ctx.Params().Get("uid"))
		filter = bson.M{
			"uid": session.Get("_id"),
			"_id": id, //未編入角色(24碼)
		}
		ctx.ViewData("template", map[string]string{
			"banner": "編輯",
			"method": "PUT",
			"action": "/char/u/" + ctx.Params().Get("uid"),
			"button": "更新",
		})
	} else {
		charid, _ := strconv.ParseInt(ctx.Params().Get("uid"), 10, 32)
		filter = bson.M{
			"uid":         session.Get("_id"),
			"characterId": charid,
		}
		ctx.ViewData("template", map[string]string{
			"banner": "編輯",
			"method": "PUT",
			"action": "/char/" + ctx.Params().Get("uid"),
			"button": "更新",
		})
	}

	result, err := APIQueryOneBase("characters", filter)

	if err != nil {
		session.SetFlash("msg", err)
		ctx.Redirect("/user/char")
	} else {
		ctx.ViewData("charData", result)
		ctx.ViewData("message", session.GetFlashString("msg")) //Show some message with error
		if err := ctx.View("users/CharForm.html"); err != nil {
			ctx.Application().Logger().Infof(err.Error())
		}
	}
}

//PutCharUpdate 更新資料庫角色
func PutCharUpdate(ctx iris.Context) {
	session := sessions.Get(ctx)

	/*
	   inputCharname
	   inputBirthday
	   inputWeekborn
	   inputRace
	   inputServer
	*/
	var filter bson.M
	inputWeekborn, _ := strconv.ParseInt(ctx.PostValue("inputWeekborn"), 10, 32)
	inputRace, _ := strconv.ParseInt(ctx.PostValue("inputRace"), 10, 32)
	inputServer, _ := strconv.ParseInt(ctx.PostValue("inputServer"), 10, 32)

	if inputWeekborn == 0 {
		session.SetFlash("msg", "錯誤：請選擇角色生日")
		ctx.Redirect(ctx.Path() + "/edit")
		return
	}
	if inputRace == 0 {
		session.SetFlash("msg", "錯誤：請選擇種族")
		ctx.Redirect(ctx.Path() + "/edit")
		return
	}
	if inputServer == 0 {
		session.SetFlash("msg", "錯誤：請選擇伺服器")
		ctx.Redirect(ctx.Path() + "/edit")
		return
	}

	updateData := bson.M{
		"$set": bson.M{
			"name":     ctx.PostValue("inputCharname"),
			"birthday": ctx.PostValue("inputBirthday"),
			"weekborn": inputWeekborn,
			"race":     inputRace,
			"server":   inputServer,
		}}
	if strings.HasPrefix(ctx.Path(), "/char/u/") {
		id, _ := primitive.ObjectIDFromHex(ctx.Params().Get("uid"))
		filter = bson.M{
			"uid": session.Get("_id"),
			"_id": id, //未編入角色(24碼)
		}
	} else {
		charid, _ := strconv.ParseInt(ctx.Params().Get("uid"), 10, 32)
		filter = bson.M{
			"uid":         session.Get("_id"),
			"characterId": charid,
		}
	}

	result, err := APIUpdateOneBase("characters", filter, updateData)

	if err != nil {
		//log.Fatal(err)
		session.SetFlash("msg", "發生錯誤，不知明原因.")
	} else {
		if result.ModifiedCount == 1 {
			session.SetFlash("msg", "更新成功")
		} else {
			session.SetFlash("msg", "更新異常，您可能未異動資料")
		}
	}
	ctx.Redirect("/user/char")
}

//PostCharUpload 上傳角色大頭照
func PostCharUpload(ctx iris.Context) {
	session := sessions.Get(ctx)
	coll := DBSource.db.Collection("characters")
	ctx.SetMaxRequestBodySize(512)
	/*
		inputImage
	*/

	file, _, err := ctx.FormFile("inputImage")
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.HTML("Error while uploading: <b>" + err.Error() + "</b>")
		return
	}
	defer file.Close()
	fname := ctx.PostValue("_id") + ".png"

	out, err := os.OpenFile("./public/avatar/"+fname, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		ctx.StatusCode(iris.StatusInternalServerError)
		ctx.HTML("Error while uploading: <b>" + err.Error() + "</b>")
		return
	}
	defer out.Close()
	io.Copy(out, file)

	var filter bson.M
	updateData := bson.M{
		"$set": bson.M{
			"imageUrl": fname,
		}}
	if strings.HasPrefix(ctx.Path(), "/char/u/") {
		id, _ := primitive.ObjectIDFromHex(ctx.Params().Get("uid"))
		filter = bson.M{
			"uid": session.Get("_id"),
			"_id": id, //未編入角色(24碼)
		}
	} else {
		charid, _ := strconv.ParseInt(ctx.Params().Get("uid"), 10, 32)
		filter = bson.M{
			"uid":         session.Get("_id"),
			"characterId": charid,
		}
	}
	result, err := coll.UpdateOne(context.TODO(), filter, updateData)
	if err != nil {
		//log.Fatal(err)
		session.SetFlash("msg", "發生錯誤，不知明原因.")
	} else {
		if result.ModifiedCount == 1 {
			session.SetFlash("msg", "更新成功")
		} else {
			session.SetFlash("msg", "更新上傳")
		}
	}
	ctx.Redirect("/user/char")
}
