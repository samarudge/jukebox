package controllers

import(
  "github.com/gin-gonic/gin"
  "jukebox/auth"
  "jukebox/helpers"
  "fmt"
  "jukebox/models"
  "strconv"
  "time"
)

func AuthLogin(c *gin.Context){
  from := c.DefaultQuery("from", "/")

  redirectUrl := fmt.Sprintf("http://%s/auth/callback?from=%s", c.Request.Host, from)
  c.Redirect(302, auth.AuthProvider.LoginLink(redirectUrl))
}

func AuthCallback(c *gin.Context){
  code := c.DefaultQuery("code", "")

  token, err := auth.AuthProvider.DoExchange(code)

  if err != nil{
    c.Status(500)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Error during authentication",
      "errorDetails": err,
    })
  } else {
    u := models.User{}
    u.CreateOrUpdateFromToken(token)

    cookieVal := helpers.SignValue(strconv.FormatUint(uint64(u.ID), 10))
    c.SetCookie(
      "jukebox_user",
      cookieVal,
      int(time.Since(token.Expiry).Seconds())*-1,
      "/",
      "",
      false,
      true,
    )

    c.Redirect(302, "/")
  }
}
