package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func main() {
	listener, _ := net.Listen("tcp", "127.0.0.1:0")
	defer listener.Close()

	addr := listener.Addr().String()
	log.Println("Listen:", addr)

	mux := http.NewServeMux()
	mux.Handle("/ws", websocket.Handler(handleWebsocket))

	server := &http.Server{
		ReadTimeout: 5 * time.Second,
		Handler:     mux,
	}
	server.Serve(listener)
}

func handleWebsocket(ws *websocket.Conn) {
	var e = ""
	capture := getWatcher()
	for {
		select {
		case m := <-capture.mouse.Vscroll:
			e = fmt.Sprintf("mouse,Vscroll,%d", m)
		case m := <-capture.mouse.Hscroll:
			e = fmt.Sprintf("mouse,Hscroll,%d", m)
		case m := <-capture.mouse.up:
			e = fmt.Sprintf("mouse,up,%s", m)
		case m := <-capture.mouse.down:
			e = fmt.Sprintf("mouse,down,%s", m)
		case m := <-capture.mouse.move:
			e = fmt.Sprintf("mouse,move,[%d,%d]", m[0], m[1])
		case k := <-capture.keyboard.down:
			e = fmt.Sprintf("keyboard,down,%s", k)
		case k := <-capture.keyboard.up:
			e = fmt.Sprintf("keyboard,up,%s", k)
		}
		log.Println(e)
		websocket.Message.Send(ws, e)
	}
}
