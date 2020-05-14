package auth

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lexffe/backend.lexffe.io/helpers"
	"github.com/patrickmn/go-cache"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Handler checks the otp key against the database, then returns a temporary api key valid for 2 hours
func (s *AuthenticateHandler) Handler(ctx *gin.Context) {

	// Parse body: otp
	// body should be just text, not json.

	body, err := ioutil.ReadAll(ctx.Request.Body)

	if err != nil {
		ctx.Error(errors.New("user cannot be authenticated. reason: body cannot be parsed"))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	bodyBuf := bytes.NewBuffer(body)

	// look for user's otp key in the db

	f, err := os.OpenFile(otpSecretPath, os.O_RDONLY, 0600)

	if err != nil {
		ctx.Error(errors.New("user cannot be authenticated. reason: .otp file cannot be accessed"))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	defer f.Close()

	content, err := ioutil.ReadAll(f)

	if err != nil {
		ctx.Error(errors.New("user cannot be authenticated. reason: file cannot be read"))
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if len(content) == 0 {
		ctx.AbortWithError(http.StatusInternalServerError, errors.New("user cannot be authenticated. reason: .otp is empty"))
		return
	}

	secret := bytes.NewBuffer(content)

	totpOpts := totp.ValidateOpts{
		Digits:    otp.DigitsEight,
		Algorithm: otp.AlgorithmSHA512,
	}

	validated, err := totp.ValidateCustom(bodyBuf.String(), secret.String(), time.Now(), totpOpts)

	if err != nil {
		ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	if validated == true { // authenticated.
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

		ctx.Header("Expires", time.Now().Add(1*time.Hour).Format(time.RFC3339))

		ctx.String(http.StatusOK, apiKey)
		return
	}

	ctx.AbortWithStatus(http.StatusUnauthorized)
	return
}
