package user_views

import (
	"cca/src/database/database_connections"
	"cca/src/database/database_utils"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	log "github.com/sirupsen/logrus"
)

type CredentialErrorPayload struct {
	Errors_data map[string]interface{} `json:"errors,omitempty"`
}

type UserCredentialReqStruct struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Username string `json:"username"`
}

// @BasePath /api
// @Summary url to signup
// @Schemes
// @Description allow people to create new to user account
// @Tags Account
// @Accept json
// @Produce json
// @Param new_user body UserCredentialReqStruct true "Add user"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /sign_up [post]
func SignUp(c *gin.Context) {
	ctx := context.Background()
	var newUserRow UserCredentialReqStruct
	// ShouldBindJSON will validate json body & convert it to structure object
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
			// Uid:          uuid.New().String(),
			Email:             newUserRow.Email,
			Username:          newUserRow.Username,
			Password:          hex.EncodeToString(ph[:]),
			AuthProvider:      "email",
			CreatedAt:         _time,
			UpdatedAt:         _time,
			AccessLevel:       my_modules.AccessLevel.CUSTOMER.Label,
			AccessLevelWeight: my_modules.AccessLevel.CUSTOMER.Weight,
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
			newUserData.ID = ins_res.InsertedID.(primitive.ObjectID)
			// log.WithFields(log.Fields{
			// 	"ins_res": ins_res.InsertedID,
			// 	"type":    fmt.Sprintf("%T", ins_res.InsertedID),
			// }).Info("successfully inserted data into mongo users")
		}
	}
	newUserData.Password = ""
	access_token_payload := my_modules.Authenticate(c, newUserData, "some safe data", false)
	{
		_time := time.Now()
		_, ins_err := database_connections.MONGO_COLLECTIONS.ActiveSessions.InsertOne(ctx, mongo_modals.ActiveSessionsModel{
			UserID:    newUserData.ID,
			TokenID:   access_token_payload.Token_id,
			IP:        c.ClientIP(),
			UA:        c.GetHeader("User-Agent"),
			Exp:       access_token_payload.Exp,
			Status:    "active",
			CreatedAt: _time,
			UpdatedAt: _time,
		})
		if ins_err != nil {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Registering new user while marking active", nil)
			return
		}
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Registered successfully", newUserData)
}

// @BasePath /api
// @Summary url to login
// @Schemes
// @Description allow people to login into their account
// @Tags Account
// @Accept json
// @Produce json
// @Param existing_user body UserCredentialReqStruct true "Add user"
// @Param access_level query string false "Access level" Enums(admin, super_admin, customer)
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /login [post]
func Login(c *gin.Context) {
	ctx := context.Background()
	_errors := make(map[string]interface{})
	var userCredential UserCredentialReqStruct
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}
	var userData mongo_modals.UsersModel
	{
		where := bson.M{
			"email": userCredential.Email,
		}
		access_level, ok := c.GetQuery("access_level")
		if ok {
			_, ok = my_modules.AllAccessLevel[access_level]
			if !ok {
				access_level = my_modules.AccessLevel.CUSTOMER.Label
			}
			where = bson.M{
				"email":        userCredential.Email,
				"access_level": access_level,
			}
		}

		err := database_connections.MONGO_COLLECTIONS.Users.FindOne(ctx, where).Decode(&userData)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.WithFields(log.Fields{
					"Error": err,
					"Email": userData.Email,
				}).Warning("Error in finding user email")
				_errors["email"] = "email doesn't exists"
				my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "email doesn't exists", CredentialErrorPayload{Errors_data: _errors})
				return
			}
			log.WithFields(log.Fields{
				"Error": err,
				"Email": userData.Email,
			}).Error("Error in finding user email")
			_errors["email"] = "Error in finding user"
			my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error in finding user", CredentialErrorPayload{Errors_data: _errors})
			return
		}

		ph := sha1.Sum([]byte(userCredential.Password))
		if userData.Password != hex.EncodeToString(ph[:]) {
			_errors["password"] = "Wrong password"
			my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Invalid credential", CredentialErrorPayload{Errors_data: _errors})
			return
		}
	}
	userData.Password = ""
	userData.AuthProvider = "email"
	access_token_payload := my_modules.Authenticate(c, userData, "some safe data", false)
	{
		_time := time.Now()
		_, ins_err := database_connections.MONGO_COLLECTIONS.ActiveSessions.InsertOne(ctx, mongo_modals.ActiveSessionsModel{
			UserID:    userData.ID,
			TokenID:   access_token_payload.Token_id,
			IP:        c.ClientIP(),
			UA:        c.GetHeader("User-Agent"),
			Exp:       access_token_payload.Exp,
			Status:    "active",
			CreatedAt: _time,
			UpdatedAt: _time,
		})
		if ins_err != nil {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Regestering new user while marking active", nil)
			return
		}
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Authorization success", userData)
}

type UserMobileCredentialReqStruct struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// @BasePath /api
// @Summary url to login with mobile number
// @Schemes
// @Description allow people to login into their account
// @Tags Account
// @Accept json
// @Produce json
// @Param existing_user body UserMobileCredentialReqStruct true "Add user"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /login_mobile [post]
func LoginWithMobile(c *gin.Context) {
	ctx := context.Background()
	_errors := make(map[string]interface{})
	var userCredential UserMobileCredentialReqStruct
	if err := c.ShouldBindJSON(&userCredential); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}
	var userData mongo_modals.UsersModel
	{
		err := database_connections.MONGO_COLLECTIONS.Users.FindOne(ctx, bson.M{
			"mobile": userCredential.Mobile,
		}).Decode(&userData)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				log.WithFields(log.Fields{
					"Error":  err,
					"Mobile": userData.Mobile,
				}).Warning("Error in finding user mobile")
				_errors["mobile"] = "mobile doesn't exists"
				my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Mobile number doesn't exists", CredentialErrorPayload{Errors_data: _errors})
				return
			}
			log.WithFields(log.Fields{
				"Error":  err,
				"Mobile": userData.Mobile,
			}).Error("Error in finding user mobile")
			_errors["email"] = "Error in finding user"
			my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error in finding user", CredentialErrorPayload{Errors_data: _errors})
			return
		}
		ph := sha1.Sum([]byte(userCredential.Password))
		phs := hex.EncodeToString(ph[:])
		if userData.Password != phs {
			_errors["password"] = "Wrong password"
			my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Invalid credential", CredentialErrorPayload{Errors_data: _errors})
			return
		}
	}
	userData.Password = ""
	userData.AuthProvider = "phone"
	access_token_payload := my_modules.Authenticate(c, userData, "some safe data", false)
	{
		_time := time.Now()
		_, ins_err := database_connections.MONGO_COLLECTIONS.ActiveSessions.InsertOne(ctx, mongo_modals.ActiveSessionsModel{
			UserID:    userData.ID,
			TokenID:   access_token_payload.Token_id,
			IP:        c.ClientIP(),
			UA:        c.GetHeader("User-Agent"),
			Exp:       access_token_payload.Exp,
			Status:    "active",
			CreatedAt: _time,
			UpdatedAt: _time,
		})
		if ins_err != nil {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Regestering new user while marking active", nil)
			return
		}
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Authorization success", userData)
}

type SocialAuthReqStruct struct {
	IdToken  string `json:"idToken" binding:"required"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type SocialAuthLoginStruct struct {
	LoginType string `json:"login_type" binding:"required"`
	mongo_modals.UsersModel
}

// @BasePath /api
// @Summary Verify Social Authentication
// @Schemes
// @Description  url to signup/login with social authentication
// @Tags Account
// @Accept json
// @Produce json
// @Param new_or_existing_user body SocialAuthReqStruct true "Add user"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Router /verify_social_auth [post]
func VerifySocialAuth(c *gin.Context) {
	var socialAuth SocialAuthReqStruct
	if err := c.ShouldBindJSON(&socialAuth); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid input data format", nil)
		return
	}
	client, err := my_modules.FIREBASE_APP.Auth(c)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("FIREBASE_APP.Auth")
	}
	token, err := client.VerifyIDToken(c, socialAuth.IdToken)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Errorln("FIREBASE_APP VerifyIDToken")
		my_modules.CreateAndSendResponse(c, http.StatusForbidden, "error", "Token verification failed", nil)
	}

	// log.WithFields(log.Fields{
	// 	"token":      token,
	// 	"claims":     token.Claims["name"],
	// 	"Identities": token.Firebase.Identities,
	// 	"Issuer":     token.Firebase.SignInProvider,
	// }).Infoln()
	// token="&{1652984212 https://securetoken.google.com/travel-planner-17377 travel-planner-17377 1652987814 1652984214 0cD0lvwH2aYKEpkx6HDvk2ZR4Af1 0cD0lvwH2aYKEpkx6HDvk2ZR4Af1 {google.com  map[email:[brguru90@gmail.com] google.com:[114141226575791102096]]} map[auth_time:1.652984212e+09 email:brguru90@gmail.com email_verified:true firebase:map[identities:map[email:[brguru90@gmail.com] google.com:[114141226575791102096]] sign_in_provider:google.com] name:GURUPRASAD BR picture:https://lh3.googleusercontent.com/a-/AOh14GjplxOLH3kgFpntqdydk9guYNJQdpDMqvPC7GEf=s96-c user_id:0cD0lvwH2aYKEpkx6HDvk2ZR4Af1]}"

	// claims="map[auth_time:1.65298436e+09 email:brguru90@gmail.com email_verified:true firebase:map[identities:map[email:[brguru90@gmail.com] google.com:[114141226575791102096]] sign_in_provider:google.com] name:GURUPRASAD BR picture:https://lh3.googleusercontent.com/a-/AOh14GjplxOLH3kgFpntqdydk9guYNJQdpDMqvPC7GEf=s96-c user_id:0cD0lvwH2aYKEpkx6HDvk2ZR4Af1]"

	ctx := context.Background()
	var newUserData mongo_modals.UsersModel
	var login bool = true
	{
		_time := time.Now()
		newUserData = mongo_modals.UsersModel{
			Uid:               token.UID,
			AuthProvider:      token.Firebase.SignInProvider,
			AccessLevel:       my_modules.AccessLevel.CUSTOMER.Label,
			AccessLevelWeight: my_modules.AccessLevel.CUSTOMER.Weight,
			CreatedAt:         _time,
			UpdatedAt:         _time,
		}
		if token.Firebase.SignInProvider == "phone" {
			ph := sha1.Sum([]byte(socialAuth.Password))
			newUserData.Mobile = token.Claims["phone_number"].(string)
			newUserData.Username = socialAuth.Name
			newUserData.Password = hex.EncodeToString(ph[:])
		} else {
			newUserData.Username = token.Claims["name"].(string)
		}
		// checking user already exists
		filter := bson.D{{
			"$or",
			bson.A{
				bson.D{
					{"email", newUserData.Email},
				},
				bson.D{
					{"uid", newUserData.Uid},
				},
			},
		}}
		err := database_connections.MONGO_COLLECTIONS.Users.FindOne(ctx, filter).Decode(&newUserData)
		log.WithFields(log.Fields{
			"newUserData": newUserData,
		}).Infoln()

		if err != nil {
			if err == mongo.ErrNoDocuments {
				// if user not already exists
				ins_res, ins_err := database_connections.MONGO_COLLECTIONS.Users.InsertOne(ctx, newUserData)
				if ins_err != nil {
					log.WithFields(log.Fields{
						"ins_err": ins_err,
					}).Errorln("Error in inserting data to mongo users")
					my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Registering new user", nil)
					return
				} else {
					newUserData.ID = ins_res.InsertedID.(primitive.ObjectID)
					login = false
				}
			} else {
				log.WithFields(log.Fields{
					"newUserData": newUserData,
					"Error":       err,
				}).Error("Error in finding user")
				my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error in finding user", nil)
				return
			}
		}

		if login && token.Firebase.SignInProvider == "phone" {
			ph := sha1.Sum([]byte(socialAuth.Password))
			var updateUserData mongo_modals.UsersModel
			// updateUserData := newUserData
			updateUserData.Uid = newUserData.Uid
			updateUserData.AuthProvider = newUserData.AuthProvider
			updateUserData.Mobile = newUserData.Mobile
			updateUserData.Username = socialAuth.Name
			updateUserData.AccessLevel = my_modules.AccessLevel.CUSTOMER.Label
			updateUserData.AccessLevelWeight = my_modules.AccessLevel.CUSTOMER.Weight
			updateUserData.Password = hex.EncodeToString(ph[:])
			if updateWith, bsonParseErr := my_modules.StructToBsonD(updateUserData); bsonParseErr == nil {
				updateRes, update_err := database_connections.MONGO_COLLECTIONS.Users.UpdateOne(ctx,
					filter,
					bson.D{{"$set", updateWith}})
				if update_err != nil {
					log.WithFields(log.Fields{
						"updateUserData": updateUserData,
						"Error":          update_err,
					}).Error("Password update failed for mobile login")
				} else if updateRes.MatchedCount == 0 {
					log.WithFields(log.Fields{
						"updateUserData": updateUserData,
						"Error":          update_err,
					}).Error("Password update failed for mobile login,due to mismatch of credential")
				} else if updateRes.ModifiedCount > 0 {
					newUserData = updateUserData
				}
				if updateRes.MatchedCount != 0 {
					err := database_connections.MONGO_COLLECTIONS.Users.FindOne(ctx, filter).Decode(&newUserData)
					if err != nil {
						newUserData = updateUserData
					}
				}
			} else {
				log.WithFields(log.Fields{
					"updateUserData": updateUserData,
					"Error":          bsonParseErr,
				}).Error("unable to convert struct to bson")
			}

		}

	}

	// access_token_payload := my_modules.Authenticate(c, newUserData)
	loginData := SocialAuthLoginStruct{
		UsersModel: newUserData,
		LoginType:  "login",
	}
	if !login {
		loginData.LoginType = "signup"
	}

	// log.WithFields(log.Fields{
	// 	"loginData":  loginData,
	// 	"UsersModel": loginData.UsersModel,
	// }).Infoln()

	access_token_payload := my_modules.Authenticate(c, loginData.UsersModel, "some safe data", false)
	{
		_time := time.Now()
		_, ins_err := database_connections.MONGO_COLLECTIONS.ActiveSessions.InsertOne(ctx, mongo_modals.ActiveSessionsModel{
			UserID:    newUserData.ID,
			TokenID:   access_token_payload.Token_id,
			IP:        c.ClientIP(),
			UA:        c.GetHeader("User-Agent"),
			Exp:       access_token_payload.Exp,
			Status:    "active",
			CreatedAt: _time,
			UpdatedAt: _time,
		})
		if ins_err != nil {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
			my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in Registering new user while marking active", nil)
			return
		}
	}
	loginData.Password = ""
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Authorization success", loginData)

}

type LoginStatusPayloadReqStruct struct {
	ExtendSession bool `json:"extend_session" binding:"required"`
}

// @BasePath /api
// @Summary Login status
// @Schemes
// @Description api used to validate user login session
// @Tags Account
// @Accept json
// @Produce json
// @Param access_level query string false "Access level" Enums(admin, super_admin, customer)
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /login_status [get]
func LoginStatus(c *gin.Context) {
	var loginStatusPayload LoginStatusPayloadReqStruct

	access_level, ok := my_modules.AllAccessLevelReverseMap[c.Query("access_level")]

	decoded_token, err, http_status, ok := my_modules.LoginStatus(c, access_level, false)
	if err != "" {
		my_modules.CreateAndSendResponse(c, http_status, "error", err, nil)
		return
	}
	if ok {
		if err := c.ShouldBindJSON(&loginStatusPayload); err == nil {
			if loginStatusPayload.ExtendSession {
				// need to test
				decoded_token = my_modules.RenewAuthentication(c, decoded_token)
			}
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "active", decoded_token.Data)
	}
}
