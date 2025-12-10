package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.WithFields(logrus.Fields{
					"error":        err,
					"stack":        string(debug.Stack()),
					"method":       c.Request.Method,
					"path":         c.Request.URL.Path,
					"client_ip":    c.ClientIP(),
					"user_agent":   c.Request.UserAgent(),
					"request_id":   c.GetString("request_id"),
				}).Error("Panic recovered")

				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Internal Server Error",
					"request_id": c.GetString("request_id"),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}