package user_views

import (
	"cca/src/database/database_connections"
	"cca/src/database/database_utils"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	log "github.com/sirupsen/logrus"
)

func GetAllUserData(c *gin.Context) {
	ctx := c.Request.Context()

	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.Users.Find(ctx, bson.M{})
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
		var usersData []mongo_modals.UsersModel = []mongo_modals.UsersModel{}
		for cursor.Next(c.Request.Context()) {
			var userData mongo_modals.UsersModel
			if err = cursor.Decode(&userData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				// continue
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retriving user data", nil)
				return
			}
			usersData = append(usersData, userData)
		}

		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", map[string]interface{}{
			"users": usersData,
		})
	}
}

type RegisterBuildReqStruct struct {
	AppID      string `form:"app_id" binding:"required"`
	AppSecret  string `form:"app_secret" binding:"required"`
	APIAuthKey string `form:"auth_key" binding:"required"`
}

// @BasePath /api
// @Summary url to Register app
// @Schemes
// @Description Register app
// @Tags App Build
// @Accept mpfd
// @Produce json
// @Param app_info formData RegisterBuildReqStruct true "Application information"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /register_build [post]
func RegisterBuild(c *gin.Context) {
	ctx := c.Request.Context()
	var buildRegistrationInfo RegisterBuildReqStruct
	if err := c.ShouldBind(&buildRegistrationInfo); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Invalid upload payload", nil)
		return
	}

	if buildRegistrationInfo.APIAuthKey != "1234" {
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Access denied", nil)
		return
	}

	_time := time.Now()
	_, ins_err := database_connections.MONGO_COLLECTIONS.AppBuildRegistration.InsertOne(ctx, mongo_modals.AppBuildRegistrationModal{
		AppID:     buildRegistrationInfo.AppID,
		AppSecret: buildRegistrationInfo.AppSecret,
		CreatedAt: _time,
		UpdatedAt: _time,
	})
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
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", nil)
}
