// @ts-check

function connectWebsocket() {
  const url = new URL(window.location.origin)
  url.pathname = "/ws/key"
  const ws = new WebSocket(url.href)

  ws.addEventListener("open", (e) => {
    console.log("open", e)
  })
  ws.addEventListener("message", (e) => {
    console.log("message", e)

    /** @type {string} */
    const message = e.data
    const sourceIndex = message.indexOf(",")
    const source = message.substring(0, sourceIndex)
    const methodIndex = message.indexOf(",", sourceIndex + 1)
    const method = message.substring(sourceIndex + 1, methodIndex)
    const value = message.substring(methodIndex + 1)

    switch (source) {
      case "keyboard": {
        switch (method) {
          case "down": {
            switch (value) {
              case "w": {
                animation_control.move.front = true
              }
              case "s": {
                animation_control.move.back = true
              }
              case "d": {
                animation_control.move.right = true
              }
              case "a": {
                animation_control.move.left = true
              }
            }
            return
          }
          case "up": {
            switch (value) {
              case "w": {
                animation_control.move.front = false
              }
              case "s": {
                animation_control.move.back = false
              }
              case "d": {
                animation_control.move.right = false
              }
              case "a": {
                animation_control.move.left = false
              }
            }
            return
          }
        }
      }
      case "mouse": {
        switch (method) {
          case "down": {
            console.log("Mdown", value)
            return
          }
          case "up": {
            console.log("Mup", value)
            return
          }
          case "move": {
            animation_control.facing.nowMousePos = JSON.parse(value)

            const diffX = animation_control.facing.nowMousePos[0] - animation_control.facing.prevMousePos[0]
            const diffY = animation_control.facing.nowMousePos[1] - animation_control.facing.prevMousePos[1]
            if (diffX < 0) {
              animation_control.facing.rotation[0] = Math.max(animation_control.facing.rotation[0] + (diffX / 20), -90)
            }
            if (diffX > 0) {
              animation_control.facing.rotation[0] = Math.min(animation_control.facing.rotation[0] + (diffX / 20), 90)
            }
            if (diffY > 0) {
              animation_control.facing.rotation[1] = Math.max(animation_control.facing.rotation[1] - (diffY / 10), -40)
            }
            if (diffY < 0) {
              animation_control.facing.rotation[1] = Math.min(animation_control.facing.rotation[1] - (diffY / 10), 40)
            }
            animation_control.facing.prevMousePos = animation_control.facing.nowMousePos
          }
        }
      }
    }
  })
  ws.addEventListener("error", (e) => {
    console.log("error", e)
  })
  ws.addEventListener("close", (e) => {
    console.log("close", e)
  })
}