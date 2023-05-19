package mongo_modals

import (
	"cca/src/database/database_connections"
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type AppBuildRegistrationModal struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	AppID     string             `json:"app_id"  binding:"required" bson:"app_id"`
	AppSecret string             `json:"app_secret,omitempty" bson:"app_secret,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitAppBuildRegistrationCollection() {
	indexes := []mongo.IndexModel{
		{
			// composite key
			Keys: bson.D{
				{
					Key:   "app_id",
					Value: 1,
				},
				{
					Key:   "app_secret",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.AppBuildRegistration.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on AppBuildRegistrationModal")
	}
}
