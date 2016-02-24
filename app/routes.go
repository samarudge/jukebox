package app

import(
  "github.com/samarudge/jukebox/controllers"
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
)

func loadRoutes(router *gin.Engine){
  router.GET("/", controllers.RoomList)
  router.GET("/auth/callback", controllers.AuthCallback)
  router.GET("/auth/logout", controllers.AuthLogout)

  userActions := router.Group("/users")
  userActions.Use(helpers.AuthorizedUser())
  userActions.GET("/:userId", controllers.UserInfo)
  userActions.POST("/:userId/renewToken", controllers.UserRenewToken)
}
