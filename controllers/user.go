package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
  "github.com/samarudge/jukebox/db"
  "fmt"
  log "github.com/Sirupsen/logrus"
)

func UserContext(c *gin.Context){
  userId := c.Param("userId")
  u := models.User{}
  u.ById(userId)
  if u.ID == 0{
    helpers.Send404(c, "User not found")
  } else {
    c.Set("contextUser", u)
    c.Next()
  }
}

func UserInfo(c *gin.Context){
  u := c.MustGet("contextUser").(models.User)

  helpers.Render(c, "users/info.html", gin.H{
    "user": u,
    "auth": u.Auth(),
    "spotify": u.GetSpotify(),
  })
}

func UserUpdate(c *gin.Context){
  u := c.MustGet("contextUser").(models.User)

  adminStr := fmt.Sprintf("%t", u.IsAdmin)
  isAdmin := c.DefaultPostForm("IsAdmin", adminStr)
  authUserAdmin, _ := c.Get("isAdmin")
  if isAdmin != adminStr && authUserAdmin.(bool){
    log.WithFields(log.Fields{
      "newStatus": isAdmin,
      "currentStatus": adminStr,
    }).Debug("Changing user admin state")

    if isAdmin == "true"{
      u.IsAdmin = true
    } else {
      u.IsAdmin = false
    }
  }

  d := db.Db()
  d.Save(&u)
  c.Redirect(302, u.ProfileLink())
}

func UserList(c *gin.Context){
  d := db.Db()

  var users []models.User
  d.Order("name").Find(&users)
  helpers.Render(c, "users/list.html", gin.H{
    "users": users,
  })
}
