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

type VideoUploadModal struct {
	ID                      primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Title                   string             `json:"title"  binding:"required" bson:"title"`
	Description             string             `json:"description,omitempty" bson:"description,omitempty"`
	Duration                int64              `json:"duration,omitempty"  binding:"required" bson:"duration,omitempty"`
	IsLive                  bool               `json:"is_live,omitempty" binding:"required" bson:"is_live"`
	CreatedByUser           string             `json:"created_by_user,omitempty" bson:"created_by_user,omitempty"`
	UploadedByUser          primitive.ObjectID `json:"uploaded_by_user,omitempty" bson:"uploaded_by_user,omitempty"`
	PathToVideoPreviewImage string             `json:"path_to_video_preview_image,omitempty" binding:"required" bson:"path_to_video_preview_image,omitempty"`
	PathToOriginalVideo     string             `json:"path_to_original_video,omitempty" binding:"required" bson:"path_to_original_video,omitempty"`
	PathToVideoStream       string             `json:"path_to_video_stream,omitempty" bson:"path_to_video_stream,omitempty"`
	VideoDecryptionKey      string             `json:"video_decryption_key,omitempty" bson:"video_decryption_key,omitempty"`
	LinkToVideoStream       string             `json:"link_to_video_stream,omitempty" bson:"link_to_video_stream,omitempty"`
	LinkToVideoPreviewImage string             `json:"link_to_video_preview_image,omitempty" bson:"link_to_video_preview_image,omitempty"`
	CreatedAt               time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt               time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitVideoUploadCollection() {
	indexes := []mongo.IndexModel{
		{
			// index
			Keys: bson.M{
				"title": 1,
			},
			Options: options.Index().SetUnique(true),
		},
		{
			// composite key
			Keys: bson.D{
				{
					Key:   "title",
					Value: 1,
				},
				{
					Key:   "path_to_original_video",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
		{
			// composite key
			Keys: bson.D{
				{
					Key:   "title",
					Value: 1,
				},
				{
					Key:   "path_to_video_stream",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.VideoUploads.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListModal")
	}
}
