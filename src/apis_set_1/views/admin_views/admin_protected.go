package admin_views

import (
	"cca/src/my_modules"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type VideoUploadStruct struct {
	VideoFile        *multipart.FileHeader `form:"video_file" binding:"required"`
	PreviewImageFile *multipart.FileHeader `form:"preview_image_file" binding:"required"`
}
type VideoInfoStruct struct {
	Title       string `form:"title" binding:"required"`
	Description string `form:"description"`
	CreatedBy   string `form:"created_by" binding:"required"`
}

// @BasePath /api/
// @Summary video upload
// @Schemes
// @Description api to upload video content for multiple adaptive bit rate streaming
// @Tags Video upload
// @Accept mpfd
// @Produce json
// @Param video_file formData file true "Video file"
// @Param preview_image_file formData file true "Preview image file"
// @Param form_data formData VideoInfoStruct true "Video upload"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/upload_streaming_video/ [post]
func UploadVideo(c *gin.Context) {
	var uploadForm VideoUploadStruct
	var infoForm VideoInfoStruct
	base_path := "uploads"
	protected_video := fmt.Sprintf("%s/private/video", base_path)
	unprotected_video := fmt.Sprintf("%s/public/video", base_path)
	upload_path := fmt.Sprintf("%s/original_video", protected_video)
	if err := os.MkdirAll(upload_path, 0755); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error creating directory", nil)
		return
	}
	if err := c.ShouldBind(&uploadForm); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid upload payload", nil)
		return
	}
	if err := c.ShouldBind(&infoForm); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid data payload", nil)
		return
	}
	// file_part := strings.Split(form.File.Filename, ".")
	file_part := strings.Split(uploadForm.VideoFile.Filename, ".")
	if len(file_part) <= 1 {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "File name doesn't contain file extension", nil)
		return
	}
	file_name, file_ext := strings.Join(file_part[:len(file_part)-1], "."), file_part[len(file_part)-1]
	dst_file_path := fmt.Sprintf("%s/%s_%d.%s", upload_path, file_name, time.Now().UnixMilli(), file_ext)
	// dst_file_path := fmt.Sprintf("./uploads/original_video/%s", video_title)
	if err := c.SaveUploadedFile(uploadForm.VideoFile, dst_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Failed to upload file", nil)
		return
	}
	if err := my_modules.UploadVideoForStream(c, unprotected_video, file_name, dst_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Failed to create  multi bit rate file chunks", nil)
		return
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "", nil)
}
