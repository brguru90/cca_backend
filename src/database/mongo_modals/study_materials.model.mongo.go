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

type StudyMaterialsModal struct {
	ID                       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	Title                    string             `json:"title"  binding:"required" bson:"title"`
	Description              string             `json:"description,omitempty" bson:"description,omitempty"`
	Category                 string             `json:"category,omitempty" bson:"category,omitempty"`
	IsLive                   bool               `json:"is_live,omitempty" binding:"required" bson:"is_live"`
	CreatedByUser            string             `json:"created_by_user,omitempty" bson:"created_by_user,omitempty"`
	UploadedByUser           primitive.ObjectID `json:"uploaded_by_user,omitempty" bson:"uploaded_by_user,omitempty"`
	PathToBookCoverImage     string             `json:"path_to_cover_image,omitempty" binding:"required" bson:"path_to_cover_image,omitempty"`
	PathToDocFile            string             `json:"path_to_doc_file,omitempty" binding:"required" bson:"path_to_doc_file,omitempty"`
	FileDecryptionKey        string             `json:"file_decryption_key,omitempty" bson:"file_decryption_key,omitempty"`
	FileDecryptionKeyBlkSize int                `json:"file_decryption_key_blk_size,omitempty" bson:"file_decryption_key_blk_size,omitempty"`
	LinkToBookCoverImage     string             `json:"link_to_book_cover_image,omitempty" bson:"link_to_book_cover_image,omitempty"`
	LinkToDocFile            string             `json:"link_to_doc_file,omitempty" bson:"link_to_doc_file,omitempty"`
	Price                    int64              `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	EnrollDays               int16              `json:"enroll_days,omitempty" bson:"enroll_days,omitempty"`
	CreatedAt                time.Time          `json:"createdAt,omitempty" swaggerignore:"true"`
	UpdatedAt                time.Time          `json:"updatedAt,omitempty" swaggerignore:"true"`
}

func InitStudyMaterialCollection() {
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
					Key:   "path_to_doc_file",
					Value: 1,
				},
			},
			Options: options.Index().SetUnique(true),
		},
	}

	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)
	_, err := database_connections.MONGO_COLLECTIONS.StudyMaterial.Indexes().CreateMany(context.Background(), indexes, opts)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("error in creating index on VideoPlayListModal")
	}
}
