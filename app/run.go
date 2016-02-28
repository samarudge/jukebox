package app

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
  "github.com/samarudge/jukebox/db"
)

var router = gin.New()

func Start(bind string){

  d := db.Db()
  d.AutoMigrate(&models.User{})
  d.AutoMigrate(&models.Oauth2{})

  router.Use(helpers.Logger())
  router.Use(gin.Recovery())
  router.Use(helpers.Auth())

  // TODO: dynamicly figure out this folder
  //router.LoadHTMLGlob("src/jukebox/views/**")

  loadRoutes(router)
  loadJobs()

  router.Run(bind)
}
