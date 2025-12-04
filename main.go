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

const (
	DiscordRpcMin = 6463
	DiscordRpcMax = 6472
)

var (
	//go:embed icon.ico
	iconBytes []byte
	// Software Config
	config Config
	// Key watcher event multiplexer
	multi multiplexer
	// Http server
	httpServer *http.Server
	// Discord
	discordPort    = DiscordRpcMin
	enableActivity = false
	changeClient   = make(chan struct{})

	debug = flag.Bool("debug", false, "verbose log")
)

type Config struct {
	Listen   string          `json:"listen"`
	Discord  *DiscordConfig  `json:"discord,omitempty"`
	Activity *map[string]any `json:"activity,omitempty"`
}

type DiscordConfig struct {
	Id     string `json:"client_id"`
	Secret string `json:"client_secret"`
}

func init() {
	c, _ := os.ReadFile("./config.json")
	json.Unmarshal(c, &config)

	if config.Discord != nil && config.Activity != nil {
		enableActivity = true
		go discordActivity()
	}
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
	// Systray menu
	mExit := systray.AddMenuItem("Exit", "アプリケーションを終了します")
	mChange := systray.AddMenuItem("Select Next Discord Client (now:0)", "監視対象のdiscord Desktop Appを変更します")
	mChange.SetTitle(fmt.Sprintf("Select Next Discord Client (now:%d)", discordIPC))
	go func() {
		for {
			select {
			case <-mExit.ClickedCh:
				log.Println("Cliked menu: Exit")
				systray.Quit()

			case <-mChange.ClickedCh:
				log.Println("Cliked menu: Select Next Discord Client")
				discordIPC++
				if discordIPC > DiscordIpcMax {
					discordIPC = DiscordIpcMin
				}

				if enableActivity {
					go func() {
						changeClient <- struct{}{}
					}()
				}
				mChange.SetTitle(fmt.Sprintf("Select Next Discord Client (now:%d)", discordIPC))
			}
		}
	}()

	// Boot key watcher
	log.Println("systray.onReady().watcher")
	go newWatcher()

	// Boot http server
	log.Println("systray.onReady().http")
	go runHttpServer()
}

func onExit() {
	log.Println("onExit(): Initiating graceful shutdown.")

	// グレースフルシャットダウンの実行
	if httpServer != nil {
		// 5秒の猶予期間を持つコンテキストを作成
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// サーバーのシャットダウンを試みる
		if err := httpServer.Shutdown(ctx); err != nil {
			// シャットダウン中にエラー（タイムアウトを含む）
			log.Printf("HTTP server Shutdown error: %v", err)
		}
	}

	// 最後にプロセスを終了
	log.Println("Exiting application process.")
	os.Exit(0)
}
