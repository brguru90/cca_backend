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

type VideoPlayListSubscriptionPackageModal struct {
	ID           primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Title        string               `json:"title"  binding:"required" bson:"title"`
	Description  string               `json:"description,omitempty" bson:"description,omitempty"`
	CreatedBy    primitive.ObjectID   `json:"created_by,omitempty" bson:"created_by,omitempty"`
	PlaylistsIDs []primitive.ObjectID `json:"playlists_ids,omitempty" binding:"required" bson:"playlists_ids,omitempty"`
	IsEnabled    bool                 `json:"is_enabled,omitempty"  binding:"required" bson:"is_enabled,omitempty"`
	Price        int64                `json:"price,omitempty" binding:"required" bson:"price"`
	CreatedAt    time.Time            `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt    time.Time            `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitVideoPlayListSubscriptionPackageCollection() {
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
	_, err := database_connections.MONGO_COLLECTIONS.VideoPlayListSubscriptionPackage.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListSubscriptionPackage")
	}
}
