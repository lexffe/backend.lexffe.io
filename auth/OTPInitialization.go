package auth

import (
	"context"
	"log"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// OTPInitialization initialises the OTP key for admin access
func OTPInitialization(ctx context.Context, db *mongo.Database) error {

	// check if database has existing OTP code. there should only be one

	res := db.Collection(collectionAuth).FindOne(ctx, bson.M{})

	if res.Err() != nil {

		// if doesn't exist, generate new one, write into database

		if res.Err() == mongo.ErrNoDocuments {

			// generate totp key
			log.Println("totp key not found, generating a new one.")

			var auth authDBModel

			totpKey, err := totp.Generate(totp.GenerateOpts{
				Issuer:      totpIssuer,
				AccountName: totpAccountName,
				Digits:      otp.DigitsEight,
				Algorithm:   otp.AlgorithmSHA512,
			})

			if err != nil {
				return err
			}

			log.Printf("new totp key: %v\n", totpKey.String())

			auth.OTPKey = totpKey.String()

			// insert model into database

			if _, err := db.Collection(collectionAuth).InsertOne(ctx, auth); err != nil {
				return err
			}

			return nil

		}

		// (other kinds of error)

		return res.Err()

	}

	return nil
}
