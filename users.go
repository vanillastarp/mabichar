package main

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"go.mongodb.org/mongo-driver/bson"
)

//GetIndex 使用者的dashboard介面
func GetIndex(ctx iris.Context) {
	//session := sessions.Get(ctx)

	var result []bson.M
	coll := DBSource.db.Collection("characters")

	query := bson.M{
		"uid": sessions.Get(ctx).Get("_id"),
	}

	cur, err := coll.Find(context.Background(), query)
	defer cur.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	} else {
		if err = cur.All(context.Background(), &result); err != nil {
			log.Fatal(err)
		}

		ctx.ViewData("charAmount", len(result))

		if err := ctx.View("users/dashboard.html"); err != nil {
			ctx.Application().Logger().Infof(err.Error())
		}
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

	var result []bson.M
	coll := DBSource.db.Collection("characters")

	query := bson.M{
		"uid": session.Get("_id"), //用使用者的_id去找characters collection的uid
	}

	// err := coll.FindOne(context.TODO(), query).Decode(&result)
	cur, err := coll.Find(context.Background(), query)
	defer cur.Close(context.Background())
	if err != nil {
		// if err == mongo.ErrNoDocuments {
		// 	ctx.ViewData("message", "目前您未有任何角色，請開始新增角色")
		// 	ctx.View("users/charList.html")
		// 	return
		// }
		log.Fatal(err)
	} else {
		if err = cur.All(context.Background(), &result); err != nil {
			log.Fatal(err)
		}
		//log.Println("result: ", result)

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

			// log.Println(map[int]string{1: "人類", 2: "精靈", 3: "巨人"})
			// log.Println(AdminDB.servers)

		}
		ctx.ViewData("message", session.GetFlashString("msg"))
		if err := ctx.View("users/charList.html"); err != nil {
			ctx.Application().Logger().Infof(err.Error())
		}
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
	inputWeekborn, _ := strconv.ParseInt(ctx.PostValue("inputWeekborn"), 10, 0)
	inputRace, _ := strconv.ParseInt(ctx.PostValue("inputRace"), 10, 0)
	inputServer, _ := strconv.ParseInt(ctx.PostValue("inputServer"), 10, 0)
	insertData := bson.M{
		"uid":              session.Get("_id"),
		"characterId":      0,
		"name":             ctx.PostValue("inputCharname"),
		"birthday":         ctx.PostValue("inputBirthday"),
		"weekborn":         int32(inputWeekborn),
		"race":             int32(inputRace),
		"server":           int32(inputServer),
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
	ctx.Writef(ctx.Path())
}

//GetEditChar 編輯角色基本資料
func GetEditChar(ctx iris.Context) {
	session := sessions.Get(ctx)
	var result bson.M

	coll := DBSource.db.Collection("characters")

	var query bson.M
	// log.Println(ctx.Path())
	if strings.HasPrefix(ctx.Path(), "/char/u/") {
		id, _ := primitive.ObjectIDFromHex(ctx.Params().Get("uid"))
		query = bson.M{
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
		charid, _ := strconv.ParseInt(ctx.Params().Get("uid"), 10, 0)
		query = bson.M{
			"uid":         session.Get("_id"),
			"characterId": int32(charid),
		}
		ctx.ViewData("template", map[string]string{
			"banner": "編輯",
			"method": "PUT",
			"action": "/char/" + ctx.Params().Get("uid"),
			"button": "更新",
		})
	}
	// log.Println(query)
	err := coll.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			session.SetFlash("msg", "命令無法操作，請確認資料是否存在 err:"+err.Error())
		} else {
			//log.Fatal(err)
			session.SetFlash("msg", "發生錯誤err: "+err.Error())
		}
		ctx.Redirect("/user/char")
	} else {
		ctx.ViewData("charData", result)
		if err := ctx.View("users/CharForm.html"); err != nil {
			ctx.Application().Logger().Infof(err.Error())
		}
	}
}

//PutCharUpdate 更新資料庫角色
func PutCharUpdate(ctx iris.Context) {

}

// //編輯已編入資料庫角色大頭照
// func GetCharUpload(ctx iris.Context) {

// }

//PostCharUpload 上傳角色大頭照
func PostCharUpload(ctx iris.Context) {

}

// //編輯未編入資料庫角色大頭照
// func GetCharUpload(ctx iris.Context) {

// }
