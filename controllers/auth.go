package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
  "strconv"
  "net/url"
  "fmt"
)

func AuthLogin(c *gin.Context){
  providerName := c.DefaultQuery("provider", "")
  p, found := auth.Providers[providerName]
  if !found{
    helpers.Send404(c, "Invalid provider")
    return
  }

  from := c.DefaultQuery("from", "")

  loginLink := p.LoginLink(from)
  c.Redirect(302, loginLink)
}

func AuthCallback(c *gin.Context){
  code := c.DefaultQuery("code", "")
  providerName := c.Param("providerName")

  provider, found := auth.Providers[providerName]
  if found == false{
    helpers.Send404(c, "Invalid provider")
    return
  }

  token, err := provider.DoExchange(code)

  if err != nil{
    helpers.Send500(c, fmt.Sprintf("%s (%s)", "Error during authentication", err))
  } else {
    stateRaw := c.DefaultQuery("state", "")
    state, err := helpers.VerifyValue(stateRaw)
    if err != nil{
      helpers.Send403(c, "State mismatch")
      return
    }

    u := models.User{}
    err = u.LoginOrSignup(provider, token)

    if err != nil{
      helpers.Send500(c, fmt.Sprintf("%s (%s)", "Error during authentication", err))
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
