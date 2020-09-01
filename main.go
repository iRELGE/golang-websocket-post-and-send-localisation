package main

import (
	"context"
	"database/sql"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"rabie.com/testlocal/models"
	//"golang.org/x/net/websocket"
)

var (
	upgrader = websocket.Upgrader{}
)
var newlocalisation models.Localization
var lastlocalisation models.Localization

func connectdb() (*sql.DB, error) {
	//open db (regular sql open call)
	db, err := sql.Open("mysql", "root:@tcp(127.0.0.1:3306)/testLocalisation")
	if err != nil {
		return db, err
	}

	//close deferred
	return db, err

}
func postlocalisationdb(localisation models.Localization) error {
	db, err := connectdb()
	defer db.Close()
	if err != nil {
		return err
	}
	err = localisation.Insert(context.Background(), db, boil.Infer())

	return err

}
func postLocalisationapi(c echo.Context) error {
	defer c.Request().Body.Close()
	l := new(models.Localization)
	if err := c.Bind(l); err != nil {
		return c.String(http.StatusInternalServerError, "something wrong contact support")
	}
	err := postlocalisationdb(*l)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong contact support")
	}
	newlocalisation = *l
	return c.String(http.StatusOK, "waiting fir next localisation")

}

func getalllocalisaion() (models.Localization, error) {
	db, err := connectdb()
	l := new(models.Localization)
	defer db.Close()
	if err != nil {
		return *l, err
	}
	ls, err := models.Localizations().All(context.Background(), db)
	if err != nil {
		return *l, err
	}
	l = ls[len(ls)-1]
	return *l, err

}

func hello(c echo.Context) error {
	upgrader.CheckOrigin = func(*http.Request) bool { return true }
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	//var lastvalue models.Localization
	if err != nil {
		return err
	}
	defer ws.Close()
	newlocalisation, err = getalllocalisaion()
	if err != nil {
		ws.Close()
		c.Logger().Error(err)
	}

	for {
		// Write
		if lastlocalisation != newlocalisation {
			err = ws.WriteJSON(newlocalisation)
			lastlocalisation = newlocalisation
			if err != nil {
				ws.Close()
				c.Logger().Error(err)
			}
		}

		//Read

	}

}
func getTest(c echo.Context) error {
	return c.String(http.StatusOK, "hi")
}

func main() {
	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
		AllowMethods: []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
	}))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.POST("/po", postLocalisationapi)
	e.GET("/hi", getTest)
	e.GET("/ws", hello)
	// l, err := net.Listen("tcp", ":8080")
	// // e.Logger.Fatal(net.tcp(":1324"))
	// if err != nil {
	// 	e.Logger.Fatal(l)
	// }
	//e.Listener = l
	e.Logger.Fatal(e.Start(":1323"))
}
