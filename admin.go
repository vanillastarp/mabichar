package main

import (
	"context"
	"log"

	"github.com/kataras/iris/v12"
	"go.mongodb.org/mongo-driver/bson"
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
	ctx.Writef("技能列表")
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
	var result []bson.M
	coll := DBSource.db.Collection("admin_Servers")
	cur, err := coll.Find(context.Background(), bson.M{})
	defer cur.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	} else {
		if err = cur.All(context.Background(), &result); err != nil {
			log.Fatal(err)
		}
		ctx.ViewData("serverList", result)
		if err := ctx.View("admin/serverList.html"); err != nil {
			ctx.Application().Logger().Infof(err.Error())
		}
	}
}
