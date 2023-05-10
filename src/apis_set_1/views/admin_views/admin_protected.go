package admin_views

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"
	"travel_planner/src/my_modules"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type VideoUploadStruct struct {
	File *multipart.FileHeader `form:"file" binding:"required"`
}

// @BasePath /api/
// @Summary video upload
// @Schemes
// @Description allow people to update their user profile data
// @Tags Video upload
// @Accept mpfd
// @Produce json
// @Param file formData file true "File"
// @Success 200 {object} my_modules.ResponseFormat
// @Failure 400 {object} my_modules.ResponseFormat
// @Failure 403 {object} my_modules.ResponseFormat
// @Failure 500 {object} my_modules.ResponseFormat
// @Router /admin/upload_streaming_video/ [post]
func UploadVideo(c *gin.Context) {
	var form VideoUploadStruct
	video_title := "tutorial_name.mp4"
	base_path := "uploads/public/video"
	upload_path := fmt.Sprintf("%s/original_video", base_path)
	if err := os.MkdirAll(upload_path, 0755); err != nil {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Error creating directory", nil)
		return
	}
	if err := c.ShouldBind(&form); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Invalid payload", nil)
		return
	}
	// file_part := strings.Split(form.File.Filename, ".")
	file_part := strings.Split(video_title, ".")
	if len(file_part) <= 1 {
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "File name doesn't contain file extension", nil)
		return
	}
	file_name, file_ext := strings.Join(file_part[:len(file_part)-1], "."), file_part[len(file_part)-1]
	dst_file_path := fmt.Sprintf("%s/%s_%d.%s", upload_path, file_name, time.Now().UnixMilli(), file_ext)
	// dst_file_path := fmt.Sprintf("./uploads/original_video/%s", video_title)
	if err := c.SaveUploadedFile(form.File, dst_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Failed to upload file", nil)
		return
	}
	if err := my_modules.UploadVideoForStream(c, base_path, file_name, dst_file_path); err != nil {
		log.Errorln(err)
		my_modules.CreateAndSendResponse(c, http.StatusBadRequest, "error", "Failed to create  multi bit rate file chunks", nil)
		return
	}
	my_modules.CreateAndSendResponse(c, http.StatusOK, "success", "", nil)
}
