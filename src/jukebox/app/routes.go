package app

import(
  "jukebox/controllers"
  "github.com/gin-gonic/gin"
)

func loadRoutes(router *gin.Engine){
  router.GET("/", controllers.RoomList)
  router.GET("/auth/callback", controllers.AuthCallback)

  router.GET("/users/:userId", controllers.UserInfo)
}
