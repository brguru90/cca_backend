package user_views

import (
	"cca/src/database/database_connections"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getUserSubscriptions(c *gin.Context) (map[primitive.ObjectID]bool, error) {
	ctx := c.Request.Context()
	playlistIDs := make(map[primitive.ObjectID]bool)
	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		return playlistIDs, fmt.Errorf("unable to get user info")
	}
	user_id, _id_err := primitive.ObjectIDFromHex(payload.Data.ID)

	if payload.Data.ID == "" || _id_err != nil {
		return playlistIDs, fmt.Errorf("UUID of user is not provided")
	}

	cursor, err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.Find(ctx, bson.M{
		"user_id":    user_id,
		"expired_on": bson.M{"$gt": time.Now()},
	})
	if err != nil {
		if err != mongo.ErrNoDocuments {
			log.WithFields(log.Fields{
				"err": err,
			}).Errorln("Failed to load session data")
		}
		return playlistIDs, fmt.Errorf("no documents")
	} else {
		defer cursor.Close(ctx)
		for cursor.Next(c.Request.Context()) {
			var userSubscription mongo_modals.VideoPlayListUserSubscriptionModal
			if err = cursor.Decode(&userSubscription); err != nil {
				continue
			}
			playlistIDs[userSubscription.PlaylistID] = true
		}

		return playlistIDs, nil
	}
}

type VideoPlayListModal struct {
	mongo_modals.VideoPlayListModal
	Paid bool `json:"paid"  binding:"required"`
}

type GetAllPlayListsRespStruct struct {
	my_modules.ResponseFormat
	Data []VideoPlayListModal `json:"data"`
}

// @BasePath /api/
// @Summary Get list of playlist
// @Schemes
// @Description api to fetch existing playlist
// @Tags Customer side
// @Accept json
// @Produce json
// @Success 200 {object} GetAllPlayListsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_playlists/ [get]
func GetAvailablePlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	subscribedPlaylistIDs, err := getUserSubscriptions(c)
	where := bson.M{
		"is_live": true,
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
		var playlistsData []VideoPlayListModal = []VideoPlayListModal{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var playlistData mongo_modals.VideoPlayListModal
			if err = cursor.Decode(&playlistData); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
				return
			}
			if len(playlistData.VideosIDs) == 0 {
				continue
			}
			temp := VideoPlayListModal{}
			paid, ok := subscribedPlaylistIDs[playlistData.ID]
			if ok {
				temp.Paid = paid
			} else {
				temp.Paid = false
			}
			temp.ID = playlistData.ID
			temp.Title = playlistData.Title
			temp.Description = playlistData.Description
			temp.CreatedByUser = playlistData.CreatedByUser
			temp.IsLive = playlistData.IsLive
			temp.VideosIDs = playlistData.VideosIDs
			temp.Price = playlistData.Price
			temp.CreatedAt = playlistData.CreatedAt
			temp.UpdatedAt = playlistData.UpdatedAt
			playlistsData = append(playlistsData, temp)
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", playlistsData)
		return
	}
}

type GetVideosReqStruct struct {
	VideoIDs []string `json:"video_ids" binding:"required"`
}

// @BasePath /api/
// @Summary Get list of videos
// @Schemes
// @Description api to fetch videos by providing ID
// @Tags Customer side
// @Accept json
// @Produce json
// @Param videos body GetVideosReqStruct true "Videos IDS"
// @Success 200 {object} GetAllPlayListsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_videos/ [post]
func GetVideos(c *gin.Context) {
	// Todo, all the videos belong to specified playlist

	ctx := c.Request.Context()
	var videosFrom GetVideosReqStruct
	if err := c.ShouldBindJSON(&videosFrom); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}

	if len(videosFrom.VideoIDs) == 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Video IDs are empty", nil)
	}

	video_ids := []primitive.ObjectID{}
	for i := 0; i < len(videosFrom.VideoIDs); i++ {
		video_id, _id_err := primitive.ObjectIDFromHex(videosFrom.VideoIDs[i])
		if videosFrom.VideoIDs[i] == "" || _id_err != nil {
			continue
		}
		video_ids = append(video_ids, video_id)
	}
	if len(video_ids) == 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Video IDs are empty", nil)
	}

	where := bson.M{
		"is_live": true,
		"_id":     bson.M{"$in": video_ids},
	}
	cursor, err := database_connections.MONGO_COLLECTIONS.VideoUploads.Find(ctx, where)
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
		var videos []mongo_modals.VideoUploadModal = []mongo_modals.VideoUploadModal{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var video mongo_modals.VideoUploadModal
			if err = cursor.Decode(&video); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
				return
			}
			videos = append(videos, video)

		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", videos)
		return
	}

}

type VideoStreamKeyReqStruct struct {
	AppID   string `json:"app_id" binding:"required"`
	VideoId string `json:"video_id" binding:"required"`
}

// @BasePath /api/
// @Summary get video decode key
// @Schemes
// @Description api to get video decryption key for hls stream
// @Tags Customer side
// @Accept json
// @Produce json
// @Param video_id body VideoStreamKeyReqStruct true "Video ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_stream_key/ [post]
func GetStreamKey(c *gin.Context) {
	ctx := c.Request.Context()
	var videoStreamInfo VideoStreamKeyReqStruct
	if err := c.ShouldBind(&videoStreamInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid payload", nil)
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

	objID, err := primitive.ObjectIDFromHex(videoStreamInfo.VideoId)
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid key", nil)
		return
	}

	var appBuildInfo mongo_modals.AppBuildRegistrationModal
	err = database_connections.MONGO_COLLECTIONS.AppBuildRegistration.FindOne(ctx, bson.M{
		"app_id": videoStreamInfo.AppID,
	}).Decode(&appBuildInfo)
	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Error in finding app id", nil)
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

	var videoPlaylist mongo_modals.VideoPlayListModal
	objIDArr := []primitive.ObjectID{objID}
	err = database_connections.MONGO_COLLECTIONS.VideoPlayList.FindOne(ctx, bson.M{
		"videos_ids": bson.M{"$in": objIDArr},
		"is_live":    true,
		// "access_level": access_level,
	}).Decode(&videoPlaylist)
	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in finding video in playlist", nil)
	}

	var videoPlaylistUserSubscriptions mongo_modals.VideoPlayListUserSubscriptionModal
	playlistIDs := []primitive.ObjectID{videoPlaylist.ID}
	err = database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.FindOne(ctx, bson.M{
		"user_id":     user_id,
		"playlist_id": bson.M{"$in": playlistIDs},
		"expired_on":  bson.M{"$gt": time.Now()},
	}).Decode(&videoPlaylistUserSubscriptions)
	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Error in finding package in user subscription ", nil)
	}

	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "found", map[string]interface{}{
		"key": my_modules.EncryptAES(videoData.VideoDecryptionKey, appBuildInfo.AppSecret),
	})
}

type EnrollToCourseReqStruct struct {
	PlaylistIDs           []string `json:"playlist_ids" binding:"required"`
	SubscriptionPackageID string   `json:"subscription_package_id"`
}

// @BasePath /api/
// @Summary enroll to course
// @Schemes
// @Description api to enroll playlist/subscription
// @Tags Customer side
// @Accept json
// @Produce json
// @Param video_id body EnrollToCourseReqStruct true "Video ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/enroll_to_course/ [post]
func EnrollToCourse(c *gin.Context) {
	var subscriptionInfo EnrollToCourseReqStruct
	if err := c.ShouldBind(&subscriptionInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid payload", nil)
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
	subscription_package_id, _id_err := primitive.ObjectIDFromHex(subscriptionInfo.SubscriptionPackageID)

	_time := time.Now()
	for i := 0; i < len(subscriptionInfo.PlaylistIDs); i++ {
		playlis_id, _id_err := primitive.ObjectIDFromHex(subscriptionInfo.PlaylistIDs[i])
		if subscriptionInfo.PlaylistIDs[i] == "" || _id_err != nil {
			continue
		}

		var videoPlaylist mongo_modals.VideoPlayListModal
		err := database_connections.MONGO_COLLECTIONS.VideoPlayList.FindOne(context.Background(), bson.M{
			"_id": playlis_id,
		}).Decode(&videoPlaylist)

		if err != nil {
			log.WithFields(log.Fields{
				"user_id":    user_id,
				"playlis_id": playlis_id,
				"err":        err,
			}).Errorln("Failed to find playlist")
			continue
		}

		{
			var videoPlaylistUserSubscription mongo_modals.VideoPlayListUserSubscriptionModal
			err = database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.FindOne(context.Background(), bson.M{
				"user_id":     user_id,
				"playlist_id": playlis_id,
				// AmountPaid: ,
			}).Decode(&videoPlaylistUserSubscription)

			if err == nil {
				//? if its a renewal
				//!Warning, enroll days will be not add up to existing subscriptions
				previous_subscription := videoPlaylistUserSubscription.Subscriptions
				previous_subscription = append(previous_subscription, mongo_modals.SubsequentUserPlaylistSubscriptionStruct{
					SubscribedOn:   _time,
					DurationInDays: int(videoPlaylist.EnrollDays),
				})
				update_status, err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.UpdateOne(
					context.Background(),
					bson.M{
						"user_id":     user_id,
						"playlist_id": playlis_id,
					},
					bson.M{
						"$set": bson.M{
							"expired_on":    _time.AddDate(0, 0, int(videoPlaylist.EnrollDays)),
							"is_enabled":    false,
							"subscriptions": previous_subscription,
							"UpdatedAt":     _time,
						},
					},
				)
				if err != nil {
					log.WithFields(log.Fields{
						"user_id":    user_id,
						"playlis_id": playlis_id,
						"err":        err,
					}).Errorln("Failed to enroll to existing subscription")
				} else if update_status.ModifiedCount == 0 {
					log.WithFields(log.Fields{
						"user_id":        user_id,
						"playlis_id":     playlis_id,
						"match_count":    update_status.MatchedCount,
						"modified_count": update_status.ModifiedCount,
					}).Errorln("Failed to enroll to existing subscription")
				}
			} else {
				_, ins_err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.InsertOne(context.Background(), mongo_modals.VideoPlayListUserSubscriptionModal{
					UserID:                  user_id,
					PlaylistID:              playlis_id,
					InitialSubscriptionDate: _time,
					IsEnabled:               false,
					ExpireOn:                _time.AddDate(0, 0, int(videoPlaylist.EnrollDays)),
					Subscriptions: []mongo_modals.SubsequentUserPlaylistSubscriptionStruct{
						{
							SubscribedOn:   _time,
							DurationInDays: int(videoPlaylist.EnrollDays),
							// AmountPaid: ,
						},
					},
					SubscriptionPackageId: subscription_package_id,
					CreatedAt:             _time,
					UpdatedAt:             _time,
				})
				if ins_err != nil {
					log.WithFields(log.Fields{
						"ins_err": ins_err,
					}).Errorln("Error in inserting data to mongo users")
					my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Regestering new user while marking active", nil)
					return
				}
			}
			my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", nil)
		}

	}

}

func GetAvailableSubscriptionPackages(c *gin.Context) {

}

func PaymentConfirmationForSubscription(c *gin.Context) {

}

func GetPlaylistAvailableOnSubscription(c *gin.Context) {

}
