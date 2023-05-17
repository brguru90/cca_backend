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

type UsersModel struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Uid               string             `json:"uid,omitempty" bson:"uid,omitempty"`
	AuthProvider      string             `json:"auth_provider,omitempty" bson:"auth_provider,omitempty"`
	Email             string             `json:"email,omitempty" binding:"required" bson:"email,omitempty"`
	Mobile            string             `json:"mobile,omitempty" binding:"required" bson:"mobile,omitempty"`
	Password          string             `json:"password,omitempty" binding:"required" bson:"password,omitempty"`
	Username          string             `json:"username,omitempty" binding:"required" bson:"username,omitempty"`
	AccessLevel       string             `json:"access_level,omitempty" binding:"required" bson:"access_level,omitempty"`
	AccessLevelWeight int                `json:"access_level_weight,omitempty" binding:"required" bson:"access_level_weight,omitempty"`
	CreatedAt         time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt         time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitUserCollection() {
	indexes := []mongo.IndexModel{
		{
			// index
			Keys: bson.M{
				"email": 1,
			},
			// Options: options.Index().SetUnique(true),
		},
		{
			// index
			Keys: bson.M{
				"mobile": 1,
			},
		},
		{
			// index
			Keys: bson.M{
				"uid": 1,
			},
		},
		{
			// composite key
			Keys: bson.D{
				{
					Key:   "uid",
					Value: 1,
				},
				{
					Key:   "email",
					Value: 1,
				},
				{
					Key:   "mobile",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.Users.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on UsersModel")
	}
}
