package app

import(
  "jukebox/controllers"
  "github.com/gin-gonic/gin"
  "jukebox/helpers"
)

func loadRoutes(router *gin.Engine){
  router.GET("/", controllers.RoomList)
  router.GET("/auth/callback", controllers.AuthCallback)

  userActions := router.Group("/users")
  userActions.Use(helpers.AuthorizedUser())
  userActions.GET("/:userId", controllers.UserInfo)
  userActions.POST("/:userId/renewToken", controllers.UserRenewToken)
}
