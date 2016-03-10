package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
)

func AdminIndex(c *gin.Context){
  systemSpotify := models.Spotify{}
  systemSpotify.LoadSystem()
  helpers.Render(c, "admin/index.html", gin.H{
    "systemSpotify": systemSpotify,
  })
}
