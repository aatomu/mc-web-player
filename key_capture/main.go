package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	hook "github.com/robotn/gohook"
	"golang.org/x/net/websocket"
)

var (
	multi multiplexer
)

const (
	MouseLeft uint16 = iota + 1
	MouseRight
	MouseMiddle
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
		evCh := hook.Start()
		defer hook.End()

		var buf [3]string
		for ev := range evCh {
			switch ev.Kind {
			case hook.KeyDown:
				buf = [3]string{"keyboard", "down", hook.RawcodetoKeychar(ev.Rawcode)}
			case hook.KeyUp:
				buf = [3]string{"keyboard", "up", hook.RawcodetoKeychar(ev.Rawcode)}
			case hook.MouseDown:
				buf = [3]string{"mouse", "down", ""}
				switch ev.Button {
				case MouseLeft:
					buf[2] = "Left"
				case MouseMiddle:
					buf[2] = "Middle"
				case MouseRight:
					buf[2] = "Right"
				}
			case hook.MouseUp:
				buf = [3]string{"mouse", "up", ""}
				switch ev.Button {
				case MouseLeft:
					buf[2] = "Left"
				case MouseMiddle:
					buf[2] = "Middle"
				case MouseRight:
					buf[2] = "Right"
				}
			case hook.MouseMove:
				buf = [3]string{"mouse", "move", fmt.Sprintf("[%d,%d]", ev.X, ev.Y)}
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
	ch := make(chan [3]string, 10)
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
