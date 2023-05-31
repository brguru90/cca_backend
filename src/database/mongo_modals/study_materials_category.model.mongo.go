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

type StudyMaterialCategoryModal struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Title     string             `json:"title"  binding:"required" bson:"title"`
	CreatedAt time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitStudyMaterialCategoryModalCollection() {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.M{
				"title": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.StudyMaterialCategory.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListModal")
	}
}
