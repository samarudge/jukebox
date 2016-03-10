package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/db"
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

  var returnState string
  systemAccount := c.DefaultQuery("system_account", "0")
  if systemAccount == "1" {
    if providerName != "spotify"{
      helpers.Send404(c, "Can only link system account for Spotify provider")
      return
    }
    returnState = helpers.SignValue("system_account")
  } else {
    returnState = c.DefaultQuery("from", "")
  }

  loginLink := p.LoginLink(returnState)
  c.Redirect(302, loginLink)
}

func AuthCallback(c *gin.Context){
  code := c.DefaultQuery("code", "")
  providerName := c.Param("providerName")

  stateRaw := c.DefaultQuery("state", "")
  state, err := helpers.VerifyValue(stateRaw)
  if err != nil{
    helpers.Send403(c, "State mismatch")
    return
  }

  provider, found := auth.Providers[providerName]
  if found == false{
    helpers.Send404(c, "Invalid provider")
    return
  }

  token, err := provider.DoExchange(code)

  if err != nil{
    helpers.Send500(c, fmt.Sprintf("%s (%s)", "Error during authentication", err))
  } else {
    var returnPage string

    a := models.Oauth2{}
    a.Provider = provider.ProviderSlug()
    err := a.CreateOrUpdate(token)
    if err != nil{
      helpers.Send500(c, fmt.Sprintf("%s (%s)", "Error during authentication", err))
      return
    }

    s := models.Spotify{}

    if providerName == "spotify"{
      s.CreateOrUpdate(a)
    }

    if state == "system_account"{
      s.MakeSystem()
      returnPage = "/admin"
    } else {
      authUser, loggedInUser := c.Get("authUser")

      u := models.User{}
      if !loggedInUser{
        if providerName == "spotify"{
          helpers.Send403(c, "Spotify cannot be used as primary auth provider")
          return
        }

        u.LoginOrSignup(a)

        cookieVal := helpers.SignValue(strconv.FormatUint(uint64(u.ID), 10))
        c.SetCookie(
          "jukebox_user",
          cookieVal,
          60*60*24*14,
          "/",
          "",
          false,
          true,
        )
      } else {
        u = authUser.(models.User)
      }

      if providerName == "spotify" && state != "system_account"{
        u.LinkSpotify(s)
      }

      returnPage = state
    }

    c.Redirect(302, returnPage)
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

func AuthList(c *gin.Context){
  d := db.Db()

  var auths []models.Oauth2
  d.Find(&auths)
  helpers.Render(c, "auths/list.html", gin.H{
    "auths": auths,
  })
}
