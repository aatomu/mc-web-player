package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"fyne.io/systray"
)

var (
	//go:embed icon.ico
	iconBytes []byte
	// Software Config
	config Config
	// Key watcher event multiplexer
	multi multiplexer
	// Http server
	hs *http.Server

	debug = flag.Bool("debug", false, "verbose log")
)

type Config struct {
	Listen string `json:"listen"`
}

func init() {
	c, _ := os.ReadFile("./config.json")
	json.Unmarshal(c, &config)
}

func main() {
	flag.Parse()

	systray.Run(onReady, onExit)
}

func onReady() {
	// Systray setting
	log.Println("systray.onReady().systray")
	systray.SetIcon(iconBytes)
	systray.SetTitle("Streamer Kit")
	systray.SetTooltip(fmt.Sprintf("=== Stream Kit ===\nListening: %s", config.Listen))

	// Boot key watcher
	log.Println("systray.onReady().watcher")
	multi = newWatcher()

	// Boot http server
	log.Println("systray.onReady().http")
	go runHttpServer()
}

func onExit() {
	log.Println("onExit(): Initiating graceful shutdown.")

	// グレースフルシャットダウンの実行
	if hs != nil {
		// 5秒の猶予期間を持つコンテキストを作成
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// サーバーのシャットダウンを試みる
		if err := hs.Shutdown(ctx); err != nil {
			// シャットダウン中にエラー（タイムアウトを含む）
			log.Printf("HTTP server Shutdown error: %v", err)
		}
	}

	// 最後にプロセスを終了
	log.Println("Exiting application process.")
	os.Exit(0)
}
