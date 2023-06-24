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

type SubsequentUserPlaylistSubscriptionStruct struct {
	SubscribedOn   time.Time `json:"subscribed_on,omitempty" binding:"required" bson:"subscribed_on"`
	AmountPaid     int64     `json:"amount_paid,omitempty" bson:"amount_paid"`
	DurationInDays int       `json:"duration_in_days,omitempty" binding:"required" bson:"duration_in_days"`
}

// user subscription could be based on individual playlist or by subscription package
// Ex: for playlist
// userID | SubPkgId | PlaylistID | ...
// guru   | nil      | pl1
// guru   | nil      | pl2

// Ex: for subscription package ( suppose a package is bundle of 2 playlist i.e., pkg_1[pl1 & pl2])
// userID | SubPkgId | PlaylistID | ...
// guru   | pkg_1      | pl1
// guru   | pkg_1      | pl2

type VideoPlayListUserSubscriptionModal struct {
	ID                      primitive.ObjectID                         `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	UserID                  primitive.ObjectID                         `json:"user_id,omitempty" binding:"required" bson:"user_id,omitempty"`
	Username                string                                     `json:"username,omitempty" binding:"required" bson:"username,omitempty"`
	SubscriptionPackageId   primitive.ObjectID                         `json:"subscription_package_id,omitempty" binding:"required" bson:"subscription_package_id,omitempty"`
	PlaylistID              primitive.ObjectID                         `json:"playlist_id,omitempty" binding:"required" bson:"playlist_id,omitempty"`
	Price                   int64                                      `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	InitialSubscriptionDate time.Time                                  `json:"initial_subscription_date"  binding:"required" bson:"initial_subscription_date"`
	ExpireOn                time.Time                                  `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
	IsEnabled               bool                                       `json:"is_enabled,omitempty"  binding:"required" bson:"is_enabled,omitempty"`
	Subscriptions           []SubsequentUserPlaylistSubscriptionStruct `json:"subscriptions,omitempty" binding:"required" bson:"subscriptions"`
	CreatedAt               time.Time                                  `json:"createdAt,omitempty" bson:"CreatedAt" swaggerignore:"true"`
	UpdatedAt               time.Time                                  `json:"updatedAt,omitempty" bson:"UpdatedAt" swaggerignore:"true"`
}

func InitVideoPlayListUserSubscriptionCollection() {
	indexes := []mongo.IndexModel{
		{
			// composite key
			Keys: bson.D{
				{
					Key:   "user_id",
					Value: 1,
				},
				{
					Key:   "playlist_id",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListUserSubscription")
	}
}
