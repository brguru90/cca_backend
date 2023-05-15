package database

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VideoPlayListSubscriptionModal struct {
	ID                      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	UserID                  primitive.ObjectID `json:"user_id,omitempty" binding:"required" bson:"user_id,omitempty"`
	PlaylistID              primitive.ObjectID `json:"playlist_id,omitempty" binding:"required" bson:"playlist_id,omitempty"`
	InitialSubscriptionDate time.Time          `json:"initial_subscription_date"  binding:"required" bson:"initial_subscription_date"`
	ExpireOn                time.Time          `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
	IsEnabled               bool               `json:"is_enabled,omitempty"  binding:"required" bson:"is_enabled,omitempty"`
	SubscribedOn            []time.Time        `json:"subscribed_on,omitempty" binding:"required" bson:"subscribed_on"`
	AmountPaid              []int64            `json:"amount_paid,omitempty" binding:"required" bson:"amount_paid"`
	CreatedAt               time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt               time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitVideoPlayListSubscriptionCollection() {
	indexes := []mongo.IndexModel{
		{
			// index
			Keys: bson.M{
				"title": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := MONGO_COLLECTIONS.VideoPlayListSubscription.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListModal")
	}
}
