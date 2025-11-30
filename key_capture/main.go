package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

var (
	multi multiplexer
)

type multiplexer struct {
	i   int
	sub map[int]chan [3]string
	mu  sync.Mutex
}

func main() {
	multi = multiplexer{
		i:   0,
		sub: map[int]chan [3]string{},
	}

	go func() {
		time.Sleep(1000 * time.Millisecond)
		capture := getWatcher()
		var buf [3]string
		for {
			select {
			case m := <-capture.mouse.Vscroll:
				buf = [3]string{"mouse", "Vscroll", fmt.Sprintf("%d", m)}
			case m := <-capture.mouse.Hscroll:
				buf = [3]string{"mouse", "Hscroll", fmt.Sprintf("%d", m)}
			case m := <-capture.mouse.down:
				buf = [3]string{"mouse", "down", m}
			case m := <-capture.mouse.up:
				buf = [3]string{"mouse", "up", m}
			case m := <-capture.mouse.move:
				buf = [3]string{"mouse", "move", fmt.Sprintf("[%d,%d]", m[0], m[1])}
			case k := <-capture.keyboard.down:
				buf = [3]string{"keyboard", "down", k}
			case k := <-capture.keyboard.up:
				buf = [3]string{"keyboard", "up", k}
			}
			log.Println(buf)

			for _, s := range multi.sub {
				select {
				case s <- buf:
				default:
				}
			}
		}
	}()

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
	multi.mu.Lock()
	id := multi.i
	multi.i++
	ch := make(chan [3]string)
	multi.sub[id] = ch
	multi.mu.Unlock()

	defer func() {
		ws.Close()

		multi.mu.Lock()
		delete(multi.sub, id)
		multi.mu.Unlock()
	}()

	for e := range ch {
		websocket.Message.Send(ws, fmt.Sprintf("%s,%s,%s", e[0], e[1], e[2]))
	}
}
