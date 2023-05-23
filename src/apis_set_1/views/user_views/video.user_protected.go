package user_views

import (
	"cca/src/configs"
	"cca/src/database/database_connections"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	razorpay "github.com/razorpay/razorpay-go"
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
			video.VideoDecryptionKey = ""
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
	// 646b8cc5a1d3db25782498df
	// {videos_ids:{"$elemMatch":{video_id:ObjectId('646a499150b48a442558cf7e')}}}
	// {"videos_ids.video_id":ObjectId('646a499150b48a442558cf7e')}
	var videoPlaylist mongo_modals.VideoPlayListModal
	err = database_connections.MONGO_COLLECTIONS.VideoPlayList.FindOne(ctx, bson.M{
		"videos_ids.video_id": objID,
		"is_live":             true,
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

	key, block_size := my_modules.EncryptWithPKCS(appBuildInfo.AppSecret, videoData.VideoDecryptionKey)
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "found", map[string]interface{}{
		"key":        key,
		"block_size": block_size,
	})
}

func createOrderForPaymentToEnrollCourse(user_id primitive.ObjectID, user_subscriptions_ids []primitive.ObjectID, amount int64, receipt_id string) (mongo_modals.PaymentOrderModal, error) {
	client := razorpay.NewClient(configs.EnvConfigs.RAZORPAY_KEY_ID, configs.EnvConfigs.RAZORPAY_KEY_SECRET)

	h := sha1.New()
	h.Write([]byte(receipt_id))
	sha1_hash := hex.EncodeToString(h.Sum(nil))

	data := map[string]interface{}{
		"amount":          amount * 100,
		"currency":        "INR",
		"receipt":         sha1_hash,
		"partial_payment": false,
		"notes": map[string]interface{}{
			"receipt_id": receipt_id,
		},
	}
	order_data, err := client.Order.Create(data, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"receipt_id": receipt_id,
			"user_id":    user_id,
			"err":        err,
		}).Errorln("Create razorpay PaymentOrder")
		return mongo_modals.PaymentOrderModal{}, err
	}

	_time := time.Now()
	ins_data := mongo_modals.PaymentOrderModal{
		UserID:               user_id,
		UserSubscriptionsIDs: user_subscriptions_ids,
		OrderID:              order_data["id"].(string),
		Amount:               amount,
		PaymentStatus:        false,
		CreatedAt:            _time,
		UpdatedAt:            _time,
	}
	order_tbl, ins_err := database_connections.MONGO_COLLECTIONS.PaymentOrder.InsertOne(context.Background(), ins_data)
	if ins_err != nil {
		log.WithFields(log.Fields{
			"receipt_id": receipt_id,
			"user_id":    user_id,
			"ins_err":    ins_err,
		}).Errorln("Insert PaymentOrder")
		return mongo_modals.PaymentOrderModal{}, ins_err
	}
	ins_data.ID = order_tbl.InsertedID.(primitive.ObjectID)
	return ins_data, nil
}

func getAndUpdateOrderStatusForPaymentToEnrollCourse(user_id primitive.ObjectID, order_id primitive.ObjectID) (mongo_modals.PaymentOrderModal, error) {
	client := razorpay.NewClient(configs.EnvConfigs.RAZORPAY_KEY_ID, configs.EnvConfigs.RAZORPAY_KEY_SECRET)

	var paymentForOrder mongo_modals.PaymentOrderModal
	err := database_connections.MONGO_COLLECTIONS.PaymentOrder.FindOne(context.Background(), bson.M{
		"_id":     order_id,
		"user_id": user_id,
	}).Decode(&paymentForOrder)

	if err != nil {
		log.WithFields(log.Fields{
			"_id":     order_id,
			"user_id": user_id,
			"err":     err,
		}).Errorln("Find PaymentOrder")
		return mongo_modals.PaymentOrderModal{}, err
	}

	order_data, err := client.Order.Fetch(paymentForOrder.OrderID, nil, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"_id":     order_id,
			"user_id": user_id,
			"err":     err,
		}).Errorln("Fetch razorpay PaymentOrder")
		return mongo_modals.PaymentOrderModal{}, err
	}

	log.Debugln(order_data)
	_time := time.Now()
	paymentForOrder.PaymentStatus = order_data["status"] == "paid"
	if paymentForOrder.PaymentStatus {
		update_status, err := database_connections.MONGO_COLLECTIONS.PaymentOrder.UpdateOne(
			context.Background(),
			bson.M{
				"_id":     order_id,
				"user_id": user_id,
			},
			bson.M{
				"$set": bson.M{
					"payment_status": paymentForOrder.PaymentStatus,
					"UpdatedAt":      _time,
				},
			},
		)
		if err != nil {
			log.WithFields(log.Fields{
				"_id":     order_id,
				"user_id": user_id,
				"err":     err,
			}).Errorln("Update PaymentOrder")
		} else if update_status.ModifiedCount == 0 {
			log.WithFields(log.Fields{
				"_id":            order_id,
				"user_id":        user_id,
				"match_count":    update_status.MatchedCount,
				"modified_count": update_status.ModifiedCount,
			}).Errorln("Update PaymentOrder")
		}
	}

	return paymentForOrder, nil
}

type EnrollToCourseReqStruct struct {
	PlaylistIDs           []string `json:"playlist_ids" binding:"required"`
	SubscriptionPackageID string   `json:"subscription_package_id"`
}

type EnrollToCourseRespStruct struct {
	my_modules.ResponseFormat
	Data mongo_modals.PaymentOrderModal `json:"data"`
}

// @BasePath /api/
// @Summary enroll to course
// @Schemes
// @Description api to enroll playlist/subscription
// @Tags Customer side
// @Accept json
// @Produce json
// @Param video_id body EnrollToCourseReqStruct true "Video ID"
// @Success 200 {object} EnrollToCourseRespStruct
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

	user_subscriptions := make(map[primitive.ObjectID]int64)

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
							"price":         videoPlaylist.Price,
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
				} else {
					user_subscriptions[videoPlaylistUserSubscription.ID] = videoPlaylist.Price
				}
			} else {
				user_sub_tbl, ins_err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.InsertOne(context.Background(), mongo_modals.VideoPlayListUserSubscriptionModal{
					UserID:                  user_id,
					PlaylistID:              playlis_id,
					InitialSubscriptionDate: _time,
					IsEnabled:               false,
					ExpireOn:                _time.AddDate(0, 0, int(videoPlaylist.EnrollDays)),
					Price:                   videoPlaylist.Price,
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
					// my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Regestering new user while marking active", nil)
					continue
				}
				videoPlaylistUserSubscription.ID = user_sub_tbl.InsertedID.(primitive.ObjectID)
				user_subscriptions[videoPlaylistUserSubscription.ID] = videoPlaylist.Price
			}
		}
	}

	var total_amount int64 = 0
	subscriptions_ids_hex := []string{}
	subscriptions_ids := []primitive.ObjectID{}

	for subscription_id, amount := range user_subscriptions {
		total_amount += amount
		subscriptions_ids_hex = append(subscriptions_ids_hex, subscription_id.Hex())
		subscriptions_ids = append(subscriptions_ids, subscription_id)
	}
	receipt_id := fmt.Sprintf("%s-%d", strings.Join(subscriptions_ids_hex, ","), total_amount)

	order_data, order_err := createOrderForPaymentToEnrollCourse(user_id, subscriptions_ids, total_amount, receipt_id)
	if order_err != nil {
		log.WithFields(log.Fields{
			"order_err":  order_err,
			"receipt_id": receipt_id,
		}).Errorln("Error creating order")
		// need to roll back subscription changes
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in creating order", nil)
		return
	}

	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", order_data)

}

// @BasePath /api
// @Summary Confirm payment on Order
// @Schemes
// @Description allow confirm payment on order created & payment status will be verified on server side
// @Tags Customer side
// @Produce json
// @Param order_id query string true "Order ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/confirm_payment_for_subscription/ [get]
func PaymentConfirmationForSubscription(c *gin.Context) {
	order_id_str := c.Query("order_id")
	if order_id_str == "" {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "'order_id' is not provided", nil)
	}

	order_id, _id_err := primitive.ObjectIDFromHex(order_id_str)
	if order_id_str == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid order id payload", _id_err)
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

	order_data, err := getAndUpdateOrderStatusForPaymentToEnrollCourse(user_id, order_id)

	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error while verifying payment status", nil)
		return
	}

	if true {
		_time := time.Now()
		update_status, err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.UpdateMany(
			context.Background(),
			bson.M{
				"_id":     bson.M{"$in": order_data.UserSubscriptionsIDs},
				"user_id": user_id,
			},
			bson.M{
				"$set": bson.M{
					"is_enabled": order_data.PaymentStatus,
					"UpdatedAt":  _time,
				},
			},
		)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id":  user_id,
				"order_id": order_data.UserSubscriptionsIDs,
				"err":      err,
			}).Errorln("Failed to enroll to existing subscription")
		} else if update_status.ModifiedCount == 0 {
			log.WithFields(log.Fields{
				"user_id":        user_id,
				"order_id":       order_data.UserSubscriptionsIDs,
				"match_count":    update_status.MatchedCount,
				"modified_count": update_status.ModifiedCount,
			}).Errorln("Failed to enroll to existing subscription")
		}
	}

	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", order_data)
}

type GetUserSubscriptionListStruct struct {
	ID                    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SubscriptionPackageId primitive.ObjectID `json:"subscription_package_id,omitempty" bson:"subscription_package_id,omitempty"`
	PlaylistID            primitive.ObjectID `json:"playlist_id,omitempty"  bson:"playlist_id,omitempty"`
	ExpireOn              time.Time          `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
}

type GetUserSubscriptionListRespPayload struct {
	my_modules.ResponseFormat
	Data GetUserSubscriptionListStruct `json:"data"`
}

// @BasePath /api/
// @Summary Get user subscriptions
// @Schemes
// @Description api to get user subscriptions
// @Tags Customer side
// @Produce json
// @Success 200 {object} GetUserSubscriptionListRespPayload
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_user_subscriptions/ [get]
func GetUserSubscriptionList(c *gin.Context) {
	ctx := c.Request.Context()

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

	where := bson.M{
		"is_enabled": true,
		"user_id":    user_id,
	}
	cursor, err := database_connections.MONGO_COLLECTIONS.VideoPlayListUserSubscription.Find(ctx, where)
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
		var user_subscriptions []GetUserSubscriptionListStruct = []GetUserSubscriptionListStruct{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var user_subscription mongo_modals.VideoPlayListUserSubscriptionModal
			if err = cursor.Decode(&user_subscription); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
				return
			}
			user_subscriptions = append(user_subscriptions, GetUserSubscriptionListStruct{
				ID:                    user_subscription.ID,
				PlaylistID:            user_subscription.PlaylistID,
				SubscriptionPackageId: user_subscription.SubscriptionPackageId,
				ExpireOn:              user_subscription.ExpireOn,
			})

		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", user_subscriptions)
		return
	}

}

func GetAvailableSubscriptionPackages(c *gin.Context) {

}

func GetPlaylistAvailableOnSubscription(c *gin.Context) {

}
