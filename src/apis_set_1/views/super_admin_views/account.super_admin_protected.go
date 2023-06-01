package super_admin_views

import (
	"cca/src/database/database_connections"
	"cca/src/database/database_utils"
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
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type GetUsersRespStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.UsersModel `json:"data"`
}

// @BasePath /api/
// @Summary Get list of user
// @Schemes
// @Description api to fetch list of users
// @Tags Super Admin (Account)
// @Produce json
// @Success 200 {object} GetUsersRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /super_admin/user [get]
func GetUsers(c *gin.Context) {
	ctx := c.Request.Context()

	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.Users.Find(ctx,
		bson.M{
			"$or": []interface{}{
				bson.M{"access_level": "super_admin"},
				bson.M{"access_level": "admin"},
			},
		},
	)
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
			userData.Password = ""
			usersData = append(usersData, userData)
		}

		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", map[string]interface{}{
			"users": usersData,
		})
	}

}

type CredentialErrorPayload struct {
	Errors_data map[string]interface{} `json:"errors,omitempty"`
}

type AddAdminUsersReqStruct struct {
	Email        string `json:"email" binding:"required"`
	Password     string `json:"password" binding:"required"`
	Username     string `json:"username" binding:"required"`
	IsSuperAdmin bool   `json:"is_super_admin"`
}

type AddAdminUsersRespStruct struct {
	my_modules.ResponseFormat
	Data mongo_modals.UsersModel `json:"data"`
}

// @BasePath /api
// @Summary url to add user
// @Schemes
// @Description allow admin to add new admin/super_admin users
// @Tags Super Admin (Account)
// @Accept json
// @Produce json
// @Param new_user body AddAdminUsersReqStruct true "Add user"
// @Success 200 {object} AddAdminUsersRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /super_admin/user  [post]
func AddAdminUsers(c *gin.Context) {
	ctx := c.Request.Context()

	var newUserRow AddAdminUsersReqStruct
	if err := c.ShouldBindJSON(&newUserRow); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}

	_errors := make(map[string]interface{})
	var newUserData mongo_modals.UsersModel
	{
		_time := time.Now()
		ph := sha1.Sum([]byte(newUserRow.Password))
		newUserData = mongo_modals.UsersModel{
			Email:             newUserRow.Email,
			Username:          newUserRow.Username,
			Password:          hex.EncodeToString(ph[:]),
			AuthProvider:      "email",
			CreatedAt:         _time,
			UpdatedAt:         _time,
			AccessLevel:       my_modules.AccessLevel.ADMIN.Label,
			AccessLevelWeight: my_modules.AccessLevel.ADMIN.Weight,
		}

		if newUserRow.IsSuperAdmin {
			newUserData.AccessLevel = my_modules.AccessLevel.SUPER_ADMIN.Label
			newUserData.AccessLevelWeight = my_modules.AccessLevel.SUPER_ADMIN.Weight
		}

		ins_res, ins_err := database_connections.MONGO_COLLECTIONS.Users.InsertOne(ctx, newUserData)
		if ins_err != nil {
			resp_err, is_known := database_utils.GetDBErrorString(ins_err)
			if is_known {
				_errors["email"] = resp_err
			} else {
				log.WithFields(log.Fields{
					"ins_err": ins_err,
				}).Errorln("Error in inserting data to mongo users")
			}
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Registering new user", CredentialErrorPayload{Errors_data: _errors})
			return
		} else {
			newUserData.Password = ""
			newUserData.ID = ins_res.InsertedID.(primitive.ObjectID)
			my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "User added success fully", newUserData)
			return
		}
	}
}

// @BasePath /api
// @Summary url to delete user
// @Schemes
// @Description allow admin to delete admin user
// @Tags Super Admin (Account)
// @Accept json
// @Produce json
// @Param user_id query string true "User ID"
// @Success 200 {object} AddAdminUsersRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /super_admin/user  [delete]
func RemoveAdminUsers(c *gin.Context) {

	user_id_str := c.Query("user_id")
	if user_id_str == "" {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}

	payload, ok := my_modules.ExtractTokenPayload(c)
	if !ok {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Unable to get user info", nil)
		return
	}

	if payload.Data.ID == user_id_str {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Not allowed to deleted current user", nil)
		return
	}

	user_id, _id_err := primitive.ObjectIDFromHex(user_id_str)

	if user_id_str == "" || _id_err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "incorrect user id", _id_err)
		return
	}

	result, err := database_connections.MONGO_COLLECTIONS.Users.DeleteOne(context.Background(), bson.M{
		"_id": user_id,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("Failed to delete user data")
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to delete user data", nil)
		return
	}
	if result.DeletedCount > 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "User entry removed", nil)
	} else {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "error", "Unable to remove user", nil)
	}
}

type UpdateAdminUsersCredentialsReqStruct struct {
	ID           primitive.ObjectID `json:"user_id" binding:"required"`
	Password     string             `json:"password"`
	IsSuperAdmin bool               `json:"is_super_admin"`
}

// @BasePath /api
// @Summary url to update user data
// @Schemes
// @Description allow admin to update admin/super_admin users credentials
// @Tags Super Admin (Account)
// @Accept json
// @Produce json
// @Param update_data body UpdateAdminUsersCredentialsReqStruct true "Data to update"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /super_admin/user  [put]
func UpdateAdminUsersCredentials(c *gin.Context) {
	ctx := c.Request.Context()

	var updateUserData UpdateAdminUsersCredentialsReqStruct
	if err := c.ShouldBindJSON(&updateUserData); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}

	access_level := my_modules.AccessLevel.ADMIN.Label
	access_level_weight := my_modules.AccessLevel.ADMIN.Weight

	if updateUserData.IsSuperAdmin {
		access_level = my_modules.AccessLevel.SUPER_ADMIN.Label
		access_level_weight = my_modules.AccessLevel.SUPER_ADMIN.Weight
	}

	update_with_data := bson.M{
		"access_level":        access_level,
		"access_level_weight": access_level_weight,
		"updatedAt":           time.Now(),
	}

	if strings.TrimSpace(updateUserData.Password) != "" {
		ph := sha1.Sum([]byte(updateUserData.Password))
		update_with_data = bson.M{
			"access_level":        access_level,
			"access_level_weight": access_level_weight,
			"password":            hex.EncodeToString(ph[:]),
			"updatedAt":           time.Now(),
		}
	}

	res, err := database_connections.MONGO_COLLECTIONS.Users.UpdateOne(
		ctx,
		bson.M{
			"_id": updateUserData.ID,
		},
		bson.M{
			"$set": update_with_data,
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

	if res.ModifiedCount > 0 {
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Updated", nil)
	} else {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to update data", nil)
	}

}
