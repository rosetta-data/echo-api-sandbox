/*
Find Wi-Fi API main.go
*/
package main

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"gopkg.in/go-playground/validator.v9"

	"./handler"
	adminhandler "./handler/admin"
	clienthandler "./handler/client"
	"./model"
)

/*
Validator | バリデーターの構造体
*/
type Validator struct {
	validator *validator.Validate
}

/*
Validate | バリデーターのセットアップ
*/
func (v *Validator) Validate(i interface{}) error {
	return v.validator.Struct(i)
}

/*
ConnectGorm | Gormの接続
*/
func ConnectGorm() *gorm.DB {
	DBMS := "mysql"
	USER := os.Getenv("MYSQL_USER")
	PASS := os.Getenv("MYSQL_PASSWORD")
	PROTOCOL := "tcp(" + os.Getenv("MYSQL_HOST") + ":" + os.Getenv("MYSQL_PORT") + ")"
	DBNAME := os.Getenv("MYSQL_DB")
	OPTION := "charset=utf8mb4&loc=Asia%2FTokyo&parseTime=true"

	CONNECT := USER + ":" + PASS + "@" + PROTOCOL + "/" + DBNAME + "?" + OPTION
	db, err := gorm.Open(DBMS, CONNECT)

	if err != nil {
		panic(err.Error())
	}

	return db
}

/*
Migrate | DBの構築
*/
func Migrate(db *gorm.DB) {
	db.AutoMigrate(&model.Area{})
	db.AutoMigrate(&model.Service{})
	db.AutoMigrate(&model.Shop{}).AddForeignKey("service_id", "services(id)", "RESTRICT", "RESTRICT")
	db.AutoMigrate(&model.Review{}).AddForeignKey("shop_id", "shops(id)", "RESTRICT", "RESTRICT")
}

/*
Router | ルーティング
*/
func Router(e *echo.Echo, db *gorm.DB) {
	// API仕様書の出力
	e.File("/doc", "app/redoc.html")

	// ルーティング
	e.GET("/", handler.Hello())
	e.GET("/areas", clienthandler.GetAreaMasterClient(db))
	e.GET("/shops", clienthandler.GetShopListClient(db))
	e.POST("/admin/areas", adminhandler.RegisterAreaAdmin(db))
	e.DELETE("/admin/areas/:areaKey", adminhandler.DeleteAreaAdmin(db))
	e.GET("/admin/services", adminhandler.GetServiceListAdmin(db))
	e.POST("/admin/services", adminhandler.RegisterServiceAdmin(db))
	e.GET("/admin/shops", adminhandler.GetShopListAdmin(db))
	e.POST("/admin/shops", adminhandler.RegisterShopAdmin(db))
}

/*
Main
*/
func main() {
	e := echo.New()

	// バリデーターのセットアップ
	e.Validator = &Validator{validator: validator.New()}

	// DBのセットアップ
	db := ConnectGorm()
	defer db.Close()
	db.Set("gorm:table_options", "ENGINE=InnoDB CHARSET=utf8mb4")
	Migrate(db)

	// ルーティング
	Router(e, db)

	// リクエスト共通処理
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	e.Logger.Fatal(e.Start(":1323"))
}
