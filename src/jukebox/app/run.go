package app

import(
  "github.com/gin-gonic/gin"
  "jukebox/helpers"
  "jukebox/models"
  "jukebox/db"
)

var router = gin.New()

func Start(bind string){

  d := db.Db()
  d.AutoMigrate(&models.User{})

  router.Use(helpers.Logger())
  router.Use(gin.Recovery())
  router.Use(helpers.Auth())

  // TODO: dynamicly figure out this folder
  //router.LoadHTMLGlob("src/jukebox/views/**")

  loadRoutes(router)
  loadJobs()

  router.Run(bind)
}
