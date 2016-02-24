package controllers

import(
  "github.com/gin-gonic/gin"
  "jukebox/auth"
  "jukebox/helpers"
  "jukebox/models"
  "strconv"
  "net/url"
)

func AuthCallback(c *gin.Context){
  code := c.DefaultQuery("code", "")

  token, err := auth.Provider.DoExchange(code)

  if err != nil{
    c.Status(500)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Error during authentication",
      "errorDetails": err,
    })
  } else {
    stateRaw := c.DefaultQuery("state", "")
    state, err := helpers.VerifyValue(stateRaw)
    if err != nil{
      c.Status(403)
      helpers.Render(c, "error.html", gin.H{
        "errorTitle": "Error during authentication",
        "errorDetails": "State mismatch",
      })
      return
    }

    u := models.User{}
    u.CreateOrUpdateFromToken(token)

    cookieVal := helpers.SignValue(strconv.FormatUint(uint64(u.ID), 10))
    c.SetCookie(
      "jukebox_user",
      cookieVal,
      60*60*24*14, // Cookie valid for 14 days
      "/",
      "",
      false,
      true,
    )

    fromPage, _ := url.Parse(state)
    c.Redirect(302, fromPage.Path)
  }
}
