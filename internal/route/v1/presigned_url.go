package v1

import (
	"net/http"

	"grocerics-backend/internal/auth"
	"grocerics-backend/internal/middleware"
	"grocerics-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

func RegisterPresignedURLRoutes(r *gin.Engine, jwt *auth.JWTService, users *repository.UserRepository) {
	group := r.Group("/v1/presigned-url")
	group.Use(middleware.AuthMiddleware(jwt, users))
	group.POST("", getPresignedURL)
}

type PresignRequest struct {
	FileName    string `json:"fileName" binding:"required"`
	Folder      string `json:"folder" binding:"required"`
	ContentType string `json:"contentType" binding:"required"`
}

// @Swagger:route POST /v1/presigned-url upload getPresignedURL
// @Summary      Get presigned URL for S3 upload
// @Description  Generate a presigned PUT URL so client can upload directly to S3
// @Tags         File Upload
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body PresignRequest true "Presign Request"
// @Success      200 {object} dto.Response{data=dto.PresignedURL}
// @Failure      400 {object} dto.Response{data=string}
// @Failure      401 {object} dto.Response{data=string}
// @Failure      403 {object} dto.Response{data=string}
// @Router       /v1/presigned-url [post]
func getPresignedURL(c *gin.Context) {
	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// result, err := storage.GeneratePresignedPutURL(c.Request.Context(), h.S3Client, h.Bucket, req.Folder, req.FileName, req.ContentType)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to generate presigned url"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"success":      true,
	// 	"presignedUrl": result.PresignedURL,
	// 	"publicUrl":    result.PublicURL,
	// })
}
