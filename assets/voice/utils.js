// @ts-check

function newError(err) {
  const MESSAGE = document.createElement("div")
  MESSAGE.innerHTML = err

  const ERRORS = document.getElementById("errors")
  if (ERRORS) {
    ERRORS.append(MESSAGE)
  }
}

/**
 * @param {string} key get cookie key
 * @return {string | null}
 */
function getCookie(key) {
  const cookies = document.cookie
  const parts = cookies.split(";")
  for (let i = 0; i < parts.length; i++) {
    const cookie = parts[i]
    const content = cookie.split("=", 2)
    content[0] = content[0].replace(/^ +/, "")
    console.log(content)
    console.log(content[0], content[0] === key)
    if (content[0] === key) {
      return content[1]
    }
  }
  return null
}

/**
 * @param {string} key
 * @param {string} value
 */
function setCookie(key, value, ops="") {
  var t = `${key}=${value}`
  if (!ops) {
    t += `; ${ops}`
  }
  document.cookie = t
}