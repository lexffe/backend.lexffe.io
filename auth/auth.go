package auth

import (
	"bytes"
	"errors"
	"strings"

	"crypto/aes"

	"github.com/gin-gonic/gin"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/mongo"
)

/*

authentication flow:

> Admin tries to login:

creds: username, totp

> Server authenticates totp

> Server returns a generated hex apikey as access token

token expires in a set time (60 min?)

> server

mongo storage:

{
  "username": string,
  "totp_key": encrypted(string)
}

in-memory kv:

{
  "username": string,
  "active_api_keys": []{
    key: string,
    expired: boolean
  }
}

> when server generates an api key and stores in in-mem kv, initiate goroutine

Authorization: Bearer <username>:<hex-string>

Security:

path-based filtering (Cloudflare?)

paths:

POST /login
*/

type AuthenticationBody struct {
  Username string `json:"username"`
  OTP uint `json:"otp"`
}

func OTPInitialization(ctx *gin.Context, encryptionKey string) error {
  
  // check if embedded library has existing OTP code

  db := ctx.MustGet("db").(*mongo.Database)

  // if it exists, decrypt and load into ctx

  // if doesn't exist, generate new one, write into database, and load it into ctx

  keybuf := bytes.NewBufferString(encryptionKey)
  _, err := aes.NewCipher(keybuf.Bytes())

  if err != nil {
    return err
  }

  return nil

}


func AuthenticateHandler(ctx *gin.Context) {
  
  var authInfo AuthenticationBody

  if err := ctx.BindJSON(&authInfo); err != nil {
    ctx.Error(errors.New("invalid authentication details"))
    ctx.String(400, "invalid authentication details")
    return
  }

  // if username == w

}

func BearerMiddleware(ctx *gin.Context) {

  ctx.Set("Authenticated", false)

  ctx.GetHeader("Authorization")

  strings.Split("", ":")

  // if authentciated,
  // ctx.Set("Authenticated", true)
  
  // if not authenticated, 401.

  ctx.Next()
}
