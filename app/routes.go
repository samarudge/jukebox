package app

import(
  "github.com/samarudge/jukebox/controllers"
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
)

func loadRoutes(router *gin.Engine){
  router.GET("/", helpers.RequireRoom(), controllers.Index)
  router.GET("/auth/login", controllers.AuthLogin)
  router.GET("/auth/callback/:providerName", controllers.AuthCallback)
  router.GET("/auth/logout", controllers.AuthLogout)

  router.GET("/rooms", controllers.RoomList)
  router.POST("/rooms", helpers.RequireAuth(), controllers.RoomCreate)
  router.POST("/rooms/:roomId/join", helpers.RequireAuth(), controllers.RoomJoin)

  adminRoutes := router.Group("/")
  adminRoutes.Use(helpers.RequireAdmin())
  adminRoutes.GET("/admin", controllers.AdminIndex)
  adminRoutes.GET("/users", controllers.UserList)

  userRoutes := router.Group("/users")
  userRoutes.Use(helpers.AuthorizedUser())
  userRoutes.Use(controllers.UserContext)
  userRoutes.GET("/:userId", controllers.UserInfo)
  userRoutes.POST("/:userId", controllers.UserUpdate)
  userRoutes.POST("/:userId/renewToken", controllers.UserRenewToken)
}
