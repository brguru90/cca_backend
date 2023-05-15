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

type VideoPlayListModal struct {
	ID          primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Title       string               `json:"title"  binding:"required" bson:"title"`
	Description string               `json:"description,omitempty" bson:"description,omitempty"`
	Price       int64                `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	IsLive      bool                 `json:"is_live,omitempty" binding:"required" bson:"is_live"`
	VideosIDs   []primitive.ObjectID `json:"videos_ids,omitempty" binding:"required"  bson:"videos_ids,omitempty"`
	CreatedAt   time.Time            `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt   time.Time            `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitVideoPlayListCollection() {
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
	_, err := MONGO_COLLECTIONS.VideoPlayList.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListModal")
	}
}
