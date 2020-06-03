package main

import (
	"strconv"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/sessions"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

//GetAdminIndex 管理人員主要介面
func GetAdminIndex(ctx iris.Context) {
	if err := ctx.View("admin/index.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetGameVersion  資料庫改版紀錄
func GetGameVersion(ctx iris.Context) {
	ctx.Writef("改版紀錄列表")
}

//GetAchievements 成就列表
func GetAchievements(ctx iris.Context) {
	ctx.Writef("成就列表")
}

//GetSkills 技能列表
func GetSkills(ctx iris.Context) {

	result, err := APIQueryBase("admin_Skills", bson.M{})

	if err != nil {
		ctx.ViewData("message", err.Error())
	}

	ctx.ViewData("skillsList", result)
	if err := ctx.View("admin/skillsList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetTitles 稱號列表
func GetTitles(ctx iris.Context) {
	ctx.Writef("稱號列表")
}

//GetTalentMasters 一代宗師列表
func GetTalentMasters(ctx iris.Context) {
	ctx.Writef("一代宗師列表")
}

//GetPets 寵物列表
func GetPets(ctx iris.Context) {
	ctx.Writef("寵物列表")
}

//GetCollections 收集日誌
func GetCollections(ctx iris.Context) {
	ctx.Writef("收集日誌")
}

//GetEvents 官方活動
func GetEvents(ctx iris.Context) {
	ctx.Writef("官方活動")
}

//GetStories 主線列表
func GetStories(ctx iris.Context) {
	ctx.Writef("主線列表")
}

//GetServers 伺服器列表
func GetServers(ctx iris.Context) {
	session := sessions.Get(ctx)

	result, err := APIQueryBase("admin_Servers", bson.M{})

	if err != nil {
		ctx.ViewData("message", err.Error())
	}

	ctx.ViewData("message", session.GetFlashString("msg"))
	ctx.ViewData("serverList", result)
	if err := ctx.View("admin/serverList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetServerCreate 新增伺服器表單
func GetServerCreate(ctx iris.Context) {
	ctx.ViewData("template", map[string]string{
		"banner": "新增",
		"method": "POST",
		"action": "/admin/servers/create",
		"button": "新增",
	})
	if err := ctx.View("admin/serverForm.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PostServerCreate 新增伺服器資料
func PostServerCreate(ctx iris.Context) {
	session := sessions.Get(ctx)
	/*
	   inputServerid
	   inputServername
	   inputServerEngname
	*/
	inputServerid, _ := strconv.ParseInt(ctx.PostValue("inputServerid"), 10, 32)
	insertData := bson.M{
		"serverid":      inputServerid,
		"serverName":    ctx.PostValue("inputServername"),
		"serverEngName": ctx.PostValue("inputServerEngname"),
	}

	result, err := APIInsertOneBase("admin_Servers", insertData)

	if err != nil {
		session.SetFlash("msg", "發生錯誤，可能原因為重複編號.")
	} else {
		// log.Println("Added a new server with objectID: ", result.InsertedID)
		id := result.InsertedID.(primitive.ObjectID).Hex()
		session.SetFlash("msg", "已成功新增一筆伺服器ID:"+id)
	}
	ctx.Redirect("/admin/servers")
}

//GetServerEdit 編輯伺服器表單
func GetServerEdit(ctx iris.Context) {
	session := sessions.Get(ctx)

	inputServerid, _ := strconv.ParseInt(ctx.Params().Get("serverid"), 10, 32)
	filter := bson.M{
		"serverid": inputServerid,
	}

	result, err := APIQueryOneBase("admin_Servers", filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			session.SetFlash("msg", "命令無法操作，請確認資料是否存在")
		} else {
			session.SetFlash("msg", "命令無法操作 err:"+err.Error())
		}
		ctx.Redirect("/admin/servers")
		return
	}
	ctx.ViewData("serverData", result)
	ctx.ViewData("template", map[string]string{
		"banner": "編輯",
		"method": "PUT",
		"action": "/admin/servers/",
		"button": "更新",
	})

	if err := ctx.View("admin/serverForm.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PutServerUpdate 更新伺服器資料
func PutServerUpdate(ctx iris.Context) {
	session := sessions.Get(ctx)

	id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id"))
	if err != nil {
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
	} else {
		/*
		   inputServerid
		   inputServername
		   inputServerEngname
		*/
		inputServerid, _ := strconv.ParseInt(ctx.PostValue("inputServerid"), 10, 32)
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"serverid":      inputServerid,
				"serverName":    ctx.PostValue("inputServername"),
				"serverEngName": ctx.PostValue("inputServerEngname"),
			}}

		result, err := APIUpdateOneBase("admin_Servers", filter, updateData)

		if err != nil {
			session.SetFlash("msg", "發生錯誤，可能原因為重複編號.")
		} else {
			if result.ModifiedCount == 1 {
				session.SetFlash("msg", "更新成功")
			} else {
				session.SetFlash("msg", "更新異常，您可能未異動資料")
			}
		}
	}
	ctx.Redirect("/admin/servers")
}

//DelServer 刪除伺服器資料
func DelServer(ctx iris.Context) {
	session := sessions.Get(ctx)

	id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id"))

	if err != nil {
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
	} else {
		deleteData := bson.M{
			"_id": id,
		}

		result, err := APIDeleteOneBase("admin_Servers", deleteData)

		if err != nil {
			session.SetFlash("msg", "DelServer err:"+err.Error())
		} else {
			if result.DeletedCount == 0 {
				session.SetFlash("msg", "發生錯誤，未能刪除id: "+id.Hex())
			} else {
				session.SetFlash("msg", "已成功刪除id: "+id.Hex())
			}
		}
	}
	ctx.Redirect("/admin/servers")
}
