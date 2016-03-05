package app

import(
  "github.com/samarudge/jukebox/controllers"
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
)

func loadRoutes(router *gin.Engine){
  router.GET("/", helpers.RequireRoom(), controllers.Index)
  router.GET("/auth/callback", controllers.AuthCallback)
  router.GET("/auth/logout", controllers.AuthLogout)

  router.GET("/rooms", controllers.RoomList)
  router.POST("/rooms", helpers.RequireAuth(), controllers.RoomCreate)
  router.POST("/rooms/:roomId/join", helpers.RequireAuth(), controllers.RoomJoin)

  userActions := router.Group("/users")
  userActions.Use(helpers.AuthorizedUser())
  userActions.GET("/:userId", controllers.UserInfo)
  userActions.POST("/:userId/renewToken", controllers.UserRenewToken)
}
