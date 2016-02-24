package controllers

import(
  "github.com/gin-gonic/gin"
  "jukebox/helpers"
  "jukebox/models"
)

func UserInfo(c *gin.Context){
  u := c.MustGet("userControllerRequest").(models.User)

  helpers.Render(c, "users/info.html", gin.H{
    "user": u,
  })
}

func UserRenewToken(c *gin.Context){
  u := c.MustGet("userControllerRequest").(models.User)
  err := u.RenewAuthToken()

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
