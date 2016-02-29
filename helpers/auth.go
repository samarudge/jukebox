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

    if userId != strconv.FormatUint(uint64(authUser.ID), 10) {
      c.Status(403)
      Render(c, "error.html", gin.H{
        "errorTitle": "Authentication Error",
        "errorDetails": "You are not authorized to view this user",
      })
      c.Abort()
    } else {
      c.Set("userControllerRequest", authUser)
      c.Next()
    }
  }
}

func Auth() gin.HandlerFunc{
  /*
    Process authentication data
  */

  return func(c *gin.Context) {
    c.Set("authProvider", auth.Provider.Provider())

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
        authExpiry := a.LastAuth.Add(auth.Provider.Provider().ReauthEvery).Sub(time.Now().UTC()).Minutes()
        if authExpiry <= 0{
          _, err := a.EnsureAuth(a.CreateToken())
          if err != nil{
            ClearAuthCookie(c)
            c.Status(500)
            Render(c, "error.html", gin.H{
              "errorTitle": "Reauth Error",
              "errorDetails": err,
            })
            c.Abort()
            return
          }
        }

        c.Set("authUserId", authUserId)
        c.Set("authUser", u)

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
    c.Set("loginLink", auth.Provider.LoginLink(pageFrom))

    logoutLink := url.URL{}
    logoutLink.Path = "/auth/logout"
    q := logoutLink.Query()
    q.Set("from", pageFrom)
    logoutLink.RawQuery = q.Encode()
    c.Set("logoutLink", logoutLink.String())

    c.Next()
  }
}