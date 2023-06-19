package mongo_modals

import (
	"context"
	"time"

	"cca/src/database/database_connections"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/x/bsonx"
)

type VideoStreamGenerationQModel struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	VideoID   primitive.ObjectID `json:"video_id,omitempty" bson:"video_id,omitempty"`
	Started   bool               `json:"started" bson:"started"`
	StartedAt time.Time          `json:"startedAt,omitempty" bson:"startedAt,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
}

func InitVideoStreamGenerationQCollection() {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"video_id": 1,
			},
		},
		{
			Keys:    bsonx.Doc{{Key: "startedAt", Value: bsonx.Int32(1)}},
			Options: options.Index().SetExpireAfterSeconds(60 * 60 * 6), // delete video stream generation entry after 6 hour from process start
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on ActiveSessionsModel")
	}
}
