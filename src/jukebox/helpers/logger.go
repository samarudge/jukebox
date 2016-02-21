package helpers

import(
  log "github.com/Sirupsen/logrus"
  "github.com/gin-gonic/gin"
  "time"
)

func Logger() gin.HandlerFunc{
  /*
    Do logging with logrus instead of GINs default Logger
  */

  return func(c *gin.Context) {
    t := time.Now()
    c.Next()
    latency := time.Since(t)
    status := c.Writer.Status()
    clientIP := c.ClientIP()
    method := c.Request.Method

    log.WithFields(log.Fields{
      "Duration": latency,
      "Status": status,
      "ClientIP": clientIP,
      "Method": method,
      "Path": c.Request.URL.Path,
    }).Info("request")
  }
}
