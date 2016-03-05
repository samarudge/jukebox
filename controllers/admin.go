package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
)

func AdminIndex(c *gin.Context){
  helpers.Render(c, "admin/index.html", gin.H{})
}
