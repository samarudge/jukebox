package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
)

func Index(c *gin.Context){
  helpers.Render(c, "index.html", gin.H{})
}
