package user_views

import (
	"cca/src/database/database_connections"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetAllUploadedStudyMaterialsRespStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.StudyMaterialsModal `json:"data"`
}

// @BasePath /api/
// @Summary get all uploaded documents
// @Schemes
// @Description api to get all uploaded documents
// @Tags Customer side(Study materials)
// @Produce json
// @Success 200 {object} GetAllUploadedStudyMaterialsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/study_materials/ [get]
func GetStudyMaterials(c *gin.Context) {
	ctx := c.Request.Context()

	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.StudyMaterial.Find(ctx, bson.M{
		"is_live": true,
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
		var docsList []mongo_modals.StudyMaterialsModal = []mongo_modals.StudyMaterialsModal{}
		for cursor.Next(c.Request.Context()) {
			var docData mongo_modals.StudyMaterialsModal
			if err = cursor.Decode(&docData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				// continue
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving user data", nil)
				return
			}
			docData.FileDecryptionKey = ""
			docsList = append(docsList, docData)
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", docsList)
	}
}

type GetUserStudyMaterialSubscriptionListStruct struct {
	ID                    primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	SubscriptionPackageId primitive.ObjectID `json:"subscription_package_id,omitempty" bson:"subscription_package_id,omitempty"`
	StudyMaterialID       primitive.ObjectID `json:"study_material_id,omitempty"  bson:"study_material_id,omitempty"`
	ExpireOn              time.Time          `json:"expired_on,omitempty" bson:"expired_on,omitempty"`
}

type GetUserStudyMaterialSubscriptionListRespPayload struct {
	my_modules.ResponseFormat
	Data GetUserStudyMaterialSubscriptionListStruct `json:"data"`
}

// @BasePath /api/
// @Summary Get user subscriptions for study material
// @Schemes
// @Description api to get  user subscriptions for study material
// @Tags Customer side(Study materials)
// @Produce json
// @Success 200 {object} GetUserStudyMaterialSubscriptionListRespPayload
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_user_study_material_subscriptions/ [get]
func GetUserStudyMaterialSubscriptionList(c *gin.Context) {
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
		"expired_on": bson.M{"$gt": time.Now()},
	}
	cursor, err := database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.Find(ctx, where)
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
		var user_subscriptions []GetUserStudyMaterialSubscriptionListStruct = []GetUserStudyMaterialSubscriptionListStruct{}
		// cursor.All(ctx,sessionsData);
		for cursor.Next(c.Request.Context()) {
			var user_subscription mongo_modals.StudyMaterialUserUserSubscriptionModal
			if err = cursor.Decode(&user_subscription); err != nil {
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving study material data", nil)
				return
			}
			user_subscriptions = append(user_subscriptions, GetUserStudyMaterialSubscriptionListStruct{
				ID:                    user_subscription.ID,
				StudyMaterialID:       user_subscription.StudyMaterialID,
				SubscriptionPackageId: user_subscription.SubscriptionPackageId,
				ExpireOn:              user_subscription.ExpireOn,
			})

		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", user_subscriptions)
		return
	}

}

type EnrollToStudyMaterialReqStruct struct {
	Documents_IDs []string `json:"document_ids" binding:"required"`
}

type EnrollToStudyMaterialRespStruct struct {
	my_modules.ResponseFormat
	Data mongo_modals.PaymentOrderModal `json:"data"`
}

// @BasePath /api/
// @Summary enroll to study materials
// @Schemes
// @Description api to study materials
// @Tags Customer side(Study materials)
// @Accept json
// @Produce json
// @Param doc_id body EnrollToStudyMaterialReqStruct true "Document ID"
// @Success 200 {object} EnrollToStudyMaterialRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/enroll_to_study_material/ [post]
func EnrollToStudyMaterial(c *gin.Context) {
	var subscriptionInfo EnrollToStudyMaterialReqStruct
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

	user_subscriptions := make(map[primitive.ObjectID]int64)

	_time := time.Now()
	for i := 0; i < len(subscriptionInfo.Documents_IDs); i++ {
		study_material_id, _id_err := primitive.ObjectIDFromHex(subscriptionInfo.Documents_IDs[i])
		if subscriptionInfo.Documents_IDs[i] == "" || _id_err != nil {
			continue
		}

		var studyMaterialList mongo_modals.StudyMaterialsModal
		err := database_connections.MONGO_COLLECTIONS.StudyMaterial.FindOne(context.Background(), bson.M{
			"_id": study_material_id,
		}).Decode(&studyMaterialList)

		if err != nil {
			log.WithFields(log.Fields{
				"user_id":           user_id,
				"study_material_id": study_material_id,
				"err":               err,
			}).Errorln("Failed to find StudyMaterial")
			continue
		}

		{

			var studyMaterialUserSubscription mongo_modals.StudyMaterialUserUserSubscriptionModal
			err = database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.FindOne(context.Background(), bson.M{
				"user_id":           user_id,
				"study_material_id": study_material_id,
			}).Decode(&studyMaterialUserSubscription)

			if err == nil {
				//? if its a renewal
				//!Warning, enroll days will be not add up to existing subscriptions
				previous_subscription := studyMaterialUserSubscription.Subscriptions
				previous_subscription = append(previous_subscription, mongo_modals.SubsequentStudyMaterialUserSubscriptionStruct{
					SubscribedOn:   _time,
					DurationInDays: int(studyMaterialList.EnrollDays),
				})
				update_status, err := database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.UpdateOne(
					context.Background(),
					bson.M{
						"user_id":           user_id,
						"study_material_id": study_material_id,
					},
					bson.M{
						"$set": bson.M{
							"expired_on":    _time.AddDate(0, 0, int(studyMaterialList.EnrollDays)),
							"is_enabled":    false,
							"subscriptions": previous_subscription,
							"UpdatedAt":     _time,
							"price":         studyMaterialList.Price,
						},
					},
				)
				if err != nil {
					log.WithFields(log.Fields{
						"user_id":           user_id,
						"study_material_id": study_material_id,
						"err":               err,
					}).Errorln("Failed to enroll to existing subscription")
				} else if update_status.ModifiedCount == 0 {
					log.WithFields(log.Fields{
						"user_id":           user_id,
						"study_material_id": study_material_id,
						"match_count":       update_status.MatchedCount,
						"modified_count":    update_status.ModifiedCount,
					}).Errorln("Failed to enroll to existing subscription")
				} else {
					user_subscriptions[studyMaterialUserSubscription.ID] = studyMaterialList.Price
				}
			} else {
				user_sub_tbl, ins_err := database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.InsertOne(context.Background(), mongo_modals.StudyMaterialUserUserSubscriptionModal{
					UserID:                  user_id,
					StudyMaterialID:         study_material_id,
					InitialSubscriptionDate: _time,
					IsEnabled:               false,
					ExpireOn:                _time.AddDate(0, 0, int(studyMaterialList.EnrollDays)),
					Price:                   studyMaterialList.Price,
					Subscriptions: []mongo_modals.SubsequentStudyMaterialUserSubscriptionStruct{
						{
							SubscribedOn:   _time,
							DurationInDays: int(studyMaterialList.EnrollDays),
							AmountPaid:     studyMaterialList.Price,
						},
					},
					CreatedAt: _time,
					UpdatedAt: _time,
				})
				if ins_err != nil {
					log.WithFields(log.Fields{
						"ins_err": ins_err,
					}).Errorln("Error in inserting data to mongo users")
					// my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Regestering new user while marking active", nil)
					continue
				}
				studyMaterialUserSubscription.ID = user_sub_tbl.InsertedID.(primitive.ObjectID)
				user_subscriptions[studyMaterialUserSubscription.ID] = studyMaterialList.Price
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
// @Summary Confirm payment on Order for study material
// @Schemes
// @Description allow confirm payment on order created & payment status will be verified on server side
// @Tags Customer side(Study materials)
// @Produce json
// @Param order_id query string true "Order ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/confirm_payment_for_study_material_subscription/ [get]
func PaymentConfirmationForSubscriptionForStudyMaterials(c *gin.Context) {
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
		update_status, err := database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.UpdateMany(
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

type GetDocumentKeyReqStruct struct {
	AppID string `json:"app_id" binding:"required"`
}

// @BasePath /api/
// @Summary get document decode key
// @Schemes
// @Description api to get document decode key
// @Tags Customer side(Study materials)
// @Accept json
// @Produce json
// @Param additionalInfo body GetDocumentKeyReqStruct true "Additional info"
// @Param doc_id query string true "Document ID"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /user/get_doc_key/ [post]
func GetDocumentKey(c *gin.Context) {
	ctx := c.Request.Context()
	doc_id_str := c.Query("doc_id")
	var docInfo GetDocumentKeyReqStruct
	if err := c.ShouldBind(&docInfo); err != nil || doc_id_str == "" {
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

	docID, err := primitive.ObjectIDFromHex(doc_id_str)
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid key", nil)
		return
	}

	var appBuildInfo mongo_modals.AppBuildRegistrationModal
	err = database_connections.MONGO_COLLECTIONS.AppBuildRegistration.FindOne(ctx, bson.M{
		"app_id": docInfo.AppID,
	}).Decode(&appBuildInfo)
	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Error in finding app id", nil)
	}

	var docData mongo_modals.StudyMaterialsModal
	err = database_connections.MONGO_COLLECTIONS.StudyMaterial.FindOne(ctx, bson.M{
		"_id": docID,
	}).Decode(&docData)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.WithFields(log.Fields{
				"Error": err,
				"Email": docData.ID,
			}).Warning("Error in finding user email")
			my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "No Doc matched to specified key", nil)
			return
		}
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in finding video", nil)
		return
	}

	var docsUserSubscriptions mongo_modals.StudyMaterialUserUserSubscriptionModal
	err = database_connections.MONGO_COLLECTIONS.StudyMaterialUserSubscription.FindOne(ctx, bson.M{
		"user_id":           user_id,
		"study_material_id": docID,
		"expired_on":        bson.M{"$gt": time.Now()},
	}).Decode(&docsUserSubscriptions)
	if err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Error in finding package in user subscription ", nil)
		return
	}

	key, block_size := my_modules.EncryptWithPKCS(appBuildInfo.AppSecret, docData.FileDecryptionKey)
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "found", map[string]interface{}{
		"key":        key,
		"block_size": block_size,
	})
}
