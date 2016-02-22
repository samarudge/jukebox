package controllers

import(
  "github.com/gin-gonic/gin"
  "jukebox/helpers"
  "jukebox/models"
  "strconv"
)

func UserInfo(c *gin.Context){
  userId := c.Param("userId")
  userInterface, _ := c.Get("authUser")
  authUser := userInterface.(models.User)

  if userId != strconv.FormatUint(uint64(authUser.ID), 10) {
    c.Status(403)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Authentication Error",
      "errorDetails": "You are not authorized to view this user",
    })
  } else {
    u := models.User{}
    u.ById(userId)

    helpers.Render(c, "users/info.html", gin.H{
      "user": u,
    })
  }
}
