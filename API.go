package main

import (
	"context"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//------API space------

// type API struct {
// 	empty string
// }

//APIConnectDB 連接資料庫
func APIConnectDB(DBSource *DBStruct) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	clientOpts := options.Client().ApplyURI(os.Getenv("DSN"))
	client, err := mongo.Connect(context.TODO(), clientOpts)

	if err != nil {
		log.Fatal(err)
	}
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	} else {
		log.Println("[system] MongoDB is running.")
	}

	DBSource.DSN = os.Getenv("DSN")
	DBSource.DBName = os.Getenv("database")
	DBSource.client = client
	DBSource.db = client.Database(os.Getenv("database"))
}

//APIGetServerList 連接資料庫
func APIGetServerList(DBSource *DBStruct, AdminDB *AdminDBStruct) {
	serverList := make(map[int]interface{})
	ServerList := APIGetServers()
	for i, s := range ServerList {
		serverList[i+1] = bson.M{
			"serverName":    s["serverName"].(string),
			"serverEngName": s["serverEngName"].(string),
		}
	}
	AdminDB.servers = serverList
}

/*
//----------------------------------------------
// 此區域為提供資料庫操作API
//----------------------------------------------
*/

//APILogin 登入
// func APILogin(filter bson.M) (bson.M, error) {
// 	result, err := APIQueryOneBase("users_role", filter)
// 	if err != nil {
// 		if err == mongo.ErrNoDocuments {
// 			return bson.M{"msg": "無法登入，請確認您的帳號密碼是否正確"}, err
// 		}
// 		return bson.M{"msg": "無法登入"}, err
// 	}
// 	return bson.M{
// 		"authenticated": true,
// 		"username":      result["username"],
// 		"_id":           result["_id"],
// 		"role":          result["rolename"],
// 	}, nil
// }

//APIQueryBase 提供多筆查詢基底
func APIQueryBase(targetCollection string, filter bson.M) ([]bson.M, error) {
	var result []bson.M
	coll := DBSource.db.Collection(targetCollection)
	cur, err := coll.Find(context.Background(), filter)
	defer cur.Close(context.Background())
	if err != nil {
		return []bson.M{}, err
	}
	err = cur.All(context.Background(), &result)
	// if err != nil {
	// 	return []bson.M{}, err
	// }
	return result, err
}

//APIUpdateOneBase 提供單筆更新基底
func APIUpdateOneBase(targetCollection string, filter bson.M, update bson.M) (*mongo.UpdateResult, error) {
	coll := DBSource.db.Collection(targetCollection)
	result, err := coll.UpdateOne(context.TODO(), filter, update)
	return result, err
}

//APIQueryOneBase 提供單筆查詢基底
func APIQueryOneBase(targetCollection string, filter bson.M) (bson.M, error) {
	var result bson.M
	coll := DBSource.db.Collection(targetCollection)
	err := coll.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return bson.M{"msg": "命令無法操作，請確認資料是否存在"}, err
		}
		return bson.M{"msg": "發生錯誤"}, err
	}
	return result, err
}

//APIInsertOneBase 提供新增單筆資料
func APIInsertOneBase(targetCollection string, insertData bson.M) (*mongo.InsertOneResult, error) {
	coll := DBSource.db.Collection(targetCollection)
	result, err := coll.InsertOne(context.TODO(), insertData)
	return result, err
}

//APIDeleteOneBase 提供刪除單筆資料
func APIDeleteOneBase(targetCollection string, deleteData bson.M) (*mongo.DeleteResult, error) {
	coll := DBSource.db.Collection(targetCollection)
	result, err := coll.DeleteOne(context.TODO(), deleteData)
	return result, err
}

/*
//----------------------------------------------
// 此區域為提供資料API
//----------------------------------------------
*/

//APIGetSkills 提供技能清單
func APIGetSkills() []bson.M {
	res, err := APIQueryBase("admin_Skills", bson.M{})
	if err != nil {
		log.Println(err.Error())
		return []bson.M{}
	}
	return res
}

//APIGetTalents 提供才能清單
func APIGetTalents() []bson.M {
	res, err := APIQueryBase("admin_TalentMasters", bson.M{})
	if err != nil {
		log.Println(err.Error())
		return []bson.M{}
	}
	return res
}

//APIGetServers 提供伺服器清單
func APIGetServers() []bson.M {
	res, err := APIQueryBase("admin_Servers", bson.M{})
	if err != nil {
		log.Println(err.Error())
		return []bson.M{}
	}
	return res
}

//APIGetCharacters 提供角色清單
func APIGetCharacters(id primitive.ObjectID) []bson.M {
	res, err := APIQueryBase("characters", bson.M{"uid": id, "enabled": true})
	if err != nil {
		log.Println(err.Error())
		return []bson.M{}
	}
	return res
}

//APIGetSkillTypes 提供技能類別清單
func APIGetSkillTypes() map[int]string {
	skillTypes := map[int]string{
		1:   "生活",
		2:   "戰鬥",
		3:   "魔法",
		4:   "煉金術",
		5:   "武鬥術",
		6:   "音樂",
		7:   "人偶術",
		8:   "雙槍",
		9:   "忍者",
		10:  "鎖鏈",
		30:  "隱藏才能",
		50:  "變身",
		51:  "半神化",
		52:  "騎士團",
		53:  "異神化",
		54:  "特性",
		55:  "寵物特技",
		100: "動作",
	}
	return skillTypes
}

//APIGetTalentTypes 提供才能類別清單
func APIGetTalentTypes() map[int]string {
	talentTypes := map[int]string{
		985:  "冒險家",
		1000: "戰士",
		999:  "魔法師",
		998:  "弓箭手",
		997:  "商人",
		996:  "戰鬥鍊金術師",
		995:  "武鬥家",
		994:  "吟遊詩人",
		1001: "人偶師",
		993:  "長槍手",
		992:  "聖職者",
		991:  "鍊成鍊金術師",
		990:  "廚師",
		989:  "鐵匠",
		988:  "裁縫師",
		987:  "藥劑師",
		986:  "木匠",
		1002: "雙槍手",
		1003: "忍者",
		1004: "闇夜鎖鏈師",
	}
	return talentTypes
}

// //APIGetTitleType 提供稱號編號範圍
// func APIGetTitleType() map[int]string {

// }

/*
//----------------------------------------------
// 此區域為雜項API
//----------------------------------------------
*/

//APIParseInt 字串轉數字
func APIParseInt(s string) int32 {
	o, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return -1
	}
	return int32(o)
}

//APIBool 字串轉布林
func APIBool(s string) bool {
	if s == "true" {
		return true
	}
	return false
}
