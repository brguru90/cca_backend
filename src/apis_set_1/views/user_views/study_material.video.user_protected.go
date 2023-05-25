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
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving video playlist data", nil)
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
