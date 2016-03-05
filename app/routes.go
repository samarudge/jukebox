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

  router.GET("/admin", helpers.RequireAdmin(), controllers.AdminIndex)

  router.GET("/users", helpers.RequireAdmin(), controllers.UserList)
  router.GET("/users/:userId", helpers.AuthorizedUser(), controllers.UserInfo)
  router.POST("/users/:userId/renewToken", helpers.AuthorizedUser(), controllers.UserRenewToken)
}
