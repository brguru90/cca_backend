package admin_views

import (
	"cca/src/configs"
	"cca/src/database/database_connections"
	"cca/src/database/database_utils"
	"cca/src/database/mongo_modals"
	"cca/src/my_modules"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
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

type DocUploadReqStruct struct {
	DocFile          *multipart.FileHeader `form:"doc_file" binding:"required"`
	PreviewImageFile *multipart.FileHeader `form:"preview_image_file" binding:"required"`
}
type DocInfoReqStruct struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	Category    string `form:"category"`
	Author      string `form:"author" binding:"required"`
	IsLive      bool   `form:"is_live"`
	Price       int64  `form:"price"`
	EnrollDays  int16  `form:"enroll_days"`
}

// @BasePath /api/
// @Summary Upload study material
// @Schemes
// @Description api Upload study material
// @Tags Manage Study material
// @Accept mpfd
// @Produce json
// @Param doc_file formData file true "Document file"
// @Param preview_image_file formData file true "Preview image file"
// @Param form_data formData DocInfoReqStruct true "Document upload"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/upload_study_material/ [post]
func UploadStudyMaterials(c *gin.Context) {
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

	var uploadForm DocUploadReqStruct
	var infoForm DocInfoReqStruct
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
	upload_path := fmt.Sprintf("%s/docs", UNPROTECTED_UPLOAD_PATH)
	if err := os.MkdirAll(upload_path, 0755); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error creating directory", nil)
		return
	}
	if err := os.MkdirAll(unprotected_image, 0755); err != nil {
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

	doc_file_name := strings.Split(uploadForm.DocFile.Filename, ".")
	doc_file_ext := doc_file_name[len(doc_file_name)-1]

	non_alphaNumeric := regexp.MustCompile(`[^a-zA-Z0-9]`)
	dst_file_name := non_alphaNumeric.ReplaceAllString(infoForm.Title, "")
	dst_file_path := fmt.Sprintf("%s/%s_%d", upload_path, dst_file_name, time.Now().UnixMilli())
	dst_image_file_path_public := fmt.Sprintf("%s/%s_%d", unprotected_image, dst_file_name, time.Now().UnixMilli())
	dst_doc_file_path := fmt.Sprintf("%s.%s", dst_file_path, doc_file_ext)
	preview_image_file_part := strings.Split(uploadForm.PreviewImageFile.Filename, ".")
	if len(preview_image_file_part) <= 1 {
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "File name doesn't contain file extension", nil)
		return
	}
	dst_preview_image_file_path := fmt.Sprintf("%s.%s", dst_image_file_path_public, preview_image_file_part[len(preview_image_file_part)-1])

	file, _, err := c.Request.FormFile("doc_file")
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}
	defer file.Close()
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}
	var fileBytes []byte
	fileBytes, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}
	var random_string string
	if _rand, r_err := my_modules.RandomBytes(16); r_err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	} else {
		random_string = hex.EncodeToString(_rand)[0:16]
	}
	encrypted, blk_size := my_modules.EncryptWithPKCS(random_string, base64.URLEncoding.EncodeToString(fileBytes))
	err = ioutil.WriteFile(dst_doc_file_path, []byte(encrypted), 0755)
	if err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}
	// if err := c.SaveUploadedFile(uploadForm.DocFile, dst_doc_file_path); err != nil {
	// 	log.Errorln(err)
	// 	my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
	// 	return
	// }

	if err := c.SaveUploadedFile(uploadForm.PreviewImageFile, dst_preview_image_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Failed to upload file", nil)
		return
	}

	_time := time.Now()
	_, ins_err := database_connections.MONGO_COLLECTIONS.StudyMaterial.InsertOne(ctx, mongo_modals.StudyMaterialsModal{
		Title:                    infoForm.Title,
		CreatedByUser:            infoForm.Author,
		Description:              infoForm.Description,
		Category:                 infoForm.Category,
		PathToBookCoverImage:     dst_preview_image_file_path,
		PathToDocFile:            dst_doc_file_path,
		LinkToBookCoverImage:     strings.Replace(dst_preview_image_file_path, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
		LinkToDocFile:            strings.Replace(dst_doc_file_path, UNPROTECTED_UPLOAD_PATH, CDN_PATH, 1),
		IsLive:                   infoForm.IsLive,
		Price:                    infoForm.Price,
		UploadedByUser:           _id,
		FileDecryptionKey:        random_string,
		EnrollDays:               infoForm.EnrollDays,
		FileDecryptionKeyBlkSize: blk_size,
		CreatedAt:                _time,
		UpdatedAt:                _time,
	})
	if ins_err != nil {
		os.Remove(dst_doc_file_path)
		os.Remove(dst_preview_image_file_path)
		resp_err, is_known := database_utils.GetDBErrorString(ins_err)
		if !is_known {
			log.WithFields(log.Fields{
				"ins_err": ins_err,
			}).Errorln("Error in inserting data to mongo users")
		}
		my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", resp_err, nil)
		return
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "success", nil)
}

type GetAllUploadedStudyMaterialsRespStruct struct {
	my_modules.ResponseFormat
	Data []mongo_modals.StudyMaterialsModal `json:"data"`
}

// @BasePath /api/
// @Summary get all uploaded documents
// @Schemes
// @Description api to get all uploaded documents
// @Tags Manage Study material
// @Produce json
// @Success 200 {object} GetAllUploadedStudyMaterialsRespStruct
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/doc_upload_list/ [get]
func GetAllUploadedStudyMaterials(c *gin.Context) {
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
	cursor, err = database_connections.MONGO_COLLECTIONS.StudyMaterial.Find(ctx, bson.M{
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
		var docsList []mongo_modals.StudyMaterialsModal = []mongo_modals.StudyMaterialsModal{}
		for cursor.Next(c.Request.Context()) {
			var docData mongo_modals.StudyMaterialsModal
			if err = cursor.Decode(&docData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				// continue
				my_modules.CreateAndSendResponse(c, http.StatusInternalServerError, "error", "Error in retrieving user data", nil)
				return
			}
			docsList = append(docsList, docData)
		}
		my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "Record found", docsList)
	}
}

type StudyMaterialIDs struct {
	Id []string `json:"docs_ids" binding:"required"`
}

// @BasePath /api/
// @Summary Delete study material
// @Schemes
// @Description api delete a document from study material
// @Tags Manage Study material
// @Accept json
// @Produce json
// @Param docs_ids body StudyMaterialIDs true "Remove documents"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/delete_study_material/ [delete]
func RemoveStudyMaterial(c *gin.Context) {
	ctx := c.Request.Context()
	var docs_data StudyMaterialIDs
	if err := c.ShouldBind(&docs_data); err != nil {
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

	var docs_ids []primitive.ObjectID
	for i := 0; i < len(docs_data.Id); i++ {
		objID, err := primitive.ObjectIDFromHex(docs_data.Id[i])
		if err == nil {
			docs_ids = append(docs_ids, objID)
		}
	}

	where := bson.M{"_id": bson.M{"$in": docs_ids}, "uploaded_by_user": user_id}

	if payload.Data.AccessLevel == "super_admin" {
		where = bson.M{"_id": bson.M{"$in": docs_ids}}
	}

	var response_data = make(map[string]interface{})
	var files_deleted = make(map[string]interface{})
	var err error
	var cursor *mongo.Cursor
	cursor, err = database_connections.MONGO_COLLECTIONS.StudyMaterial.Find(ctx, where)
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
			var docsData mongo_modals.StudyMaterialsModal
			if err = cursor.Decode(&docsData); err != nil {
				log.Errorln(fmt.Sprintf("Scan failed: %v\n", err))
				continue
			}
			files_deleted[docsData.Title] = map[string]interface{}{
				"doc":         false,
				"cover_image": false,
			}
			err1 := os.Remove(docsData.PathToDocFile)
			err2 := os.Remove(docsData.PathToBookCoverImage)
			files_deleted[docsData.Title] = map[string]interface{}{
				"doc":         err1 == nil,
				"cover_image": err2 == nil,
			}
		}
	}

	result, err := database_connections.MONGO_COLLECTIONS.StudyMaterial.DeleteMany(context.Background(), where)
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
