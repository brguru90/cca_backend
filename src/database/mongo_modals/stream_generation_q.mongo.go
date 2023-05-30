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
)

type VideoStreamGenerationQModel struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	VideoID   primitive.ObjectID `json:"video_id,omitempty" bson:"video_id,omitempty"`
	Started   bool               `json:"started" bson:"started"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty"`
}

func InitVideoStreamGenerationQCollection() {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"video_id": 1,
			},
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
