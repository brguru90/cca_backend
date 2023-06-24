package admin_views

import (
	"cca/src/configs"
	"cca/src/database/database_connections"
	"cca/src/database/database_utils"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type VideoUploadReqStruct struct {
	VideoFile        *multipart.FileHeader `form:"video_file" binding:"required"`
	PreviewImageFile *multipart.FileHeader `form:"preview_image_file" binding:"required"`
}
type VideoInfoReqStruct struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	CreatedBy   string `form:"created_by" binding:"required"`
	IsLive      bool   `form:"is_live"`
}

// @BasePath /api/
// @Summary video upload
// @Schemes
// @Description api to upload video content for multiple adaptive bit rate streaming
// @Tags Manage Videos
// @Accept mpfd
// @Produce json
// @Param video_file formData file true "Video file"
// @Param preview_image_file formData file true "Preview image file"
// @Param form_data formData VideoInfoReqStruct true "Video upload"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/upload_streaming_video/ [post]
func UploadVideo(c *gin.Context) {
	ctx := c.Request.Context()

	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	var uploadForm VideoUploadReqStruct
	var infoForm VideoInfoReqStruct
	PROTECTED_UPLOAD_PATH := configs.EnvConfigs.PROTECTED_UPLOAD_PATH
	UNPROTECTED_UPLOAD_PATH := configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH
	if strings.HasSuffix(PROTECTED_UPLOAD_PATH, "/") {
		PROTECTED_UPLOAD_PATH = PROTECTED_UPLOAD_PATH[:len(PROTECTED_UPLOAD_PATH)-1]
	}
	if strings.HasSuffix(UNPROTECTED_UPLOAD_PATH, "/") {
		UNPROTECTED_UPLOAD_PATH = UNPROTECTED_UPLOAD_PATH[:len(UNPROTECTED_UPLOAD_PATH)-1]
	}
	CDN_PATH := "/" + configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH_ROUTE
	unprotected_image := fmt.Sprintf("%s/image", UNPROTECTED_UPLOAD_PATH)
	protected_video := fmt.Sprintf("%s/video", PROTECTED_UPLOAD_PATH)
	upload_path := fmt.Sprintf("%s/original_video", protected_video)
	if err := os.MkdirAll(upload_path, 0755); err != nil {
		log.Warnln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error creating directory", nil)
		return
	}
	if err := os.MkdirAll(unprotected_image, 0755); err != nil {
		log.Warnln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error creating directory", nil)
		return
	}
	if err := c.ShouldBind(&infoForm); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid data payload", nil)
		return
	}
	if err := c.ShouldBind(&uploadForm); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}
	// file_part := strings.Split(uploadForm.VideoFile.Filename, ".")
	// file_part := strings.Split(uploadForm.VideoFile.Filename, ".")

	// file_name, file_ext := strings.Join(file_part[:len(file_part)-1], "."), file_part[len(file_part)-1]
	// dst_file_path := fmt.Sprintf("%s/%s_%d.%s", upload_path, file_name, time.Now().UnixMilli(), file_ext)
	// dst_file_path := fmt.Sprintf("./uploads/original_video/%s", video_title)

	non_alphaNumeric := regexp.MustCompile(`[^a-zA-Z0-9]`)
	dst_file_name := non_alphaNumeric.ReplaceAllString(infoForm.Title, "")
	dst_file_path := fmt.Sprintf("%s/%s_%d", upload_path, dst_file_name, time.Now().UnixMilli())
	dst_image_file_path_public := fmt.Sprintf("%s/%s_%d", unprotected_image, dst_file_name, time.Now().UnixMilli())
	dst_video_file_path := fmt.Sprintf("%s.mp4", dst_file_path)
	preview_image_file_part := strings.Split(uploadForm.PreviewImageFile.Filename, ".")
	if len(preview_image_file_part) <= 1 {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "File name doesn't contain file extension", nil)
		return
	}
	dst_preview_image_file_path := fmt.Sprintf("%s.%s", dst_image_file_path_public, preview_image_file_part[len(preview_image_file_part)-1])

	if err := c.SaveUploadedFile(uploadForm.VideoFile, dst_video_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}

	if err := c.SaveUploadedFile(uploadForm.PreviewImageFile, dst_preview_image_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}

	_time := time.Now()
	_, ins_err := database_connections.MONGO_COLLECTIONS.VideoUploads.InsertOne(ctx, mongo_modals.VideoUploadModal{
		Title:                   infoForm.Title,
		CreatedByUser:           infoForm.CreatedBy,
		Description:             infoForm.Description,
		PathToOriginalVideo:     dst_video_file_path,
		PathToVideoPreviewImage: dst_preview_image_file_path,
		LinkToVideoPreviewImage: strings.Replace(dst_preview_image_file_path, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
		IsLive:                  infoForm.IsLive,
		UploadedByUser:          _id,
		CreatedAt:               _time,
		UpdatedAt:               _time,
	})
	if ins_err != nil {
		os.Remove(dst_video_file_path)
		resp_err, is_known := database_utils.GetDBErrorString(ins_err)
		if !is_known {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
		}
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", resp_err, nil)
		return
	}

	// unprotected_video := fmt.Sprintf("%s/public/video", base_path)
	// if err := my_modules.UploadVideoForStream(c, unprotected_video, file_name, dst_file_path); err != nil {
	// 	log.Errorln(err)
	// 	my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Failed to create  multi bit rate file chunks", nil)
	// 	return
	// }
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", nil)
}

type VideoStreamReqStruct struct {
	Id []string `json:"video_ids" binding:"required"`
}

// @BasePath /api/
// @Summary Generate video stream
// @Schemes
// @Description api to get the list of all the videos uploaded by the logged user
// @Tags Manage Videos
// @Accept json
// @Produce json
// @Param video_ids body VideoStreamReqStruct true "Video IDs"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/generate_video_stream/ [post]
func GenerateVideoStream(c *gin.Context) {
	ctx := c.Request.Context()

	var videoStreamInfo VideoStreamReqStruct
	if err := c.ShouldBind(&videoStreamInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}

	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	user_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if payload.Data.ID == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	stream_generation_status := make(map[primitive.ObjectID]string)
	for i := 0; i < len(videoStreamInfo.Id); i++ {
		objID, err := primitive.ObjectIDFromHex(videoStreamInfo.Id[i])
		if err == nil {
			var streamQ mongo_modals.VideoStreamGenerationQModel
			err = database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.FindOne(ctx, bson.M{
				"video_id": objID,
			}).Decode(&streamQ)
			if err == nil {
				stream_generation_status[objID] = "in_progress"
			} else if err == mongo.ErrNoDocuments {

				where := bson.M{"_id": objID, "uploaded_by_user": user_id}

				if payload.Data.AccessLevel == "super_admin" {
					where = bson.M{"_id": objID}
				}

				var videoData mongo_modals.VideoUploadModal
				err := database_connections.MONGO_COLLECTIONS.VideoUploads.FindOne(ctx, where).Decode(&videoData)
				if err != nil {
					stream_generation_status[objID] = "not started"
					continue
				}

				_time := time.Now()
				_, ins_err := database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.InsertOne(ctx, mongo_modals.VideoStreamGenerationQModel{
					VideoID:   objID,
					Started:   false,
					CreatedAt: _time,
					UpdatedAt: _time,
				})
				if ins_err == nil {
					stream_generation_status[objID] = "started"
				} else {
					stream_generation_status[objID] = "unknown_error"
				}

			} else {
				stream_generation_status[objID] = "unknown_error"
			}
		}
	}

	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Started", stream_generation_status)
	// UNPROTECTED_UPLOAD_PATH := configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH
	// if strings.HasSuffix(UNPROTECTED_UPLOAD_PATH, "/") {
	// 	UNPROTECTED_UPLOAD_PATH = UNPROTECTED_UPLOAD_PATH[:len(UNPROTECTED_UPLOAD_PATH)-1]
	// }
	// CDN_PATH := "/" + configs.EnvConfigs.UNPROTECTED_UPLOAD_PATH_ROUTE
	// unprotected_video := fmt.Sprintf("%s/video", UNPROTECTED_UPLOAD_PATH)

	// where := bson.M{"_id": bson.M{"$in": videos_doc_id}, "uploaded_by_user": user_id}

	// if payload.Data.AccessLevel == "super_admin" {
	// 	where = bson.M{"_id": bson.M{"$in": videos_doc_id}}
	// }

	// var err error
	// var cursor *mongo.Cursor
	// defer cursor.Close(ctx)
	// cursor, err = database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, where)
	// if err != nil {
	// 	if err != context.Canceled {
	// 		log.WithFields(log.Fields{
	// 			"error": err,
	// 		}).Errorln("QueryRow failed ==>")
	// 	}
	// 	my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
	// } else {
	// 	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Started video stream creation", nil)
	// 	// c.Abort()
	// 	// conn, _, err := c.Writer.Hijack()
	// 	// if err == nil {
	// 	// 	conn.Close()
	// 	// }
	// 	go func(ids []primitive.ObjectID) {
	// 		ctx, _ := context.WithTimeout(context.Background(), time.Hour*1)
	// 		cursor, err := database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	// 		if err == nil {
	// 			defer cursor.Close(ctx)
	// 		}
	// 		if err := database_connections.RedisPoolSet("video_stream_generation_in_progress", "value", 60*time.Minute); err != nil {
	// 			log.WithFields(log.Fields{
	// 				"Error": err,
	// 			}).Panic("Unable to write into redis pool")
	// 		}
	// 		defer func() {
	// 			if err := database_connections.RedisPoolDel("video_stream_generation_in_progress"); err != nil {
	// 				log.WithFields(log.Fields{
	// 					"Error": err,
	// 				}).Errorln("Unable to write into redis pool")
	// 			}
	// 		}()
	// 		for cursor.Next(ctx) {
	// 			var videoData mongo_modals.VideoUploadModal
	// 			if err = cursor.Decode(&videoData); err != nil {
	// 				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
	// 				return
	// 			}
	// 			path_split := strings.Split(videoData.PathToOriginalVideo, "/")
	// 			file_full_name := strings.Split(path_split[len(path_split)-1], ".")
	// 			file_name := strings.Join(file_full_name[:len(file_full_name)-1], "_")
	// 			path_to_video_stream := ""
	// 			video_decryption_key := ""
	// 			var data my_modules.UploadedVideoInfoStruct
	// 			if data, err = my_modules.UploadVideoForStream(videoData.ID.Hex(), unprotected_video, file_name, videoData.PathToOriginalVideo); err == nil {
	// 				path_to_video_stream = data.StreamGeneratedLocation
	// 				video_decryption_key = data.DecryptionKey
	// 			} else {
	// 				os.Remove(data.OutputDir)
	// 			}

	// 			log.WithFields(log.Fields{
	// 				"path_to_video_stream": path_to_video_stream,
	// 				"link_to_video_stream": strings.Replace(path_to_video_stream, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
	// 				"video_decryption_key": video_decryption_key,
	// 			}).Debugln("Unable to write into redis pool")

	// 			database_connections.MONGO_COLLECTIONS.VideoUploads.UpdateOne(
	// 				context.Background(),
	// 				bson.M{
	// 					"_id": videoData.ID,
	// 				},
	// 				bson.M{
	// 					"$set": bson.M{
	// 						"path_to_video_stream": path_to_video_stream,
	// 						"link_to_video_stream": strings.Replace(path_to_video_stream, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
	// 						"video_decryption_key": video_decryption_key,
	// 						"is_live":              true,
	// 					},
	// 				},
	// 			)
	// 		}
	// 	}(videos_doc_id)
	// }

}

type GetAllUploadedVideosRespStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.VideoUploadModal `json:"data"`
}

// @BasePath /api/
// @Summary get all uploaded videos
// @Schemes
// @Description api to get the list of all the videos uploaded by the logged user
// @Tags Manage Videos
// @Produce json
// @Success 200 {object} GetAllUploadedVideosRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/upload_list/ [get]
func GetAllUploadedVideos(c *gin.Context) {
	ctx := c.Request.Context()
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, bson.M{
		"uploaded_by_user": _id,
	})
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
		return
	} else {
		defer cursor.Close(ctx)
		var videosList []mongo_modals.VideoUploadModal = []mongo_modals.VideoUploadModal{}
		for cursor.Next(c.Request.Context()) {
			var videoData mongo_modals.VideoUploadModal
			if err = cursor.Decode(&videoData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				// continue
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving user data", nil)
				return
			}
			videosList = append(videosList, videoData)
		}

		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", videosList)
	}
}

// @BasePath /api/
// @Summary Delete videos
// @Schemes
// @Description api delete videos by id
// @Tags Manage Videos
// @Accept json
// @Produce json
// @Param video_ids body VideoStreamReqStruct true "Remove videos"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/delete_streaming_video/ [delete]
func RemoveVideos(c *gin.Context) {
	ctx := c.Request.Context()
	var videos_data VideoStreamReqStruct
	if err := c.ShouldBind(&videos_data); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}

	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	user_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if payload.Data.ID == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	var videos_doc_id []primitive.ObjectID
	for i := 0; i < len(videos_data.Id); i++ {
		objID, err := primitive.ObjectIDFromHex(videos_data.Id[i])
		if err == nil {
			videos_doc_id = append(videos_doc_id, objID)
		}
	}

	where := bson.M{"_id": bson.M{"$in": videos_doc_id}, "uploaded_by_user": user_id}

	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{"_id": bson.M{"$in": videos_doc_id}}
	}

	opts := options.Count().SetHint("_id_")
	listCount, _ := database_connections.MONGO_COLLECTIONS.VideoPlayList.CountDocuments(c.Request.Context(),
		bson.M{
			"videos_ids.video_id": bson.M{"$in": videos_doc_id},
		},
		opts)

	if listCount > 0 {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Remove video from playlist first", nil)
		return
	}

	var response_data = make(map[string]interface{})
	var files_deleted = make(map[string]interface{})
	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, where)
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
	} else {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var videoData mongo_modals.VideoUploadModal
			if err = cursor.Decode(&videoData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				continue
			}
			files_deleted[videoData.Title] = map[string]interface{}{
				"original_video": false,
				"stream_video":   false,
			}
			video_stream_path := strings.Split(videoData.PathToVideoStream, "/")
			video_stream_path = video_stream_path[:len(video_stream_path)-1]
			err1 := os.Remove(videoData.PathToOriginalVideo)
			err2 := os.RemoveAll(strings.Join(video_stream_path, "/"))
			files_deleted[videoData.Title] = map[string]interface{}{
				"original_video": err1 == nil,
				"stream_video":   err2 == nil,
			}
		}
	}

	result, err := database_connections.MONGO_COLLECTIONS.VideoUploads.DeleteMany(context.Background(), where)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("Failed to delete user data")
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to delete user data", nil)
		return
	}
	response_data["files_deleted_for_title"] = files_deleted
	if result.DeletedCount > 0 {
		response_data["deleted_count"] = result.DeletedCount
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Removed successfully", response_data)
		return
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "error", "Failed to removed", response_data)
}

type GetAllPlayListsRespStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.VideoPlayListModal `json:"data"`
}

// @BasePath /api/
// @Summary Get list of playlist
// @Schemes
// @Description api to fetch existing playlist
// @Tags Playlist
// @Accept json
// @Produce json
// @Success 200 {object} GetAllPlayListsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/playlist/ [get]
func GetAllPlayLists(c *gin.Context) {
	// If Admin: All playlist created by user
	// If Super Admin: All playlist created by all user
	ctx := c.Request.Context()
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	where := bson.M{
		"created_by_user": _id,
	}
	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{}
	}
	cursor, err := database_connections.MONGO_COLLECTIONS.VideoPlayList.Find(ctx, where)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.WithFields(log.Fields{
				"err": err,
			}).Errorln("Failed to load session data")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
		return
	} else {
		defer cursor.Close(ctx)
		var playlistsData []mongo_modals.VideoPlayListModal = []mongo_modals.VideoPlayListModal{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var playlistData mongo_modals.VideoPlayListModal
			if err = cursor.Decode(&playlistData); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
				return
			}
			playlistsData = append(playlistsData, playlistData)
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", playlistsData)
		return
	}
}

type CreatePlayListRespStruct struct {
	my_modules.ResponseFormat
	Data mongo_modals.VideoPlayListModal `json:"data"`
}

// @BasePath /api/
// @Summary Create new playlist
// @Schemes
// @Description api to create new empty playlist
// @Tags Playlist
// @Accept json
// @Produce json
// @Param new_playlist_data body mongo_modals.VideoPlayListModal true "New Playlist"
// @Success 200 {object} CreatePlayListRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/playlist/ [post]
func CreatePlayList(c *gin.Context) {
	ctx := c.Request.Context()
	var newVideoPlayList mongo_modals.VideoPlayListModal
	if err := c.ShouldBind(&newVideoPlayList); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid playlist payload", nil)
		return
	}
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	_time := time.Now()
	newVideoPlayList.CreatedAt = _time
	newVideoPlayList.UpdatedAt = _time
	newVideoPlayList.CreatedByUser = _id
	ins_res, ins_err := database_connections.MONGO_COLLECTIONS.VideoPlayList.InsertOne(ctx, newVideoPlayList)
	if ins_err != nil {
		resp_err, is_known := database_utils.GetDBErrorString(ins_err)
		if !is_known {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update", nil)
		} else {
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", resp_err, nil)
		}
		return
	}
	newVideoPlayList.ID = ins_res.InsertedID.(primitive.ObjectID)
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", newVideoPlayList)
}

type PlaylistVideoReqStruct struct {
	PlaylistId string                       `json:"playlist_id" binding:"required"`
	EnrollDays int16                        `json:"enroll_days,omitempty" binding:"required" bson:"enroll_days,omitempty"`
	Price      int64                        `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	Ids        []mongo_modals.VideosInOrder `json:"videos_ids" binding:"required"`
}

// @BasePath /api/
// @Summary Update playlist videos
// @Schemes
// @Description api to update playlist videos
// @Tags Playlist
// @Accept json
// @Produce json
// @Param video_list body PlaylistVideoReqStruct true "Playlist videos"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/playlist/ [put]
func UpdatePlayList(c *gin.Context) {
	ctx := c.Request.Context()
	var videos_info PlaylistVideoReqStruct
	if err := c.ShouldBind(&videos_info); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid playlist payload", nil)
		return
	}
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)
	if payload.Data.ID == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	playlist_id, _id_err := primitive.ObjectIDFromHex(videos_info.PlaylistId)
	if videos_info.PlaylistId == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	where := bson.M{
		"_id":             playlist_id,
		"created_by_user": _id,
	}
	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{"_id": playlist_id}
	}
	res, err := database_connections.MONGO_COLLECTIONS.VideoPlayList.UpdateOne(
		ctx,
		where,
		bson.M{
			"$set": bson.M{
				"videos_ids":  videos_info.Ids,
				"updatedAt":   time.Now(),
				"price":       videos_info.Price,
				"enroll_days": videos_info.EnrollDays,
			},
		},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("Failed to update user data")
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update data", nil)
		return
	}
	var response_data = make(map[string]interface{})
	response_data["updated_count"] = res.ModifiedCount
	response_data["match_count"] = res.MatchedCount
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Updated successfully", response_data)

}

type RemovePlayListReqStruct struct {
	PlaylistId string `json:"playlist_id" binding:"required"`
}

// @BasePath /api/
// @Summary Delete playlist
// @Schemes
// @Description api to Delete playlist
// @Tags Playlist
// @Accept json
// @Produce json
// @Param playlists body RemovePlayListReqStruct true "Playlist id"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/playlist/ [delete]
func RemovePlayList(c *gin.Context) {
	ctx := c.Request.Context()
	var playlists RemovePlayListReqStruct
	if err := c.ShouldBind(&playlists); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid playlist payload", nil)
		return
	}
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)
	if payload.Data.ID == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	playlist_id, _id_err := primitive.ObjectIDFromHex(playlists.PlaylistId)
	if playlists.PlaylistId == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	where := bson.M{
		"_id":             playlist_id,
		"created_by_user": _id,
	}
	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{"_id": playlist_id}
	}

	if _, err := database_connections.MONGO_COLLECTIONS.VideoPlayList.DeleteOne(ctx, where); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Errorln("delete playlists")
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to delete playlist", nil)
		return
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Deleted successfully", nil)
}

type GetAllSubscriptionPackagesStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.VideoPlayListSubscriptionPackageModal `json:"data"`
}

// @BasePath /api/
// @Summary Get list of playlist subscription packages
// @Schemes
// @Description api to fetch playlist subscription packages
// @Tags Playlist Subscription Package
// @Accept json
// @Produce json
// @Success 200 {object} GetAllSubscriptionPackagesStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/subscription_package/ [get]
func GetAllSubscriptionPackages(c *gin.Context) {
	// If Admin: All subscription package created by user
	// If Super Admin: All subscription package created by all user

	ctx := c.Request.Context()
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	where := bson.M{
		"created_by": _id,
	}
	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{}
	}
	cursor, err := database_connections.MONGO_COLLECTIONS.VideoPlayListSubscriptionPackage.Find(ctx, where)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.WithFields(log.Fields{
				"err": err,
			}).Errorln("Failed to load session data")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
		return
	} else {
		defer cursor.Close(ctx)
		var playlistSubscriptionPackagesData []mongo_modals.VideoPlayListSubscriptionPackageModal = []mongo_modals.VideoPlayListSubscriptionPackageModal{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var playlistSubscriptionPackageData mongo_modals.VideoPlayListSubscriptionPackageModal
			if err = cursor.Decode(&playlistSubscriptionPackageData); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
				return
			}
			playlistSubscriptionPackagesData = append(playlistSubscriptionPackagesData, playlistSubscriptionPackageData)
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", playlistSubscriptionPackagesData)
		return
	}

}

// @BasePath /api/
// @Summary Create new playlist subscription package
// @Schemes
// @Description api to create new empty playlist subscription package
// @Tags Playlist Subscription Package
// @Accept json
// @Produce json
// @Param new_playlist_subscription body mongo_modals.VideoPlayListSubscriptionPackageModal true "New Playlist Subscription Package"
// @Success 200 {object} GetAllSubscriptionPackagesStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/subscription_package/ [post]
func CreateSubscriptionPackage(c *gin.Context) {

	ctx := c.Request.Context()
	var newVideoPlayListSubscriptionPackage mongo_modals.VideoPlayListSubscriptionPackageModal
	if err := c.ShouldBind(&newVideoPlayListSubscriptionPackage); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid playlist payload", nil)
		return
	}
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	var id string = payload.Data.ID
	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if id == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	_time := time.Now()
	newVideoPlayListSubscriptionPackage.CreatedAt = _time
	newVideoPlayListSubscriptionPackage.UpdatedAt = _time
	newVideoPlayListSubscriptionPackage.CreatedBy = _id
	ins_res, ins_err := database_connections.MONGO_COLLECTIONS.VideoPlayListSubscriptionPackage.InsertOne(ctx, newVideoPlayListSubscriptionPackage)
	if ins_err != nil {
		resp_err, is_known := database_utils.GetDBErrorString(ins_err)
		if !is_known {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update", nil)
		} else {
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", resp_err, nil)
		}
		return
	}
	newVideoPlayListSubscriptionPackage.ID = ins_res.InsertedID.(primitive.ObjectID)
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", newVideoPlayListSubscriptionPackage)

}

type VideoPlayListSubscriptionPackageModalReqStruct struct {
	SubscriptionPackageId string   `json:"subscription_package_id" binding:"required"`
	PlaylistIds           []string `json:"playlists_ids" binding:"required"`
}

// @BasePath /api/
// @Summary Update playlist subscription packages
// @Schemes
// @Description api to update playlist subscription packages
// @Tags Playlist Subscription Package
// @Accept json
// @Produce json
// @Param subscription_packages body VideoPlayListSubscriptionPackageModalReqStruct true "Playlist subscription packages"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/subscription_package/ [put]
func UpdateSubscriptionPackage(c *gin.Context) {

	ctx := c.Request.Context()
	var playlists_info VideoPlayListSubscriptionPackageModalReqStruct
	if err := c.ShouldBind(&playlists_info); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid playlist payload", nil)
		return
	}
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)
	if payload.Data.ID == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	subscription_package_id, _id_err := primitive.ObjectIDFromHex(playlists_info.SubscriptionPackageId)
	if playlists_info.SubscriptionPackageId == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "UUID of user is not provided", _id_err)
		return
	}

	playlists_ids := []primitive.ObjectID{}
	for i := 0; i < len(playlists_info.PlaylistIds); i++ {
		playlist_id, _id_err := primitive.ObjectIDFromHex(playlists_info.PlaylistIds[i])
		if playlists_info.PlaylistIds[i] == "" || _id_err != nil {
			continue
		}
		playlists_ids = append(playlists_ids, playlist_id)
	}
	if len(playlists_ids) == 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "No video provided", nil)
	}

	where := bson.M{
		"_id":             subscription_package_id,
		"created_by_user": _id,
	}
	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{"_id": subscription_package_id}
	}
	res, err := database_connections.MONGO_COLLECTIONS.VideoPlayListSubscriptionPackage.UpdateOne(
		ctx,
		where,
		bson.M{
			"$set": bson.M{
				"playlists_ids": playlists_ids,
				"updatedAt":     time.Now(),
			},
		},
	)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("Failed to update user data")
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update data", nil)
		return
	}
	var response_data = make(map[string]interface{})
	response_data["updated_count"] = res.ModifiedCount
	response_data["match_count"] = res.MatchedCount
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Updated successfully", response_data)
}

func RemoveSubscriptionPackage(c *gin.Context) {

}

type VideoPlayListUserSubscriptionList struct {
	ID                      primitive.ObjectID                                    `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	UserID                  primitive.ObjectID                                    `json:"user_id,omitempty" binding:"required" bson:"user_id,omitempty"`
	Username                string                                                `json:"username,omitempty" binding:"required" bson:"username,omitempty"`
	PlaylistID              primitive.ObjectID                                    `json:"playlist_id,omitempty" binding:"required" bson:"playlist_id,omitempty"`
	Price                   int64                                                 `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	InitialSubscriptionDate time.Time                                             `json:"initial_subscription_date"  binding:"required" bson:"initial_subscription_date"`
	ExpireOn                time.Time                                             `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
	IsEnabled               bool                                                  `json:"is_enabled,omitempty"  binding:"required" bson:"is_enabled,omitempty"`
	Subscriptions           mongo_modals.SubsequentUserPlaylistSubscriptionStruct `json:"subscriptions,omitempty" binding:"required" bson:"subscriptions"`
	CreatedAt               time.Time                                             `json:"createdAt,omitempty" bson:"CreatedAt" swaggerignore:"true"`
	UpdatedAt               time.Time                                             `json:"updatedAt,omitempty" bson:"UpdatedAt" swaggerignore:"true"`
}

type GetUserPlaylistSubscriptionsRespStruct struct {
	my_modules.ResponseFormat
	Data []VideoPlayListUserSubscriptionList `json:"data"`
}

// @BasePath /api/
// @Summary Get video subscriptions for all user
// @Schemes
// @Description api to fetch video subscriptions for all user
// @Tags User paid subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} GetUserPlaylistSubscriptionsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/user_subscriptions/playlist/ [get]
func GetUserPlaylistSubscriptions(c *gin.Context) {

	unwindStage := bson.D{
		{"$unwind", "$subscriptions"},
	}
	sortStage := bson.D{{"$sort", bson.D{{"subscriptions.subscribed_on", 1}}}}
	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.Aggregate(c.Request.Context(), mongo.Pipeline{unwindStage, sortStage})
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
		return
	}
	var subscriptionsData []VideoPlayListUserSubscriptionList = []VideoPlayListUserSubscriptionList{}
	for cursor.Next(c.Request.Context()) {
		var subscriptionData VideoPlayListUserSubscriptionList
		if err = cursor.Decode(&subscriptionData); err != nil {
			log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
			// continue
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retriving user data", nil)
			return
		}
		subscriptionsData = append(subscriptionsData, subscriptionData)
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", subscriptionsData)
}

type StudyMaterialUserUserSubscriptionList struct {
	ID                      primitive.ObjectID                                         `json:"_id,omitempty" bson:"_id,omitempty" swaggerignore:"true"`
	UserID                  primitive.ObjectID                                         `json:"user_id,omitempty" binding:"required" bson:"user_id,omitempty"`
	Username                string                                                     `json:"username,omitempty" binding:"required" bson:"username,omitempty"`
	SubscriptionPackageId   primitive.ObjectID                                         `json:"subscription_package_id,omitempty" binding:"required" bson:"subscription_package_id,omitempty"`
	StudyMaterialID         primitive.ObjectID                                         `json:"study_material_id,omitempty" binding:"required" bson:"study_material_id,omitempty"`
	Price                   int64                                                      `json:"price,omitempty"  binding:"required" bson:"price,omitempty"`
	InitialSubscriptionDate time.Time                                                  `json:"initial_subscription_date"  binding:"required" bson:"initial_subscription_date"`
	ExpireOn                time.Time                                                  `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
	IsEnabled               bool                                                       `json:"is_enabled,omitempty"  binding:"required" bson:"is_enabled,omitempty"`
	Subscriptions           mongo_modals.SubsequentStudyMaterialUserSubscriptionStruct `json:"subscriptions,omitempty" binding:"required" bson:"subscriptions"`
	CreatedAt               time.Time                                                  `json:"createdAt,omitempty" bson:"CreatedAt" swaggerignore:"true"`
	UpdatedAt               time.Time                                                  `json:"updatedAt,omitempty" bson:"UpdatedAt" swaggerignore:"true"`
}

type GetUserDocSubscriptionsRespStruct struct {
	my_modules.ResponseFormat
	Data []StudyMaterialUserUserSubscriptionList `json:"data"`
}

// @BasePath /api/
// @Summary Get document subscriptions for all user
// @Schemes
// @Description api to fetch document subscriptions for all user
// @Tags User paid subscriptions
// @Accept json
// @Produce json
// @Success 200 {object} GetUserDocSubscriptionsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/user_subscriptions/doc/ [get]
func GetUserDocSubscriptions(c *gin.Context) {
	unwindStage := bson.D{
		{"$unwind", "$subscriptions"},
	}
	sortStage := bson.D{{"$sort", bson.D{{"subscriptions.subscribed_on", 1}}}}
	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.Aggregate(c.Request.Context(), mongo.Pipeline{unwindStage, sortStage})
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
		return
	}
	var subscriptionsData []StudyMaterialUserUserSubscriptionList = []StudyMaterialUserUserSubscriptionList{}
	for cursor.Next(c.Request.Context()) {
		var subscriptionData StudyMaterialUserUserSubscriptionList
		if err = cursor.Decode(&subscriptionData); err != nil {
			log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
			// continue
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retriving user data", nil)
			return
		}
		subscriptionsData = append(subscriptionsData, subscriptionData)
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", subscriptionsData)
}
