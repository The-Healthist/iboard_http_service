package http_base_controller

import (
	"net/http"

	base_services "github.com/The-Healthist/iboard_http_service/services/base"
	"github.com/gin-gonic/gin"
)

type UploadController struct {
	ctx     *gin.Context
	service base_services.IUploadService
}

func NewUploadController(
	ctx *gin.Context,
	service base_services.IUploadService,
) *UploadController {
	return &UploadController{
		ctx:     ctx,
		service: service,
	}
}

func (c *UploadController) GetUploadParams() {
	var req struct {
		UploadDir   string `json:"upload_dir" binding:"required"`
		CallbackURL string `json:"callback_url" binding:"required"`
	}

	// Try JSON binding first
	if err := c.ctx.ShouldBindJSON(&req); err != nil {
		// If JSON fails, try form data
		req.UploadDir = c.ctx.PostForm("upload_dir")
		req.CallbackURL = c.ctx.PostForm("callback_url")
		if req.UploadDir == "" || req.CallbackURL == "" {
			c.ctx.JSON(http.StatusBadRequest, gin.H{
				"error": "Missing required parameters",
			})
			return
		}
	}

	policy, err := c.service.GetUploadParams(req.UploadDir, req.CallbackURL)
	if err != nil {
		c.ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.ctx.JSON(http.StatusOK, policy)
}
