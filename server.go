package main

import (
	_ "fmt"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

type Human struct {
	Name string
	Age  int
}

var myaox = &Human{
	Name: "myaox",
	Age:  20,
}

func hello(c echo.Context) error {
	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		for {
			err := websocket.JSON.Send(ws, myaox)
			if err != nil {
				c.Logger().Error(err)
			}

			//err = websocket.JSON.Receive(ws, &myaox)
			//if err != nil {
			//	c.Logger().Error(err)
			//}
			//fmt.Printf("%v\n", myaox)
		}
	}).ServeHTTP(c.Response(), c.Request())

	return nil
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Static("/", "./public")
	e.GET("/ws", hello)

	e.Logger.Fatal(e.Start(":1323"))
}
