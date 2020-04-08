package auth

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
  authentication flow:
  > Admin tries to login (username, totp)
  > Server authenticates totp
  > Server returns a generated hex apikey as access token
    - token expires in a set time (60 min?)
  > storage

  mongo storage:

  {
    "username": string,
    "totp_key": string
  }

  in-memory kv:

  {
    "username": string,
    "active_api_keys": []{
      key: string,
      expire_at: time
    }
  }

  > when server generates an api key and stores in in-mem kv, initiate goroutine

  Authorization: Bearer <hex-string>
  Security:
    path-based filtering (Cloudflare?)
  paths, methods:
    POST /login

  init: kickstart otp, populate
*/

const (
	collectionAuth = "auth"
	totpIssuer = "backend"
	totpAccountName = "admin@backend"
)

type authHandlerBody struct {
	OTPToken string `json:"otp_token" bson:"otp_token"`
}

type authDBModel struct {
	ID     primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	OTPKey string             `json:"otp_key" bson:"otp_key"`
}
