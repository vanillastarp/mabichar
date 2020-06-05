package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/methodoverride"
	"github.com/kataras/iris/v12/middleware/recover"
	"github.com/kataras/iris/v12/sessions"
	"golang.org/x/crypto/bcrypt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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

func main() {

	APIConnectDB(&DBSource)
	APIGetServerList(&DBSource, &AdminDB)

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

	//HTTP method override
	mo := methodoverride.New(
		methodoverride.SaveOriginalMethod("_originalMethod"),
	)
	app.WrapRouter(mo)
	app.RegisterView(iris.HTML("./views", ".html").Reload(true))
	app.Layout("shared/layout.html")
	config := iris.WithConfiguration(iris.YAML("./configs/iris.yaml"))

	// app.Post("/api/Login", func(ctx iris.Context) {
	// 	session := sessions.Get(ctx)

	// 	filter := bson.M{
	// 		"username": ctx.PostValue("username"),
	// 		"password": "",
	// 		"enabled":  true,
	// 	}

	// 	result, err := APILogin(filter)

	// 	if err != nil {
	// 		ctx.JSON(result)
	// 	} else {
	// 		session.Set("authenticated", true)
	// 		session.Set("username", result["username"])
	// 		session.Set("_id", result["_id"])
	// 		session.Set("role", result["rolename"])
	// 		// ctx.SetCookie()
	// 		ctx.JSON(result)
	// 	}
	// })

	app.PartyFunc("/api", func(api iris.Party) {
		api.Use(authentication)
		api.Get("/GetSkills", func(ctx iris.Context) { ctx.JSON(APIGetSkills()) })
		api.Get("/GetServers", func(ctx iris.Context) { ctx.JSON(APIGetServers()) })
		api.Get("/GetCharacters", func(ctx iris.Context) { ctx.JSON(APIGetCharacters(sessions.Get(ctx).Get("_id").(primitive.ObjectID))) })
	})

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

		user.Get("/{uid: int}", GetChar)                                              //瀏覽已編入資料庫角色
		user.Get("/u/{uid: string regexp([0-9a-f]) max(24)}", GetChar)                //瀏覽未編入資料庫角色
		user.Get("/{uid: int}/edit", GetEditChar)                                     //編輯已編入資料庫角色基本資料
		user.Get("/u/{uid: string regexp([0-9a-f]) max(24)}/edit", GetEditChar)       //編輯未編入資料庫角色基本資料
		user.Put("/{uid: int}", PutCharUpdate)                                        //更新已編入資料庫角色
		user.Put("/u/{uid: string regexp([0-9a-f]) max(24)}", PutCharUpdate)          //更新未編入資料庫角色
		user.Post("/{uid: int}/upload", PostCharUpload)                               //上傳已編入資料庫角色大頭照
		user.Post("/u/{uid: string regexp([0-9a-f]) max(24)}/upload", PostCharUpload) //上傳未編入資料庫角色大頭照
		user.Delete("/{uid: int}", DeleteChar)                                        //刪除資料庫角色
		user.Delete("/u/{uid: string regexp([0-9a-f]) max(24)}", DeleteChar)          //刪除資料庫角色

	})
	app.PartyFunc("/admin", func(user iris.Party) {
		user.Use(authentication)
		user.Use(adminOnly)
		user.Layout("admin/layout.html")

		user.Get("/", GetAdminIndex)
		user.Get("/game_version", GetGameVersion)
		user.Get("/achievements", GetAchievements)
		user.Get("/skills", GetSkills)

		user.Get("/pets", GetPets)
		user.Get("/collections", GetCollections)
		user.Get("/events", GetEvents)
		user.Get("/stories", GetStories)

		user.Get("/titles", GetTitles)                                             //列出稱號清單
		user.Get("/titles/create", GetTitleCreate)                                 //新增稱號表單
		user.Post("/titles/create", PostTitleCreate)                               //新增稱號資料
		user.Get("/titles/{titleid: int}/edit", GetTitleEdit)                      //編輯稱號表單
		user.Put("/titles/{_id: string regexp([0-9a-f]) max(24)}", PutTitleUpdate) //更新稱號資料
		user.Delete("/titles/{_id: string regexp([0-9a-f]) max(24)}", DelTitle)    //刪除稱號資料

		user.Get("/talentmasters", GetTalentMasters)                                             //列出才能清單
		user.Get("/talentmasters/create", GetTalentMasterCreate)                                 //新增才能表單
		user.Post("/talentmasters/create", PostTalentMasterCreate)                               //新增才能資料
		user.Get("/talentmasters/{titleid: int}/edit", GetTalentMasterEdit)                      //編輯才能表單
		user.Put("/talentmasters/{_id: string regexp([0-9a-f]) max(24)}", PutTalentMasterUpdate) //更新才能資料
		user.Delete("/talentmasters/{_id: string regexp([0-9a-f]) max(24)}", DelTalentMaster)    //刪除才能資料

		user.Get("/servers", GetServers)                                             //列出伺服器清單
		user.Get("/servers/create", GetServerCreate)                                 //新增伺服器表單
		user.Post("/servers/create", PostServerCreate)                               //新增伺服器資料
		user.Get("/servers/{serverid: int}/edit", GetServerEdit)                     //編輯伺服器表單
		user.Put("/servers/{_id: string regexp([0-9a-f]) max(24)}", PutServerUpdate) //更新伺服器資料
		user.Delete("/servers/{_id: string regexp([0-9a-f]) max(24)}", DelServer)    //刪除伺服器資料
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
			session := sessions.Get(ctx)
			ctx.ViewData("message", session.GetFlashString("msg"))
			if err := ctx.View("register.html"); err != nil {
				ctx.Application().Logger().Infof(err.Error())
			}
		})
		guest.Post("/register", func(ctx iris.Context) {
			session := sessions.Get(ctx)
			/*
			   inputUsername
			   inputEmail
			   inputPassword
			*/
			username := ctx.PostValue("inputUsername")
			email := ctx.PostValue("inputEmail")
			password := ctx.PostValue("inputPassword")

			if len(username) < 4 {
				session.SetFlash("msg", "帳號長度不足，請重新輸入")
				ctx.Redirect("/register")
				return
			}
			if len(email) < 10 {
				session.SetFlash("msg", "Email長度不足，請重新輸入")
				ctx.Redirect("/register")
				return
			}
			if len(password) < 8 {
				session.SetFlash("msg", "密碼長度不足，請重新輸入")
				ctx.Redirect("/register")
				return
			}
			coll := DBSource.db.Collection("users")

			hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

			if err != nil {
				//fmt.Println(err)
				session.SetFlash("msg", "Password hash err: "+err.Error())
			} else {
				insertData := bson.M{
					"username":         username,     //Username
					"password":         string(hash), //hash
					"enabled":          true,         //有效帳號
					"role":             2,            //帳號角色
					"create_timestamp": primitive.Timestamp{T: uint32(time.Now().Unix())},
					"modify_timestamp": primitive.Timestamp{T: uint32(time.Now().Unix())},
				}

				//err := coll.FindOne(context.TODO(), filter).Decode(&result)
				insertResult, err := coll.InsertOne(context.TODO(), insertData)
				if err != nil {
					//log.Fatal(err)
					session.SetFlash("msg", "發生錯誤：可能原因為重複帳號.")
					ctx.Redirect("/register")
					return
				}
				log.Println("Added a new server with objectID: ", insertResult.InsertedID)
				session.SetFlash("msg", "已成功建立帳號")
			}
			ctx.Redirect("/user")
		})
		guest.Get("/login", func(ctx iris.Context) {
			session := sessions.Get(ctx)
			if auth, _ := session.GetBoolean("authenticated"); auth {
				// ctx.Redirect("/user")
				if session.Get("role") == "Admin" {
					ctx.Redirect("/admin")
				} else {
					ctx.Redirect("/user")
				}
			} else {
				ctx.ViewData("message", session.GetFlashString("msg"))
				ctx.View("login.html")
			}
		})
		guest.Post("/login", func(ctx iris.Context) {
			session := sessions.Get(ctx)

			//原始簡單版
			filter := bson.M{
				"username": ctx.PostValue("username"),
				"password": "",
				"enabled":  true,
			}

			result, err := APIQueryOneBase("users_role", filter)

			if err != nil {
				if err == mongo.ErrNoDocuments {
					session.SetFlash("msg", "無法登入，請確認您的帳號密碼是否正確")
				} else {
					session.SetFlash("msg", "無法登入，原因:"+err.Error())
				}
				ctx.Redirect("/login")
			} else {
				session.Set("authenticated", true)
				session.Set("username", result["username"])
				session.Set("_id", result["_id"])
				session.Set("role", result["rolename"])

				if result["rolename"] == "Admin" {
					ctx.Redirect("/admin")
				} else {
					ctx.Redirect("/user")
				}
			}
		})

		//上線後須改回POST
		guest.Get("/logout", func(ctx iris.Context) {
			session := sessions.Get(ctx)
			if auth, _ := session.GetBoolean("authenticated"); auth {
				session.Clear()
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
	}
	ctx.Next()
}

//RunningLog  記錄使用者瀏覽路徑
func RunningLog(ctx iris.Context) {
	ctx.Application().Logger().Infof("Runs method:%s %s", ctx.Method(), ctx.Path())
	ctx.Next()
}

func notFound(ctx iris.Context) {
	ctx.View("errors/404.html") //設定找不到頁面
}
