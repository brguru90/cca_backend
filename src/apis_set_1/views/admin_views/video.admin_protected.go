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
	protected_video := fmt.Sprintf("%s/private/video", configs.EnvConfigs.VIDEO_UPLOAD_PATH)
	upload_path := fmt.Sprintf("%s/original_video", protected_video)
	if err := os.MkdirAll(upload_path, 0755); err != nil {
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
	dst_video_file_path := fmt.Sprintf("%s.mp4", dst_file_path)
	preview_image_file_part := strings.Split(uploadForm.PreviewImageFile.Filename, ".")
	if len(preview_image_file_part) <= 1 {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "File name doesn't contain file extension", nil)
		return
	}
	dst_preview_image_file_path := fmt.Sprintf("%s.%s", dst_file_path, preview_image_file_part[len(preview_image_file_part)-1])

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
		LinkToOriginalVideo:     dst_video_file_path,
		LinkToVideoPreviewImage: dst_preview_image_file_path,
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
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "", nil)
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
	if _, err := database_connections.REDIS_DB_CONNECTION.Get(ctx, "video_stream_generation_in_progress").Result(); err == nil {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "warning", "Process is already running", nil)
		return
	}

	var videoStreamInfo VideoStreamReqStruct
	if err := c.ShouldBind(&videoStreamInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}

	var videos_doc_id []primitive.ObjectID
	for i := 0; i < len(videoStreamInfo.Id); i++ {
		objID, err := primitive.ObjectIDFromHex(videoStreamInfo.Id[i])
		if err == nil {
			videos_doc_id = append(videos_doc_id, objID)
		}
	}

	unprotected_video := fmt.Sprintf("%s/public/video", configs.EnvConfigs.VIDEO_UPLOAD_PATH)

	var err error
	var cursor *mongo.Cursor
	defer cursor.Close(ctx)
	cursor, err = database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, bson.M{"_id": bson.M{"$in": videos_doc_id}})
	if err != nil {
		if err != context.Canceled {
			log.WithFields(log.Fields{
				"error": err,
			}).Errorln("QueryRow failed ==>")
		}
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No record found", nil)
	} else {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Started video stream creation", nil)
		// c.Abort()
		// conn, _, err := c.Writer.Hijack()
		// if err == nil {
		// 	conn.Close()
		// }
		go func(ids []primitive.ObjectID) {
			ctx, _ := context.WithTimeout(context.Background(), time.Hour*1)
			cursor, err := database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
			if err == nil {
				defer cursor.Close(ctx)
			}
			if err := database_connections.RedisPoolSet("video_stream_generation_in_progress", "value", 60*time.Minute); err != nil {
				log.WithFields(log.Fields{
					"Error": err,
				}).Panic("Unable to write into redis pool")
			}
			defer func() {
				if err := database_connections.RedisPoolDel("video_stream_generation_in_progress"); err != nil {
					log.WithFields(log.Fields{
						"Error": err,
					}).Panic("Unable to write into redis pool")
				}
			}()
			for cursor.Next(ctx) {
				var videoData mongo_modals.VideoUploadModal
				if err = cursor.Decode(&videoData); err != nil {
					log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
					return
				}
				path_split := strings.Split(videoData.LinkToOriginalVideo, "/")
				file_full_name := strings.Split(path_split[len(path_split)-1], ".")
				file_name := strings.Join(file_full_name[:len(file_full_name)-1], "_")
				// my_modules.UploadVideoForStream(path_split[len(path_split)-1], unprotected_video, v.LinkToOriginalVideo)
				link_to_video_stream := ""
				video_decryption_key := ""
				var data my_modules.UploadedVideoInfoStruct
				if data, err = my_modules.UploadVideoForStream(videoData.ID.Hex(), unprotected_video, file_name, videoData.LinkToOriginalVideo); err == nil {
					link_to_video_stream = data.StreamGeneratedLocation
					video_decryption_key = data.DecryptionKey
				} else {
					os.Remove(data.OutputDir)
				}

				log.WithFields(log.Fields{
					"link_to_video_stream": link_to_video_stream,
					"video_decryption_key": video_decryption_key,
				}).Debugln("Unable to write into redis pool")

				database_connections.MONGO_COLLECTIONS.VideoUploads.UpdateOne(
					context.Background(),
					bson.M{
						"_id": videoData.ID,
					},
					bson.M{
						"$set": bson.M{"link_to_video_stream": link_to_video_stream, "video_decryption_key": video_decryption_key},
					},
				)
			}
		}(videos_doc_id)
	}

}

type VideoStreamKeyReqStruct struct {
	VideoId string `json:"video_id" binding:"required"`
}

// @BasePath /api/
// @Summary get video decode key
// @Schemes
// @Description api to get video decryption key for hls stream
// @Tags Manage Videos
// @Accept json
// @Produce json
// @Param video_id body VideoStreamKeyReqStruct true "Video ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/get_stream_key/ [post]
func GetStreamKey(c *gin.Context) {
	ctx := c.Request.Context()
	var videoStreamInfo VideoStreamKeyReqStruct
	if err := c.ShouldBind(&videoStreamInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}

	objID, err := primitive.ObjectIDFromHex(videoStreamInfo.VideoId)
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid key", nil)
		return
	}

	var videoData mongo_modals.VideoUploadModal
	err = database_connections.MONGO_COLLECTIONS.VideoUploads.FindOne(ctx, bson.M{
		"_id": objID,
		// "access_level": access_level,
	}).Decode(&videoData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.WithFields(log.Fields{
				"Error": err,
				"Email": videoData.ID,
			}).Warning("Error in finding user email")
			my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No video matched to specified key", nil)
			return
		}
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in finding video", nil)
		return
	}

	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "found", map[string]interface{}{
		"key": videoData.VideoDecryptionKey,
	})

}

// @BasePath /api/
// @Summary get all uploaded videos
// @Schemes
// @Description api to get the list of all the videos uploaded by the logged user
// @Tags Manage Videos
// @Produce json
// @Success 200 {object} my_modules.ResponseFormat
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

		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", map[string]interface{}{
			"list": videosList,
		})
	}
}

// @BasePath /api/
// @Summary Get list of playlist
// @Schemes
// @Description api to fetch existing playlist
// @Tags Playlist
// @Accept json
// @Produce json
// @Success 200 {object} my_modules.ResponseFormat
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

// @BasePath /api/
// @Summary Create new playlist
// @Schemes
// @Description api to create new empty playlist
// @Tags Playlist
// @Accept json
// @Produce json
// @Param new_playlist_data body mongo_modals.VideoPlayListModal true "New Playlist"
// @Success 200 {object} my_modules.ResponseFormat
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
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", resp_err, nil)
		if !is_known {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update", nil)
		}
		return
	}
	newVideoPlayList.ID = ins_res.InsertedID.(primitive.ObjectID)
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", newVideoPlayList)
}

type PlaylistVideoReqStruct struct {
	PlaylistId string   `json:"playlist_id" binding:"required"`
	Ids        []string `json:"videos_ids" binding:"required"`
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

	video_ids := []primitive.ObjectID{}
	for i := 0; i < len(videos_info.Ids); i++ {
		video_id, _id_err := primitive.ObjectIDFromHex(videos_info.Ids[i])
		if videos_info.Ids[i] == "" || _id_err != nil {
			continue
		}
		video_ids = append(video_ids, video_id)
	}
	if len(video_ids) == 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "No video provided", nil)
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
				"videos_ids": video_ids,
				"updatedAt":  time.Now(),
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

func GetAllSubscriptionPackages(c *gin.Context) {
	// If Admin: All subscription package created by user
	// If Super Admin: All subscription package created by all user

}

func CreateSubscriptionPackage(c *gin.Context) {

}

func UpdateSubscriptionPackage(c *gin.Context) {

}
