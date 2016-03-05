package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
  "github.com/samarudge/jukebox/db"
)

func UserInfo(c *gin.Context){
  u := c.MustGet("userControllerRequest").(models.User)
  a := u.Auth()

  helpers.Render(c, "users/info.html", gin.H{
    "user": u,
    "auth": a,
  })
}

func UserRenewToken(c *gin.Context){
  u := c.MustGet("userControllerRequest").(models.User)
  a := u.Auth()
  err := a.RenewAuthToken()

  if err != nil{
    c.Status(500)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Token Renew Error",
      "errorDetails": err,
    })
  } else {
    c.Redirect(302, u.ProfileLink())
  }
}

func UserList(c *gin.Context){
  d := db.Db()

  var users []models.User
  d.Order("name").Find(&users)
  helpers.Render(c, "users/list.html", gin.H{
    "users": users,
  })
}
