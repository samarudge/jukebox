package helpers

import(
  "github.com/gin-gonic/gin"
  "github.com/hoisie/mustache"
  "path"
  log "github.com/Sirupsen/logrus"
  "github.com/samarudge/jukebox/views"
)

func Render(c *gin.Context, name string, obj map[string]interface{}){
  filePath := path.Join("views", name)
  templateData, err := views.Asset(filePath)

  if err != nil {
    log.WithFields(log.Fields{
      "path": filePath,
    }).Error("Could not find template file")
    c.String(500, "Template not found")
  } else {
    // Add the context keys to template output
    for k, v := range c.Keys {
      obj[k] = v
    }

    mainTemplate, _ := views.Asset("views/main.html")
    html := mustache.RenderInLayout(string(templateData), string(mainTemplate), obj)

    if c.Writer.Status() == 200{
      c.Status(200)
    }
    c.Writer.Write([]byte(html))
  }
}
