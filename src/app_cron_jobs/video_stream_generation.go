package app_cron_jobs

import (
	"cca/src/database/database_connections"
	"cca/src/my_modules"
	"context"
	"time"

	"cca/src/configs"
	"cca/src/database/mongo_modals"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func VideoStreamGeneration() {
	log.WithFields(log.Fields{
		"time": time.Now(),
	}).Debug(" -- VideoStreamGeneration Cron job started -- ")

	ctx := context.Background()

	// opts := options.Count().SetHint("_id_")
	// listCount, count_err := database_connections.MONGO_COLLECTIONS.StudyMaterial.CountDocuments(ctx, bson.M{
	// 	"started": true,
	// }, opts)
	// if count_err == nil && listCount > 0 {
	// 	return
	// }

	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.Find(ctx, bson.M{
		"started": false,
	})
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		return
	}
	var streamQ []mongo_modals.VideoStreamGenerationQModel = []mongo_modals.VideoStreamGenerationQModel{}
	if err = cursor.All(context.TODO(), &streamQ); err != nil {
		return
	}

	videos_ids := []primitive.ObjectID{}
	for i := 0; i < len(streamQ); i++ {
		videos_ids = append(videos_ids, streamQ[i].VideoID)
	}

	if len(videos_ids) == 0 {
		return
	}

	go func() {
		var err error
		var cursor *mongo.Cursor
		cursor, err = database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, bson.M{"_id": bson.M{"$in": videos_ids}})
		if err != nil {
			if err != context.Canceled {
				log.WithFields(log.Fields{
					"error": err,
				}).Panic("QueryRow failed ==>")
			}
			return
		}
		defer cursor.Close(ctx)

		database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.UpdateMany(ctx,
			bson.M{"video_id": bson.M{"$in": videos_ids}},
			bson.M{
				"$set": bson.M{
					"started":   true,
					"startedAt": time.Now(),
				},
			},
		)

		UNPROTECTED_UPLOAD_PATH := configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH
		if strings.HasSuffix(UNPROTECTED_UPLOAD_PATH, "/") {
			UNPROTECTED_UPLOAD_PATH = UNPROTECTED_UPLOAD_PATH[:len(UNPROTECTED_UPLOAD_PATH)-1]
		}
		CDN_PATH := "/" + configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH_ROUTE
		unprotected_video := fmt.Sprintf("%s/video", UNPROTECTED_UPLOAD_PATH)

		for cursor.Next(ctx) {
			var videoData mongo_modals.VideoUploadModal
			if err = cursor.Decode(&videoData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				return
			}
			path_split := strings.Split(videoData.PathToOriginalVideo, "/")
			file_full_name := strings.Split(path_split[len(path_split)-1], ".")
			file_name := strings.Join(file_full_name[:len(file_full_name)-1], "_")
			path_to_video_stream := ""
			video_decryption_key := ""
			var data my_modules.UploadedVideoInfoStruct
			log.WithFields(log.Fields{
				"title": videoData.Title,
				"id":    videoData.ID,
			}).Debugln("ffmpeg process started")
			if data, err = my_modules.UploadVideoForStream(videoData.ID.Hex(), unprotected_video, file_name, videoData.PathToOriginalVideo); err == nil {
				path_to_video_stream = data.StreamGeneratedLocation
				video_decryption_key = data.DecryptionKey
			} else {
				log.WithFields(log.Fields{
					"title": videoData.Title,
					"id":    videoData.ID,
					"err":   err,
				}).Errorln("ffmpeg process failed")
				os.Remove(data.OutputDir)
			}

			log.WithFields(log.Fields{
				"path_to_video_stream": path_to_video_stream,
				"link_to_video_stream": strings.Replace(path_to_video_stream, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
				"video_decryption_key": video_decryption_key,
			}).Debugln("ffmpeg process done")

			database_connections.MONGO_COLLECTIONS.VideoUploads.UpdateOne(
				context.Background(),
				bson.M{
					"_id": videoData.ID,
				},
				bson.M{
					"$set": bson.M{
						"path_to_video_stream": path_to_video_stream,
						"link_to_video_stream": strings.Replace(path_to_video_stream, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
						"video_decryption_key": video_decryption_key,
						"is_live":              true,
					},
				},
			)
			if _, err := database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.DeleteOne(ctx, bson.M{"video_id": videoData.ID}); err != nil {
				log.WithFields(log.Fields{
					"Error": err,
				}).Errorln("delete VideoStreamGenerationQ")
			}
		}

	}()
}
