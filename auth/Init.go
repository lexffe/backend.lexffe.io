package auth

import (
	"io/ioutil"
	"log"

	"os"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// OTPInitialization initialises the OTP key for admin access
func (s *AuthenticateHandler) OTPInitialization() error {

	// check if database has existing OTP code. there should only be one

	// open file with rw access, or create if it doesn't exist
	f, err := os.OpenFile(otpSecretPath, os.O_RDWR|os.O_CREATE, 0600)

	if err != nil {
		return err
	}

	defer f.Close()

	content, err := ioutil.ReadAll(f)

	if err != nil {
		return err
	}

	// nothing is in the file.
	if len(content) == 0 {
		totpKey, err := totp.Generate(totp.GenerateOpts{
			Issuer:      s.Issuer,
			AccountName: "admin@" + s.Issuer,
			Digits:      otp.DigitsEight,
			Algorithm:   otp.AlgorithmSHA512,
		})

		if err != nil {
			return err
		}

		log.Printf("new totp key: %v\n", totpKey.String())

		_, err = f.WriteString(totpKey.Secret())

		if err != nil {
			return err
		}

		if err := f.Sync(); err != nil {
			return err
		}
	}

	return nil
}
