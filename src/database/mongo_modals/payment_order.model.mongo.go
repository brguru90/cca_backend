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

type PaymentOrderModal struct {
	ID                   primitive.ObjectID   `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	UserID               primitive.ObjectID   `json:"user_id,omitempty" binding:"required" bson:"user_id,omitempty"`
	UserSubscriptionsIDs []primitive.ObjectID `json:"user_subscriptions_ids,omitempty" binding:"required" bson:"user_subscriptions_ids,omitempty"`
	OrderID              string               `json:"order_id"  binding:"required" bson:"order_id"`
	Amount               int64                `json:"amount,omitempty"  binding:"required" bson:"amount,omitempty"`
	CreatedAt            time.Time            `json:"createdAt,omitempty" bson:"CreatedAt" swaggerignore:"true"`
	UpdatedAt            time.Time            `json:"updatedAt,omitempty" bson:"UpdatedAt" swaggerignore:"true"`
}

func InitPaymentOrderCollection() {
	indexes := []mongo.IndexModel{
		{
			// index
			Keys: bson.M{
				"order_id": 1,
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.PaymentOrder.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on PaymentOrder")
	}
}
