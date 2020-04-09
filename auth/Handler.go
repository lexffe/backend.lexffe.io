package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/patrickmn/go-cache"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Handler checks the otp key against the database, then returns a temporary api key valid for 2 hours
func (s *AuthenticateHandler) Handler(ctx *gin.Context) {

	// Parse body: username, otp

	var authInfo authHandlerBody

	if err := ctx.BindJSON(&authInfo); err != nil {
		ctx.AbortWithError(http.StatusBadRequest, errors.New("invalid authentication details"))
		ctx.Error(err)
		return
	}

	// look for user's otp key in the db

	res := s.DB.Collection(s.Collection).FindOne(ctx.Request.Context(), bson.M{})

	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			ctx.Status(http.StatusNotFound)
		} else {
			ctx.Status(http.StatusInternalServerError)
		}
		ctx.Error(res.Err())
		return
	}

	var authData authDBModel

	if err := res.Decode(&authData); err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if validated := totp.Validate(authInfo.OTPToken, authData.OTPKey); validated == true { // authenticated.
		// generate apikey subroutine

		apiKey, err := helpers.HexStringGen(8)

		if err != nil {
			ctx.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		val, exists := s.Cache.Get("keys")

		if exists != true {
			s.Cache.Add("keys", []string{apiKey}, cache.DefaultExpiration)
		} else {
			var nval = append(val.([]string), apiKey)
			s.Cache.Set("keys", nval, cache.DefaultExpiration)
		}

		ctx.Header("Expires", time.Now().Add( 1 * time.Hour ).String())

		ctx.JSON(http.StatusOK, gin.H{
			"api_key": apiKey,
		})
	}

	ctx.AbortWithStatus(http.StatusUnauthorized)
}
