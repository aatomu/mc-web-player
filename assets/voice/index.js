// @ts-check
/// <reference path="index.ts"/>
/// <reference path="utils.js"/>
/// <reference path="voice.js"/>

let localUserID

const nonce = new class Nonce {
  num
  constructor() {
    this.num = 0
  }
  get() {
    this.num++
    return this.num.toString(16).padStart(4, "0")
  }
}

/**
 * 
 * @param {{readonly client_id:string, readonly client_secret:string}} config 
 */
function connect2Discord(config) {
  // MARK: websocket initialize
  const url = new URL(window.location.origin)
  url.pathname = "/ws/discord"
  const WEBSOCKET = new WebSocket(url)

  let currentVoiceChannelId = ""
  /**
 * @param {DiscordPayloadCmd} command
 * @param {DiscordPayloadEvt|""} event
 * @param {object} argumentsObject
 */
  function sendMessage(command, event, argumentsObject) {
    WEBSOCKET.send(JSON.stringify({
      nonce: nonce.get(),
      cmd: command,
      evt: event,
      args: argumentsObject
    }))
  }

  WEBSOCKET.addEventListener("open", function (event) {
    console.log("Open", event)
  })
  WEBSOCKET.addEventListener("message", async function (event) {
    if (!event.data) {
      return
    }

    /** @type {DiscordPayload} */
    const RPC = JSON.parse(event.data)
    console.log("Message", RPC)


    // Receive Event
    switch (RPC.cmd) {
      case "DISPATCH": {
        switch (RPC.evt) {
          // MARK: DISPATCH/READY
          case "READY": {
            const token = getCookie("token")
            if (!token) {
              console.log("cached token missing")
              sendMessage("AUTHORIZE", "", { "client_id": getCookie("client_id"), "scopes": ["identify", "rpc"] })
            } else {
              console.log("cached token found")
              sendMessage("AUTHENTICATE", "", { access_token: getCookie("token") })  
            }
          }
            return

          // Voice events
          case "VOICE_STATE_CREATE": {
            const STATE = RPC.data
            userAdd(STATE.user)
            userSort()
            return
          }
          case "VOICE_STATE_UPDATE": {
            const STATE = RPC.data
            userUpdate(STATE.nick, STATE.user, STATE.voice_state)
            userSort()
            return
          }
          case "VOICE_STATE_DELETE": {
            const USER_ID = RPC.data.user.id
            const USER = document.getElementById(USER_ID)
            if (USER) {
              USER.remove()
            }

            // if me
            if (USER_ID == localUserID) {
              sendMessage("UNSUBSCRIBE", "VOICE_STATE_CREATE", { channel_id: currentVoiceChannelId }) // Connect
              sendMessage("UNSUBSCRIBE", "VOICE_STATE_UPDATE", { channel_id: currentVoiceChannelId }) // Change VC state
              sendMessage("UNSUBSCRIBE", "VOICE_STATE_DELETE", { channel_id: currentVoiceChannelId }) // Disconnect
              sendMessage("UNSUBSCRIBE", "SPEAKING_START", { channel_id: currentVoiceChannelId }) // Speak start
              sendMessage("UNSUBSCRIBE", "SPEAKING_STOP", { channel_id: currentVoiceChannelId }) // Speak stop
              const USERS = document.getElementById("users")
              if (USERS) {
                while (USERS.firstChild) {
                  USERS.removeChild(USERS.firstChild);
                }
              }
              const CHANNEL_NAME = document.getElementById("channel")
              if (CHANNEL_NAME) {
                CHANNEL_NAME.innerText = ""
              }
              currentVoiceChannelId = ""
            }
            return
          }
          case "SPEAKING_START": {
            const USER_ID = RPC.data.user_id
            const USER = document.getElementById(USER_ID)
            if (USER) {
              USER.classList.add("speaking")
            }
            return
          }
          case "SPEAKING_STOP": {
            const USER_ID = RPC.data.user_id
            const USER = document.getElementById(USER_ID)
            if (USER) {
              USER.classList.remove("speaking")
            }
            return
          }
        }
        return
      }
      case "AUTHORIZE": {
        if (RPC.evt == "ERROR") {
          newError(`Discord App との認証に失敗しました: ${RPC.data.message}`)
          return
        }

        const OAUTH = await fetch("https://discordapp.com/api/oauth2/token", {
          method: "POST",
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Accept': 'application/json'
          },
          body: new URLSearchParams({
            client_id: config.client_id,
            client_secret: config.client_secret,
            grant_type: "authorization_code",
            code: RPC.data.code,
            redirect_uri: window.location.origin
          }).toString()
        }).then(res => {
          return res.json()
        })

        console.log("AUTHORIZE", OAUTH)
        sendMessage("AUTHENTICATE", "", { access_token: OAUTH.access_token })
        setCookie("token", OAUTH.access_token)
        setCookie("refresh_token", OAUTH.refresh_token)
        return
      }
      case "AUTHENTICATE": {
        if (RPC.evt == "ERROR") {
          const OAUTH = await fetch("https://discordapp.com/api/oauth2/token", {
            method: "POST",
            headers: {
              'Content-Type': 'application/x-www-form-urlencoded',
              'Accept': 'application/json'
            },
            body: new URLSearchParams({
              client_id: config.client_id,
              client_secret: config.client_secret,
              grant_type: "refresh_token",
              "refresh_token": getCookie("refresh_token") ?? "",
            }).toString()
          }).then(res => {
            return res.json()
          })
          sendMessage("AUTHENTICATE", "", { access_token: OAUTH.access_token })
          setCookie("token", OAUTH.access_token)
          setCookie("refresh_token", OAUTH.refresh_token)

          return
        }

        console.log("aaaa")
        localUserID = RPC.data.user.id

        setInterval(function () {
          if (currentVoiceChannelId == "") {
            console.log("Check New Channel!")
            sendMessage("GET_SELECTED_VOICE_CHANNEL", "", {}) // Get current channel
          }
        }, 500)

        return
      }
      // Voice events
      case "GET_SELECTED_VOICE_CHANNEL": {
        console.log("GET_SELECTED_VOICE_CHANNEL", RPC.data)
        if (!RPC.data || RPC.evt != null) {
          return
        }

        // Channel Name
        const CHANNEL_NAME = document.getElementById("channel")
        if (CHANNEL_NAME) {
          CHANNEL_NAME.innerText = RPC.data.name
        }
        // SUBSCRIBE
        const VOICE_CHANNEL_ID = RPC.data.id.toString()
        sendMessage("SUBSCRIBE", "VOICE_STATE_CREATE", { channel_id: VOICE_CHANNEL_ID }) // Connect
        sendMessage("SUBSCRIBE", "VOICE_STATE_UPDATE", { channel_id: VOICE_CHANNEL_ID }) // Change VC state
        sendMessage("SUBSCRIBE", "VOICE_STATE_DELETE", { channel_id: VOICE_CHANNEL_ID }) // Disconnect
        sendMessage("SUBSCRIBE", "SPEAKING_START", { channel_id: VOICE_CHANNEL_ID }) // Speak start
        sendMessage("SUBSCRIBE", "SPEAKING_STOP", { channel_id: VOICE_CHANNEL_ID }) // Speak stop
        currentVoiceChannelId = VOICE_CHANNEL_ID
        // Add users
        RPC.data.voice_states.forEach((state) => {
          userAdd(state.user)
          userUpdate(state.nick, state.user, state.voice_state)
        });
        userSort()
        return
      }
    }
  })

  WEBSOCKET.addEventListener("error", function (event) {
    console.log("Error", event)
    newError(`Stream Kitとの接続でエラーが発生しました: ${JSON.stringify(event)}`)
  })
  WEBSOCKET.addEventListener("close", function (event) {
    console.log("Close", event)
    newError(`Stream Kitとの接続が切断されました`)
  })
}

async function main() {
  const url = new URL(window.location.origin)
  url.pathname = "/env"
  const response = await fetch(url.href);
  if (!response.ok) {
    console.error(`設定の取得に失敗しました: HTTP Status ${response.status}`);
    return
  }

  /** @type {{client_id:string,client_secret:string}} */
  const data = await response.json();
  setCookie("client_id", data.client_id)
  setCookie("client_secret", data.client_secret)
  connect2Discord(data)
}

main()