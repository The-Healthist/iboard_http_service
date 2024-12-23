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

	// 如果没有提供 callback_url，使用默认的 frp 地址
	if req.CallbackURL == "" {
		req.CallbackURL = "http://your_domain.com/api/upload/callback"
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

func (c *UploadController) UploadCallback() {
	// OSS 回调时会发送文件信息
	var callbackData struct {
		Filename string `form:"filename"`
		Size     int64  `form:"size"`
		MimeType string `form:"mimeType"`
		Height   int    `form:"height"`
		Width    int    `form:"width"`
	}

	if err := c.ctx.ShouldBind(&callbackData); err != nil {
		c.ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid callback data",
		})
		return
	}

	if err := c.service.UploadCallback(); err != nil {
		c.ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.ctx.JSON(http.StatusOK, gin.H{
		"status": "success",
		"data":   callbackData,
	})
}
