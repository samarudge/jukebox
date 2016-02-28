package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
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
    err = u.LoginOrSignup(token)

    if err != nil{
      c.Status(403)
      helpers.Render(c, "error.html", gin.H{
        "errorTitle": "Error during authentication",
        "errorDetails": err,
      })
      return
    }

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

    c.Redirect(302, state)
  }
}

func AuthLogout(c *gin.Context){
  fromPageRaw := c.DefaultQuery("from", "")
  fromPage, err := helpers.VerifyValue(fromPageRaw)

  if err != nil{
    fromPage = "/"
  }

  helpers.ClearAuthCookie(c)

  from, _ := url.Parse(fromPage)
  c.Redirect(302, from.Path)
}
