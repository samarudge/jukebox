package helpers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/models"
  "github.com/samarudge/jukebox/db"
  log "github.com/Sirupsen/logrus"
  "time"
  "strconv"
  "net/url"
  "fmt"
)

func ClearAuthCookie(c *gin.Context){
  c.SetCookie(
    "jukebox_user",
    "",
    -1,
    "/",
    "",
    false,
    true,
  )
}

func AuthorizedUser() gin.HandlerFunc{
  /*
    Is user authorized to view self or other user
  */

  return func(c *gin.Context){
    userId := c.Param("userId")
    authUser := models.User{}
    userInterface, _ := c.Get("authUser")
    if userInterface != nil{
      authUser = userInterface.(models.User)
    }

    if userId != strconv.FormatUint(uint64(authUser.ID), 10) && !authUser.IsAdmin {
      Send403(c, "You are not authorized to view this uer")
    } else {
      c.Next()
    }
  }
}

func RequireAuth() gin.HandlerFunc{
  return func(c *gin.Context){
    userId, _ := c.Get("authUserId")

    if userId == nil {
      Send403(c, "You must log in to view this page")
    } else {
      c.Next()
    }
  }
}

func RequireRoom() gin.HandlerFunc{
  return func(c *gin.Context){
    room, _ := c.Get("currentRoom")

    if room == nil{
      c.Redirect(302, "/rooms")
      c.Abort()
    } else {
      c.Next()
    }
  }
}

func RequireAdmin() gin.HandlerFunc{
  return func(c *gin.Context){
    userId, _ := c.Get("authUserId")
    u := models.User{}

    if userId != nil{
      u = c.MustGet("authUser").(models.User)
    }

    d := db.Db()
    if d.NewRecord(u) || !u.IsAdmin {
      Send403(c, "You must log in to view this page")
    } else {
      c.Next()
    }
  }
}

func Auth() gin.HandlerFunc{
  /*
    Process authentication data
  */

  return func(c *gin.Context) {
    authUserCookie, err := c.Cookie("jukebox_user")
    if err == nil{
      authUserId, err := VerifyValue(authUserCookie)

      d := db.Db()

      u := models.User{}
      u.ById(authUserId)
      a := u.Auth()

      if err != nil || d.NewRecord(u) || ! a.AuthValid{
        log.WithFields(log.Fields{
          "authCookie": authUserCookie,
          "error": err,
          "userNew": d.NewRecord(u),
          "authNew": d.NewRecord(a),
          "authValid": a.AuthValid,
        }).Warning("Invalid user cookie")

        ClearAuthCookie(c)
      } else {
        provider := a.LoadProvider()
        authExpiry := a.LastAuth.Add(provider.Provider().ReauthEvery).Sub(time.Now().UTC()).Minutes()
        if authExpiry <= 0{
          _, err := a.EnsureAuth(a.CreateToken())
          if err != nil{
            ClearAuthCookie(c)
            Send500(c, fmt.Sprintf("%s (%s)", "Reauth Error", err))
            return
          }
        }

        c.Set("authUserId", authUserId)
        c.Set("authUser", u)

        if u.IsAdmin{
          c.Set("isAdmin", true)
        } else {
          c.Set("isAdmin", false)
        }

        room := models.Room{}
        room.ById(strconv.FormatUint(uint64(u.RoomID), 10))
        if !d.NewRecord(room){
          log.WithFields(log.Fields{
            "userId": u.ID,
            "roomId": room.ID,
            "room": room.Name,
          }).Debug("Loaded active room")
          c.Set("currentRoom", room)
        }

        if time.Now().UTC().Sub(u.LastSeen).Minutes() > 5 {
          log.WithFields(log.Fields{
            "User": authUserId,
          }).Debug("Updating User Last Seen")

          u.LastSeen = time.Now().UTC()

          d.Save(&u)
        }
      }
    }

    from := c.Request.URL.String()
    pageFrom := SignValue(from)

    var loginLinks []map[string]string

    for _,providerName := range auth.ConfiguredProviders{
      p := auth.Providers[providerName]

      providerLoginLink := url.URL{}
      providerLoginLink.Path = "/auth/login"
      q := providerLoginLink.Query()
      q.Set("from", pageFrom)
      q.Set("provider", providerName)
      providerLoginLink.RawQuery = q.Encode()

      if p.ProviderSlug() == "spotify"{
        c.Set("spotifyLogin", providerLoginLink.String())
      } else {
        loginLinks = append(loginLinks, map[string]string{
          "name": p.Provider().Name,
          "loginLink": providerLoginLink.String(),
        })
      }
    }
    c.Set("loginLinks", loginLinks)

    logoutLink := url.URL{}
    logoutLink.Path = "/auth/logout"
    q := logoutLink.Query()
    q.Set("from", pageFrom)
    logoutLink.RawQuery = q.Encode()
    c.Set("logoutLink", logoutLink.String())

    s := models.Spotify{}
    s.LoadSystem()
    c.Set("spotifyConfigured", s.Exists())

    c.Next()
  }
}
