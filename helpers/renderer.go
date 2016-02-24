package helpers

import(
  "github.com/gin-gonic/gin"
  "github.com/hoisie/mustache"
  "path"
  "os"
  log "github.com/Sirupsen/logrus"
)

func Render(c *gin.Context, name string, obj map[string]interface{}){
  filePath := path.Join("views", name)

  if _, err := os.Stat(filePath); os.IsNotExist(err) {
    c.Status(500)
    log.WithFields(log.Fields{
      "path": filePath,
    }).Error("Could not find template file")
  } else {
    // Add the context keys to template output
    for k, v := range c.Keys {
      obj[k] = v
    }

    html := mustache.RenderFileInLayout(filePath, "src/jukebox/views/main.html", obj)

    if c.Writer.Status() == 200{
      c.Status(200)
    }
    c.Writer.Write([]byte(html))
  }
}
