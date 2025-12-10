package middleware

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type responseBodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r responseBodyLogWriter) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Read request body
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Capture response body
		w := &responseBodyLogWriter{
			ResponseWriter: c.Writer,
			body:          bytes.NewBufferString(""),
		}
		c.Writer = w

		c.Next()

		duration := time.Since(start)
		statusCode := c.Writer.Status()

		entry := logrus.WithFields(logrus.Fields{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"status":      statusCode,
			"duration":    duration,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"request_id":  c.GetString("request_id"),
			"request_size": len(requestBody),
			"response_size": w.body.Len(),
		})

		if len(requestBody) > 0 && len(requestBody) < 1024 {
			entry = entry.WithField("request_body", string(requestBody))
		}

		if statusCode >= 400 {
			entry.Error("HTTP Request completed with error")
			if w.body.Len() > 0 && w.body.Len() < 1024 {
				entry = entry.WithField("response_body", w.body.String())
			}
		} else {
			entry.Info("HTTP Request completed")
		}
	}
}

func RequestID() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = fmt.Sprintf("%d", time.Now().UnixNano())
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	})
}