package main

import (
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

//GetAdminListBase 提供資料列表
func GetAdminListBase(ctx iris.Context, targetCollecion string, view string) {
	session := sessions.Get(ctx)

	result, err := APIQueryBase(targetCollecion, bson.M{})

	if err != nil {
		ctx.ViewData("message", err.Error())
	}

	ctx.ViewData("message", session.GetFlashString("msg"))
	ctx.ViewData(view, result)
	if err := ctx.View("admin/" + view + ".html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetAdminCreateBase 提供建立表格頁面
func GetAdminCreateBase(ctx iris.Context, function string, form string) {
	ctx.ViewData("template", map[string]string{
		"banner": "新增",
		"method": "POST",
		"action": "/admin/" + function + "/create",
		"button": "新增",
	})
	if err := ctx.View("admin/" + form + ".html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PostAdminCreateBase 提供新增資料
func PostAdminCreateBase(ctx iris.Context, targetCollection string, insertData bson.M, backTo string) {
	session := sessions.Get(ctx)

	result, err := APIInsertOneBase(targetCollection, insertData)

	if err != nil {
		session.SetFlash("msg", "發生錯誤，可能原因為重複編號.")
	} else {
		id := result.InsertedID.(primitive.ObjectID).Hex()
		session.SetFlash("msg", "已成功新增一筆ID:"+id)
	}

	ctx.Redirect("/admin/" + backTo)
}

//GetAdminEditBase 提供編輯表格頁面
func GetAdminEditBase(ctx iris.Context, targetCollecion string, filter bson.M, function string, view string, form string) {
	session := sessions.Get(ctx)

	result, err := APIQueryOneBase(targetCollecion, filter)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			session.SetFlash("msg", "命令無法操作，請確認資料是否存在")
		} else {
			session.SetFlash("msg", "命令無法操作 err:"+err.Error())
		}
		ctx.Redirect("/admin/" + function)
		return
	}
	ctx.ViewData(view, result)
	ctx.ViewData("template", map[string]string{
		"banner": "編輯",
		"method": "PUT",
		"action": "/admin/" + function + "/",
		"button": "更新",
	})

	if err := ctx.View("admin/" + form + ".html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//PutAdminUpdateBase 提供修改資料
func PutAdminUpdateBase(ctx iris.Context, targetCollecion string, filter bson.M, updateData bson.M, backTo string) {
	session := sessions.Get(ctx)

	result, err := APIUpdateOneBase(targetCollecion, filter, updateData)

	if err != nil {
		session.SetFlash("msg", "發生錯誤，可能原因為重複編號.")
	} else {
		if result.ModifiedCount == 1 {
			session.SetFlash("msg", "更新成功")
		} else {
			session.SetFlash("msg", "更新異常，您可能未異動資料")
		}
	}
	ctx.Redirect("/admin/" + backTo)
}

//DelAdminBase 提供刪除資料基底
func DelAdminBase(ctx iris.Context, targetCollection string, backTo string) {
	session := sessions.Get(ctx)

	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
	} else {
		deleteData := bson.M{
			"_id": id,
		}

		if result, err := APIDeleteOneBase(targetCollection, deleteData); err != nil {
			session.SetFlash("msg", "["+targetCollection+"]DelAdminBase err:"+err.Error())
		} else {
			if result.DeletedCount == 0 {
				session.SetFlash("msg", "發生錯誤，未能刪除id: "+id.Hex())
			} else {
				session.SetFlash("msg", "已成功刪除id: "+id.Hex())
			}
		}
	}
	ctx.Redirect("/admin/" + backTo)
}

//GetGameVersion  資料庫改版紀錄
func GetGameVersion(ctx iris.Context) {
	ctx.Writef("改版紀錄列表")
}

//GetAchievements 成就列表
func GetAchievements(ctx iris.Context) {
	ctx.Writef("成就列表")
}

//------------------- Skills --------------------

//GetSkills 技能列表
func GetSkills(ctx iris.Context) {
	session := sessions.Get(ctx)
	ctx.ViewData("skilltypes", APIGetSkillTypes())
	ctx.ViewData("message", session.GetFlashString("msg"))
	if err := ctx.View("admin/skillsList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetSkillCreate 新增技能表單
func GetSkillCreate(ctx iris.Context) {
	GetAdminCreateBase(ctx, "skills", "skillsForm")
}

//PostSkillCreate 新增技能資料
func PostSkillCreate(ctx iris.Context) {
	/*
	   inputSkillType
	   inputSkillid
	   inputSkillName
	   inputskilllevel
	   inputRace
	   inputUpgrade
	*/
	insertData := bson.M{
		"skilltype": APIParseInt(ctx.PostValue("inputSkillType")),
		"skillid":   APIParseInt(ctx.PostValue("inputSkillid")),
		"skillName": ctx.PostValue("inputSkillName"),
		"maxlv":     APIParseInt(ctx.PostValue("inputskilllevel")),
		"race":      APIParseInt(ctx.PostValue("inputRace")),
		"upgrade":   APIBool(ctx.PostValue("inputUpgrade")),
	}
	PostAdminCreateBase(ctx, "admin_Skills", insertData, "skills")
}

//GetSkillEdit 編輯技能表單
func GetSkillEdit(ctx iris.Context) {
	filter := bson.M{
		"skillid": APIParseInt(ctx.Params().Get("skillid")),
	}

	GetAdminEditBase(ctx, "admin_Skills", filter, "skills", "skillData", "skillsForm")
}

//PutSkillUpdate 更新技能資料
func PutSkillUpdate(ctx iris.Context) {
	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session := sessions.Get(ctx)
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
		ctx.Redirect("/admin/skills")
	} else {
		/*
		   inputSkillType
		   inputSkillid
		   inputSkillName
		   inputskilllevel
		   inputRace
		   inputUpgrade
		*/
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"skilltype": APIParseInt(ctx.PostValue("inputSkillType")),
				"skillid":   APIParseInt(ctx.PostValue("inputSkillid")),
				"skillName": ctx.PostValue("inputSkillName"),
				"maxlv":     APIParseInt(ctx.PostValue("inputskilllevel")),
				"race":      APIParseInt(ctx.PostValue("inputRace")),
				"upgrade":   APIBool(ctx.PostValue("inputUpgrade")),
			}}
		PutAdminUpdateBase(ctx, "admin_Skills", filter, updateData, "skills")
	}
}

//DelSkill 刪除技能資料
func DelSkill(ctx iris.Context) {
	DelAdminBase(ctx, "admin_Skills", "skills")
}

//------------------- Titles --------------------

//GetTitles 列出稱號清單
func GetTitles(ctx iris.Context) {
	session := sessions.Get(ctx)
	ctx.ViewData("titletypes", APIGetTitleTypes())
	ctx.ViewData("message", session.GetFlashString("msg"))
	if err := ctx.View("admin/titlesList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetTitleCreate 新增稱號表單
func GetTitleCreate(ctx iris.Context) {
	GetAdminCreateBase(ctx, "titles", "titlesForm")
}

//PostTitleCreate 新增稱號資料
func PostTitleCreate(ctx iris.Context) {
	/*
	   inputType
	   inputTitleid
	   inputTitleDefaultName
	   inputTitleMaleName
	   inputTitleFemaleName
	   inputTitleUniName
	*/
	insertData := bson.M{
		"type":        APIParseInt(ctx.PostValue("inputType")),
		"id":          APIParseInt(ctx.PostValue("inputTitleid")),
		"defaultname": ctx.PostValue("inputTitleDefaultName"),
		"malename":    ctx.PostValue("inputTitleMaleName"),
		"femalename":  ctx.PostValue("inputTitleFemaleName"),
		"uniname":     ctx.PostValue("inputTitleUniName"),
	}
	PostAdminCreateBase(ctx, "admin_Titles", insertData, "titles")
}

//GetTitleEdit 編輯稱號表單
func GetTitleEdit(ctx iris.Context) {
	filter := bson.M{
		"id": APIParseInt(ctx.Params().Get("titleid")),
	}

	GetAdminEditBase(ctx, "admin_Titles", filter, "titles", "titleData", "titlesForm")
}

//PutTitleUpdate 更新稱號資料
func PutTitleUpdate(ctx iris.Context) {
	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session := sessions.Get(ctx)
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
		ctx.Redirect("/admin/titles")
	} else {
		/*
		   inputType
		   inputTitleid
		   inputTitleDefaultName
		   inputTitleMaleName
		   inputTitleFemaleName
		   inputTitleUniName
		*/
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"type":        APIParseInt(ctx.PostValue("inputType")),
				"id":          APIParseInt(ctx.PostValue("inputTitleid")),
				"defaultname": ctx.PostValue("inputTitleDefaultName"),
				"malename":    ctx.PostValue("inputTitleMaleName"),
				"femalename":  ctx.PostValue("inputTitleFemaleName"),
				"uniname":     ctx.PostValue("inputTitleUniName"),
			}}
		PutAdminUpdateBase(ctx, "admin_Titles", filter, updateData, "titles")
	}
}

//DelTitle 刪除稱號資料
func DelTitle(ctx iris.Context) {
	DelAdminBase(ctx, "admin_Titles", "titles")
}

//------------------- Talent_masters --------------------

//GetTalentMasters 才能列表
func GetTalentMasters(ctx iris.Context) {
	session := sessions.Get(ctx)
	ctx.ViewData("talenttypes", APIGetTalentTypes())
	ctx.ViewData("message", session.GetFlashString("msg"))
	if err := ctx.View("admin/talentMastersList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetTalentMasterCreate 新增才能表單
func GetTalentMasterCreate(ctx iris.Context) {
	GetAdminCreateBase(ctx, "talentmasters", "talentMastersForm")
}

//PostTalentMasterCreate 新增才能資料
func PostTalentMasterCreate(ctx iris.Context) {
	/*
		inputCategory
		inputTalentid
		inputTalentTitle
		inputTalentlevel
	*/
	insertData := bson.M{
		"category":    APIParseInt(ctx.PostValue("inputCategory")),
		"talentid":    APIParseInt(ctx.PostValue("inputTalentid")),
		"talenttitle": ctx.PostValue("inputTalentTitle"),
		"talentlevel": APIParseInt(ctx.PostValue("inputTalentlevel")),
	}
	PostAdminCreateBase(ctx, "admin_TalentMasters", insertData, "talentmasters")
}

//GetTalentMasterEdit 編輯才能表單
func GetTalentMasterEdit(ctx iris.Context) {
	filter := bson.M{
		"talentid": APIParseInt(ctx.Params().Get("talentid")),
	}

	GetAdminEditBase(ctx, "admin_TalentMasters", filter, "talentmasters", "talentData", "talentMastersForm")
}

//PutTalentMasterUpdate 更新才能資料
func PutTalentMasterUpdate(ctx iris.Context) {
	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session := sessions.Get(ctx)
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
		ctx.Redirect("/admin/talentmasters")
	} else {
		/*
			inputCategory
			inputTalentid
			inputTalentTitle
			inputTalentlevel
		*/
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"category":    APIParseInt(ctx.PostValue("inputCategory")),
				"talentid":    APIParseInt(ctx.PostValue("inputTalentid")),
				"talenttitle": ctx.PostValue("inputTalentTitle"),
				"talentlevel": APIParseInt(ctx.PostValue("inputTalentlevel")),
			}}

		PutAdminUpdateBase(ctx, "admin_TalentMasters", filter, updateData, "talentmasters")
	}
}

//DelTalentMaster 刪除才能資料
func DelTalentMaster(ctx iris.Context) {
	DelAdminBase(ctx, "admin_TalentMasters", "talentmasters")
}

//------------------- Pets --------------------

//GetPets 列出寵物清單
func GetPets(ctx iris.Context) {
	session := sessions.Get(ctx)
	ctx.ViewData("pettypes", APIGetPetTypes())
	ctx.ViewData("message", session.GetFlashString("msg"))
	if err := ctx.View("admin/petsList.html"); err != nil {
		ctx.Application().Logger().Infof(err.Error())
	}
}

//GetPetCreate 新增寵物表單
func GetPetCreate(ctx iris.Context) {
	GetAdminCreateBase(ctx, "pets", "petsForm")
}

//PostPetCreate 新增寵物資料
func PostPetCreate(ctx iris.Context) {
	/*
		inputType
		inputPetid
		inputPetName
	*/
	insertData := bson.M{
		"type":      APIParseInt(ctx.PostValue("inputType")),
		"ID":        APIParseInt(ctx.PostValue("inputPetid")),
		"localName": ctx.PostValue("inputPetName"),
	}
	PostAdminCreateBase(ctx, "admin_Pets", insertData, "pets")
}

//GetPetEdit 編輯寵物表單
func GetPetEdit(ctx iris.Context) {
	filter := bson.M{
		"ID": APIParseInt(ctx.Params().Get("petid")),
	}

	GetAdminEditBase(ctx, "admin_Pets", filter, "pets", "petData", "petsForm")
}

//PutPetUpdate 更新寵物資料
func PutPetUpdate(ctx iris.Context) {
	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session := sessions.Get(ctx)
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
		ctx.Redirect("/admin/talentmasters")
	} else {
		/*
			inputType
			inputPetid
			inputPetName
		*/
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"type":      APIParseInt(ctx.PostValue("inputType")),
				"ID":        APIParseInt(ctx.PostValue("inputPetid")),
				"localName": ctx.PostValue("inputPetName"),
			}}

		PutAdminUpdateBase(ctx, "admin_Pets", filter, updateData, "pets")
	}
}

//DelPet 刪除寵物資料
func DelPet(ctx iris.Context) {
	DelAdminBase(ctx, "admin_Pets", "pets")
}

//------------------- Collections --------------------

//GetCollections 收集日誌
func GetCollections(ctx iris.Context) {
	ctx.Writef("收集日誌")
}

//------------------- Events --------------------

//GetEvents 官方活動
func GetEvents(ctx iris.Context) {
	ctx.Writef("官方活動")
}

//------------------- Stories --------------------

//GetStories 主線列表
func GetStories(ctx iris.Context) {
	ctx.Writef("主線列表")
}

//------------------- Servers --------------------

//GetServers 伺服器列表
func GetServers(ctx iris.Context) {
	GetAdminListBase(ctx, "admin_Servers", "serverList")
}

//GetServerCreate 新增伺服器表單
func GetServerCreate(ctx iris.Context) {
	GetAdminCreateBase(ctx, "servers", "serverForm")
}

//PostServerCreate 新增伺服器資料
func PostServerCreate(ctx iris.Context) {
	/*
	   inputServerid
	   inputServername
	   inputServerEngname
	*/
	insertData := bson.M{
		"serverid":      APIParseInt(ctx.PostValue("inputServerid")),
		"serverName":    ctx.PostValue("inputServername"),
		"serverEngName": ctx.PostValue("inputServerEngname"),
	}
	PostAdminCreateBase(ctx, "admin_Servers", insertData, "servers")
}

//GetServerEdit 編輯伺服器表單
func GetServerEdit(ctx iris.Context) {
	filter := bson.M{
		"serverid": APIParseInt(ctx.Params().Get("serverid")),
	}
	GetAdminEditBase(ctx, "admin_Servers", filter, "servers", "serverData", "serverForm")
}

//PutServerUpdate 更新伺服器資料
func PutServerUpdate(ctx iris.Context) {
	if id, err := primitive.ObjectIDFromHex(ctx.PostValue("_id")); err != nil {
		session := sessions.Get(ctx)
		session.SetFlash("msg", "primitive.ObjectIDFromHex ERROR: "+err.Error())
		ctx.Redirect("/admin/servers")
	} else {
		/*
		   inputServerid
		   inputServername
		   inputServerEngname
		*/
		filter := bson.M{
			"_id": id,
		}
		updateData := bson.M{
			"$set": bson.M{
				"serverid":      APIParseInt(ctx.PostValue("inputServerid")),
				"serverName":    ctx.PostValue("inputServername"),
				"serverEngName": ctx.PostValue("inputServerEngname"),
			}}
		PutAdminUpdateBase(ctx, "admin_Servers", filter, updateData, "servers")
	}
}

//DelServer 刪除伺服器資料
func DelServer(ctx iris.Context) {
	DelAdminBase(ctx, "admin_Servers", "servers")
}
