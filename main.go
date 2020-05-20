package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

//DBStruct 資料庫連線結構
type DBStruct struct {
	DSN    string
	DBName string
	db     *mongo.Database
	client *mongo.Client
}

//AdminDBStruct 固定型資料結構
type AdminDBStruct struct {
	servers map[int]interface{}
}

//DBSource 資料庫共用資源
var DBSource = DBStruct{}

//AdminDB 資料庫固定型資料
var AdminDB = AdminDBStruct{}

//SessionStruct Session結構表
type SessionStruct struct {
	authenticated bool
	username      string `bson:"username"`
	_id           int32  `bson:"_id"`
	role          string `bson:"rolename"`
}

//connDB 連接資料庫
func connDB(DBSource *DBStruct, AdminDB *AdminDBStruct) {
	var envFileName = ".env"

	err := godotenv.Load(envFileName)
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	//test Database
	//ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOpts := options.Client().ApplyURI(os.Getenv("DSN"))
	client, err := mongo.Connect(context.TODO(), clientOpts)

	if err != nil {
		log.Fatal(err)
	}
	// defer func() {
	// 	if err = client.Disconnect(context.TODO()); err != nil {
	// 		log.Fatal(err)
	// 	} else {
	// 		log.Println("[test] End test MongoDB.")
	// 	}
	// }()
	if err = client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	} else {
		log.Println("[test] MongoDB is running.")
	}

	DBSource.DSN = os.Getenv("DSN")
	DBSource.DBName = os.Getenv("database")
	DBSource.client = client
	DBSource.db = client.Database(os.Getenv("database"))

	//Take admin_Servers List
	// type Servers struct {
	// 	serverName string
	// 	serverid   int
	// }
	coll := DBSource.db.Collection("admin_Servers")
	cur, err := coll.Find(context.Background(), bson.M{})
	defer cur.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	} else {
		serverList := make(map[int]interface{})
		var i = 1
		for cur.Next(context.Background()) {
			var tmp bson.M
			if err = cur.Decode(&tmp); err != nil {
				log.Fatal(err)
			}
			//log.Println(tmp)
			// serverList[i] = tmp["serverName"].(string)
			serverList[i] = bson.M{
				"serverName":    tmp["serverName"].(string),
				"serverEngName": tmp["serverEngName"].(string),
			}
			i++
		}
		AdminDB.servers = serverList

	}

}

func main() {

	connDB(&DBSource, &AdminDB)
	defer DBSource.client.Disconnect(context.TODO())

	app := iris.New()
	app.OnErrorCode(iris.StatusNotFound, notFound)
	app.Use(RunningLog)
	app.Use(recover.New())

	sess := sessions.New(sessions.Config{
		Cookie:  "mabicharSession",
		Expires: 45 * time.Minute,
	})
	app.Use(sess.Handler())

	// Load all templates from the "./views" folder
	// where extension is ".html" and parse them
	// using the standard `html/template` package.
	var t = iris.HTML("./views", ".html").Reload(true)
	// t.AddFunc("serverList", func(i int) string {
	// 	return AdminDB.servers[i].(string)
	// })
	app.RegisterView(t)
	app.Layout("shared/layout.html")
	config := iris.WithConfiguration(iris.YAML("./configs/iris.yaml"))

	app.PartyFunc("/user", func(user iris.Party) {
		user.Use(authentication)

		user.Get("/", GetIndex) //dashboard here
		user.Get("/edit", GetEditUser)
		user.Post("/update", PostUpdateUser)
		user.Get("/char", GetCharList)
		user.Get("/newchar", GetNewChar)
		user.Post("/newchar", PostNewChar)
	})
	app.PartyFunc("/char", func(user iris.Party) {
		user.Use(authentication)

		user.Get("/{uid: int}", GetChar)                               //已編入資料庫角色
		user.Get("/u/{uid: string regexp([0-9a-f]) max(24)}", GetChar) //未編入資料庫角色
	})
	app.PartyFunc("/admin", func(user iris.Party) {
		user.Use(authentication)
		user.Use(adminOnly)
		user.Layout("admin/layout.html")

		user.Get("/", GetAdminIndex)
		user.Get("/game_version", GetGameVersion)
		user.Get("/achievements", GetAchievements)
		user.Get("/skills", GetSkills)
		user.Get("/titles", GetTitles)
		user.Get("/talent_masters", GetTalentMasters)
		user.Get("/pets", GetPets)
		user.Get("/collections", GetCollections)
		user.Get("/events", GetEvents)
		user.Get("/stories", GetStories)
		user.Get("/servers", GetServers)
	})
	app.PartyFunc("/", func(guest iris.Party) {
		guest.Use(authenticatedGuest)

		guest.Get("/", func(ctx iris.Context) {
			// session := sessions.Get(ctx)
			// auth, _ := session.GetBoolean("authenticated")
			// ctx.ViewData("auth", strconv.FormatBool(auth))

			ctx.View("index.html")
		})
		guest.Get("/register", func(ctx iris.Context) {
			ctx.View("register.html")
		})
		guest.Post("/register", func(ctx iris.Context) {
			ctx.Writef("in dev")
		})
		guest.Get("/login", func(ctx iris.Context) {
			session := sessions.Get(ctx)
			if auth, _ := session.GetBoolean("authenticated"); auth {
				ctx.Redirect("/user")
			} else {
				ctx.View("login.html")
			}
		})
		guest.Post("/login", func(ctx iris.Context) {
			session := sessions.Get(ctx)

			// type User struct {
			// 	_id             primitive.ObjectID `bson:"_id"`
			// 	username        string             `bson:"username"`
			// 	password        string             `bson:"password"`
			// 	enabled         bool               `bson:"enabled"`
			// 	createTimestamp int                `bson:"create_timestamp"`
			// 	modifyTimestamp int                `bson:"modify_timestamp"`
			// }

			var result bson.M

			// coll := DBSource.client.Database(DBSource.DBName).Collection("users")
			// filter := bson.M{"username": "root"}
			// cur, err := coll.Find(context.TODO(), filter, findOptions) //.Decode(&result)

			//coll := DBSource.client.Database(DBSource.DBName).Collection("users_role")
			coll := DBSource.db.Collection("users_role")

			//dbctx, _ := context.WithTimeout(context.Background(), 15*time.Second)

			//原始簡單版
			filter := bson.M{
				"username": ctx.PostValue("username"),
				"password": "",
				"enabled":  true,
			}

			err := coll.FindOne(context.TODO(), filter).Decode(&result)
			// log.Println("result data: ", result)

			// matchStage := bson.D{{"$match", bson.M{
			// 	"username": ctx.PostValue("username"),
			// 	"password": "",
			// 	"enabled":  true}}}

			// lookupStage := bson.D{{"$lookup", bson.M{
			// 	"from":         "roles",
			// 	"localField":   "role",
			// 	"foreignField": "roleid",
			// 	"as":           "roles"}}}

			// unwindStage := bson.D{{"$unwind", bson.M{
			// 	"path": "$roles"}}}
			// addFieldStage := bson.D{{"$addFields", bson.M{
			// 	"rolename": "$roles.rolename"}}}

			// pip := mongo.Pipeline{matchStage, lookupStage, unwindStage, addFieldStage}
			// cur, err := coll.Aggregate(context.Background(), pip)

			// defer cur.Close(context.Background())
			// var results []bson.M
			// if err := cur.All(context.Background(), &results); err != nil {
			// 	log.Fatal(err)
			// }
			// log.Println("result data: ", result)

			// log.Println("cursorID: ", cur.ID())
			// if !cur.TryNext(context.Background()) {
			// 	ctx.ViewData("message", "無法登入，請確認您的帳號密碼是否正確!")
			// 	ctx.View("login.html")
			// 	return
			// } else {

			// 	log.Println(cur.Decode(&result))
			// 	log.Println("cursorID2: ", cur.ID())
			// 	// for cur.Next(context.Background()) {
			// 	// 	cur.Decode(&result)
			// 	// }
			// }

			if err != nil {
				if err == mongo.ErrNoDocuments {
					ctx.ViewData("message", "無法登入，請確認您的帳號密碼是否正確")
					ctx.View("login.html")
					return
				}
				log.Fatal(err)
			} else {
				session.Set("authenticated", true)
				session.Set("username", result["username"])
				session.Set("_id", result["_id"])
				// log.Println("_id::::::", result["_id"])
				// _id, _ := session.GetInt("_id")
				// log.Println("session._id::::", _id)
				session.Set("role", result["rolename"])
				if result["rolename"] == "Admin" {
					ctx.Redirect("/admin")
				} else {
					ctx.Redirect("/user")
				}
				// if result["role"] == int32(1) {
				// 	session.Set("role", "Admin")
				// 	ctx.Redirect("/admin")
				// } else {
				// 	session.Set("role", "User")
				// 	ctx.Redirect("/user")
				// }
			}
		})

		//上線後須改回POST
		guest.Get("/logout", func(ctx iris.Context) {
			session := sessions.Get(ctx)
			if auth, _ := session.GetBoolean("authenticated"); auth {
				//session.Set("authenticated", false)
				session.Clear()
				// session.Delete("authenticated")
				// session.Delete("username")
				// session.Delete("role")
			}
			ctx.Redirect("/")
		})
		guest.Get("/forget_password", func(ctx iris.Context) {
			ctx.Writef("here is forget_password page")
		})
		guest.Post("/forget_password", func(ctx iris.Context) {
			ctx.Writef("[forget_password]in dev")
		})
		guest.Get("/about", func(ctx iris.Context) {
			ctx.Writef("Mabinogi Charactor Memoir")
		})
		guest.Get("/readme", func(ctx iris.Context) {
			ctx.Writef("使用前請參考說明書")
		})
		guest.Get("/share", func(ctx iris.Context) {
			ctx.ViewData("serverList", AdminDB.servers)
			if err := ctx.View("share.html"); err != nil {
				ctx.Application().Logger().Infof(err.Error())
			}
		})
	})

	app.HandleDir("/static", "./public")
	app.Listen(":8080", config)
}
func authentication(ctx iris.Context) {
	//驗證使用者是否登入
	//否則導向到登入介面
	session := sessions.Get(ctx)
	auth, _ := session.GetBoolean("authenticated")
	if !auth {
		ctx.Redirect("/login")
	}

	ctx.ViewData("auth", strconv.FormatBool(auth))
	ctx.ViewData("username", session.GetString("username"))
	// ctx.ViewData("_id", session.Get("_id"))
	// ctx.ViewData("role", session.GetString("role"))

	// log.Println("role: ", session.Get("role"))
	// log.Println("_id: ", session.Get("_id"))

	ctx.Next()
}

func adminOnly(ctx iris.Context) {
	session := sessions.Get(ctx)

	if session.GetString("role") != "Admin" {
		ctx.Redirect("/user")
	}
	ctx.Next()
}

func authenticatedGuest(ctx iris.Context) {
	session := sessions.Get(ctx)
	if auth, _ := session.GetBoolean("authenticated"); auth {
		ctx.ViewData("auth", strconv.FormatBool(auth))
		ctx.ViewData("username", session.GetString("username"))
		// ctx.ViewData("_id", session.Get("_id"))
		// ctx.ViewData("role", session.GetString("role"))
		//ctx.ViewData("role", session.Get("role"))
	}
	ctx.Next()
}

//RunningLog  記錄使用者瀏覽路徑
func RunningLog(ctx iris.Context) {
	ctx.Application().Logger().Infof("Runs method:%s %s", ctx.Method(), ctx.Path())
	ctx.Next()
}

func notFound(ctx iris.Context) {
	//ctx.ViewData("message", "Did you forget something?")
	ctx.View("errors/404.html") //設定找不到業面
}
