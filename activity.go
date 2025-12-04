package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func discordActivity() {
	var token, reflesh string

	if _, err := os.Stat("./cache"); err == nil {
		b, _ := os.ReadFile("./cache")
		p := strings.Split(string(b), "\n")
		if len(p) == 2 {
			token = p[0]
			reflesh = p[1]
			log.Println(reflesh)
		}
	}

	for {
		func() {
			ipc, err := NewIPC(config.Discord.Id, discordIPC)
			if err != nil {
				log.Println("Discord IPC connection failed:", err)
				time.Sleep(5 * time.Second)
				return
			}
			defer func() {
				log.Println("Discord IPC close called.")
				err := ipc.Close()
				if err != nil {
					log.Println("Discord IPC close failed:", err)
				}
			}()

			rpcChan := make(chan DiscordPayload)
			errChan := make(chan error)
			go func() {
				for {
					var rpc DiscordPayload
					_, err := ipc.ReadJSON(&rpc)
					if err != nil {
						errChan <- err
						return
					}
					rpcChan <- rpc
				}
			}()

			for {
				select {
				case rpc := <-rpcChan:
					switch rpc.Cmd {
					case DISPATCH:
						switch rpc.Evt {
						case "READY":
							if token == "" {
								ipc.WriteJSON(Frame, makeAuthorize(fmt.Sprintf("AUTHORIZE_NEW-%d", ipc.packetId)))
								continue
							} else {
								ipc.WriteJSON(Frame, makeAuthenticate(fmt.Sprintf("AUTHENTICATE_CACHE-%d", ipc.packetId), token))
							}
						default:
							log.Printf("Unknown Event: %#v", rpc)
						}

					case AUTHORIZE:
						if rpc.Evt == "ERROR" {
							return
						}

						token, reflesh, err = convertCodeToToken(rpc.Data["code"].(string))
						if err != nil {
							log.Println("autorize failed:", err)
							return
						}
						WriteTokenCache(token, reflesh)
						ipc.WriteJSON(Frame, makeAuthenticate(fmt.Sprintf("AUTHENTICATE_NEW-%d", ipc.packetId), token))

					case AUTHENTICATE:
						if rpc.Evt == "ERROR" {
							token, reflesh, err = refreshAccessToken(reflesh)
							if err != nil {
								log.Println("reflesh failed:", err)
								return
							}
							WriteTokenCache(token, reflesh)
							ipc.WriteJSON(Frame, makeAuthenticate(fmt.Sprintf("AUTHENTICATE_REFLESH-%d", ipc.packetId), token))
						}

						ipc.WriteJSON(Frame, DiscordPayload{
							Nonce: "AUTOMATION",
							Cmd:   SET_ACTIVITY,
							Args: map[string]any{
								"pid":      1,
								"activity": config.Activity,
							},
						})
					}
				case <-errChan:
					log.Println("discordActivity() read error:", err)
					return
				case <-changeClient:
					return
				}
			}
		}()
	}
}

func makeAuthorize(nonce string) DiscordPayload {
	return DiscordPayload{
		Nonce: nonce,
		Cmd:   AUTHORIZE,
		Args: map[string]any{
			"client_id": config.Discord.Id,
			"scopes":    []string{"identify", "rpc"},
		},
	}
}

func makeAuthenticate(nonce, token string) DiscordPayload {
	return DiscordPayload{
		Nonce: nonce,
		Cmd:   AUTHENTICATE,
		Args: map[string]any{
			"access_token": token,
		},
	}
}

func convertCodeToToken(code string) (token, reflesh string, err error) {
	payload := map[string]string{
		"client_id":     config.Discord.Id,
		"client_secret": config.Discord.Secret,
		"grant_type":    "authorization_code",
		"code":          code,
		"redirect_uri":  "http://127.0.0.1",
	}

	formData := new(bytes.Buffer)
	// mapをURLエンコードされたフォームデータ形式に変換（application/x-www-form-urlencoded）
	for key, value := range payload {
		if formData.Len() > 0 {
			formData.WriteString("&")
		}
		formData.WriteString(fmt.Sprintf("%s=%s", key, value))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", DiscordTokenEndpoint, formData)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send request to Discord API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return "", "", fmt.Errorf("discord API returned error status %d: %v", resp.StatusCode, errorBody)
	}

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", "", fmt.Errorf("failed to decode access token response: %w", err)
	}

	token = tokenRes.AccessToken
	reflesh = tokenRes.RefreshToken
	return
}

func refreshAccessToken(refreshToken string) (token, reflesh string, err error) {
	// 1. POSTリクエストのペイロード（フォームデータ）を作成
	payload := map[string]string{
		"client_id":     config.Discord.Id,
		"client_secret": config.Discord.Secret,
		"grant_type":    "refresh_token", // リフレッシュトークンフローを指定
		"refresh_token": refreshToken,
	}

	formData := new(bytes.Buffer)
	// mapをURLエンコードされたフォームデータ形式に変換（application/x-www-form-urlencoded）
	for key, value := range payload {
		if formData.Len() > 0 {
			formData.WriteString("&")
		}
		formData.WriteString(fmt.Sprintf("%s=%s", key, value))
	}

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", DiscordTokenEndpoint, formData)
	if err != nil {
		return "", "", fmt.Errorf("failed to create refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to send refresh request to Discord API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		return "", "", fmt.Errorf("discord API returned error status %d during refresh: %v", resp.StatusCode, errorBody)
	}

	var tokenRes struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		ExpiresIn    int    `json:"expires_in"`
		RefreshToken string `json:"refresh_token"`
		Scope        string `json:"scope"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", "", fmt.Errorf("failed to decode access token response: %w", err)
	}

	token = tokenRes.AccessToken
	reflesh = tokenRes.RefreshToken
	return
}

func WriteTokenCache(token, reflesh string) error {
	return os.WriteFile("./cache", []byte(fmt.Sprintf("%s\n%s", token, reflesh)), 0755)
}
