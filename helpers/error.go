package helpers

import(
  "github.com/gin-gonic/gin"
  log "github.com/Sirupsen/logrus"
)

func Send500(c *gin.Context, err string){
  c.Status(500)
  Render(c, "error.html", gin.H{
    "errorTitle": "Application Error",
    "errorDetails": err,
  })
  log.WithFields(log.Fields{
    "error": err,
  }).Error("Server error")
  c.Abort()
}

func Send403(c *gin.Context, err string){
  c.Status(403)
  Render(c, "error.html", gin.H{
    "errorTitle": "Authorization Error",
    "errorDetails": err,
    "showLogin": true,
  })
  c.Abort()
}

func Send404(c *gin.Context, err string){
  c.Status(404)
  Render(c, "error.html", gin.H{
    "errorTitle": "Not Found",
    "errorDetails": err,
  })
  c.Abort()
}
