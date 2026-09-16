package servers

import (
	"errors"

	"payment-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

func handle(fn func(*gin.Context) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := fn(c); err != nil {
			var appErr *utils.AppError
			if errors.As(err, &appErr) {
				c.JSON(appErr.Status, gin.H{
					"error": gin.H{"code": appErr.Code, "message": appErr.Message},
				})
				return
			}
			c.JSON(500, gin.H{
				"error": gin.H{"code": "INTERNAL_ERROR", "message": "an unexpected error occurred"},
			})
		}
	}
}
