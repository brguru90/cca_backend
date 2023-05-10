package my_modules

import (
	"context"
	"path/filepath"

	firebase "firebase.google.com/go/v4"
	log "github.com/sirupsen/logrus"

	"google.golang.org/api/option"
)

var FIREBASE_APP *firebase.App

func InitFirebase() {
	var err error
	service_account_file, err := filepath.Abs("./env/cca-vijayapura-firebase-adminsdk-ghz2d-1f8e7ad071.json")
	if err != nil {
		log.WithFields(log.Fields{
			"service_account_file": service_account_file,
			"Error":                err,
		}).Panic("FIREBASE_APP: Error getting absolute path")
	}
	opt := option.WithCredentialsFile(service_account_file)
	FIREBASE_APP, err = firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.WithFields(log.Fields{
			"opt":   opt,
			"Error": err,
		}).Panic("Unable to initialize FIREBASE_APP")
	}
	log.Infoln(FIREBASE_APP)
}
