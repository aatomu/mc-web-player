package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

func runHttpServer() {
	log.Println("runServer()")

	fileServer := http.FileServer(http.Dir("./assets"))

	mux := http.NewServeMux()
	mux.Handle("/ws/key", websocket.Handler(handleKeySocket))
	mux.Handle("/ws/discord", websocket.Handler(handleDiscordSocket))
	mux.Handle("/", fileServer)

	hs = &http.Server{
		Addr:         config.Listen,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := hs.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
	log.Println("runServer(): HTTP server stopped.")
}

func handleDiscordSocket(ws *websocket.Conn) {
}

func handleKeySocket(ws *websocket.Conn) {
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
