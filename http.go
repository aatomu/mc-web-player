package main

import (
	"encoding/json"
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
	mux.HandleFunc("/env", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(map[string]any{
			"client_id":     config.Discord.Id,
			"client_secret": config.Discord.Secret,
		})
	})
	mux.Handle("/", fileServer)

	httpServer = &http.Server{
		Addr:         config.Listen,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed: %v", err)
	}
	log.Println("runServer(): HTTP server stopped.")
}

func handleDiscordSocket(ws *websocket.Conn) {
	defer ws.Close()

	ipc, err := NewIPC(config.Discord.Id, discordIPC)
	if err != nil {
		log.Println("handleDiscordSocket()/connect ipc failed:", err)
		return
	}

	go func() {
		for {
			rpcRes, err := ipc.ReadJSON(nil)
			if rpcRes.Code == Pong {
				continue
			}
			if err != nil {
				return
			}

			_, err = ws.Write(rpcRes.Message)
			if err != nil {
				return
			}
		}
	}()

	var rpc DiscordPayload
	for {
		err := websocket.JSON.Receive(ws, &rpc)
		if err != nil {
			return
		}

		err = ipc.WriteJSON(Frame, rpc)
		if err != nil {
			return
		}
	}
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
