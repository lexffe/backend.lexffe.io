package auth

import (
	"bytes"
	"context"
	"crypto/aes"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/patrickmn/go-cache"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
    "totp_key": encrypted(string)
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

  Authorization: Bearer <username>:<hex-string>
  Security:
    path-based filtering (Cloudflare?)
  paths, methods:
    POST /login

  init: kickstart otp, populate
*/

type authenticationBody struct {
	OTP string `json:"otp" bson:"otp"`
}

type authenticationModel struct {
	ID           primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	OTPEncrypted []byte             `json:"otp_crypt" bson:"otp_crypt"`
}

// OTPInitialization initialises the OTP key for admin access
func OTPInitialization(ctx context.Context, encryptionKey string, db *mongo.Database) error {

	// check if database has existing OTP code. there should only be one

	res := db.Collection("otp").FindOne(ctx, bson.M{})

	if res.Err() != nil {

		// if doesn't exist, generate new one, write into database

		if res.Err() == mongo.ErrNoDocuments {

			// generate totp key

			totpKey, err := totp.Generate(totp.GenerateOpts{
				Issuer:      "backend",
				AccountName: "admin@backend",
				Digits:      otp.DigitsEight,
				Algorithm:   otp.AlgorithmSHA512,
			})

			totpKeyBuf := bytes.NewBufferString(totpKey.String())

			// initialise cipher

			encryptionKeyBuf := bytes.NewBufferString(encryptionKey)
			cipher, err := aes.NewCipher(encryptionKeyBuf.Bytes())

			if err != nil {
				return err
			}

			// encrypt

			var authModelInstance authenticationModel

			cipher.Encrypt(authModelInstance.OTPEncrypted, totpKeyBuf.Bytes())

			// insert model into database

			if _, err := db.Collection("otp").InsertOne(ctx, authModelInstance); err != nil {
				return err
			}

			return nil

		}

		return res.Err()

	}

	return nil
}

// AuthenticateHandler checks the otp key against the database, then returns a temporary api key valid for 2 hours
func AuthenticateHandler(ctx *gin.Context) {

	// Parse body: username, otp

	var authInfo authenticationBody

	if err := ctx.BindJSON(&authInfo); err != nil {
		ctx.Error(errors.New("invalid authentication details"))
		ctx.String(400, "invalid authentication details")
		return
	}

	// look for user's otp key in the db

	db := ctx.MustGet("db").(*mongo.Database)
	coll := db.Collection("otp")

	encryptionKey := ctx.MustGet("otp_crypt").(string)

	res := coll.FindOne(ctx, bson.M{})

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
			ctx.Error(res.Err())
			return
		}
	}

	var authData authenticationModel
	var otpKey bytes.Buffer
	var otpBytes []byte

	if err := res.Decode(&authData); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	// decrypt the otp key

	keybuf := bytes.NewBufferString(encryptionKey)
	cipher, err := aes.NewCipher(keybuf.Bytes())

	if err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	cipher.Decrypt(otpBytes, authData.OTPEncrypted)
	if _, err := otpKey.Write(otpBytes); err != nil {
		ctx.Status(http.StatusInternalServerError)
		ctx.Error(err)
		return
	}

	validated := totp.Validate(authInfo.OTP, otpKey.String())

	if validated == true { // authenticated.
		// generate apikey subroutine

		keycache := ctx.MustGet("keycache").(*cache.Cache)
		apiKey, err := helpers.HexStringGen(5)

		if err != nil {
      ctx.Status(http.StatusInternalServerError)
      ctx.Error(err)
      return
		}

		val, exists := keycache.Get("keys")

		if exists != true {
			keycache.Add("keys", []string{apiKey}, cache.DefaultExpiration)
		} else {
			var nval = append(val.([]string), apiKey)
			keycache.Set("keys", nval, cache.DefaultExpiration)
    }
    
    ctx.String(http.StatusOK, apiKey)
    return
  }
  
  ctx.Status(http.StatusUnauthorized)
  return
}

// BearerMiddleware -
func BearerMiddleware(ctx *gin.Context) {

  key := strings.Split(ctx.GetHeader("Authorization"), " ") // Bearer aabbccddeeff

  keycache := ctx.MustGet("keycache").(*cache.Cache)

  val, exists := keycache.Get("keys")

  if exists != true {
    ctx.Status(http.StatusUnauthorized)
    return
  }

  var found = false

  for i := range val.([]string) {
    if val.([]string)[i] == key[1] {
      found = true
    }
  }

  if (found == true) {
    ctx.Set("Authenticated", true)
    ctx.Next()
    return
  }

  ctx.Status(http.StatusUnauthorized)
  return
}
