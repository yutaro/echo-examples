package main

import (
	"fmt"
	_ "time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"golang.org/x/net/websocket"
)

type client struct {
	socket *websocket.Conn
	send   chan *comment
	room   *room
}

type room struct {
	id      string
	forward chan *comment
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

type comment struct {
	Author string
	Val    string
	Replys []comment
}

func (r *room) run() {
L:
	for {
		select {
		case client := <-r.join:
			r.clients[client] = true
		case client := <-r.leave:
			delete(r.clients, client)
			close(client.send)

		case msg := <-r.forward:
			for client := range r.clients {
				client.send <- msg
			}
		default:
			if len(r.clients) == 0 {
				delete(rooms, r.id)
				close(r.forward)
				close(r.join)
				close(r.leave)

				fmt.Println("server closed")

				break L
			}
		}
	}
}

var rooms = make(map[string]*room)

func chatws(c echo.Context) error {
	id := c.Param("id")

	websocket.Handler(func(ws *websocket.Conn) {
		defer ws.Close()
		user := &client{
			socket: ws,
			send:   make(chan *comment),
			room:   nil,
		}

		defer func() {
			user.room.leave <- user
		}()

		if foundRoom, ok := rooms[id]; ok {
			user.room = foundRoom
			foundRoom.join <- user
		} else {
			newRoom := &room{
				id:      id,
				forward: make(chan *comment),
				join:    make(chan *client),
				leave:   make(chan *client),
				clients: make(map[*client]bool),
			}
			user.room = newRoom
			rooms[id] = newRoom

			go newRoom.run()
			newRoom.join <- user
		}
	L:
		for {
			select {
			case comment := <-user.send:
				err := websocket.JSON.Send(ws, comment)

				if err != nil {
					break L
				}

			default:
				cmt := new(comment)
				err := websocket.JSON.Receive(ws, cmt)
				if err != nil {
					break L
				}

				if cmt != nil {
					fmt.Println(cmt)
					user.room.forward <- cmt
				}
			}
		}

	}).ServeHTTP(c.Response(), c.Request())
	return nil
}

func main() {
	e := echo.New()

	e.Use(middleware.Logger())
	//e.Use(middleware.Recover())

	e.File("/chat/*", "./public/index.html")
	e.GET("/chatws/:id", chatws)

	e.Logger.Fatal(e.Start(":1323"))
}
