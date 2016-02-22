package helpers

import(
  "github.com/gin-gonic/gin"
  "jukebox/auth"
  "jukebox/models"
  "jukebox/db"
  log "github.com/Sirupsen/logrus"
  "time"
)

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

      if err != nil || d.NewRecord(u){
        log.WithFields(log.Fields{
          "authCookie": authUserCookie,
          "error": err,
        }).Warning("Invalid user cookie")

        c.SetCookie(
          "jukebox_user",
          "",
          -1,
          "/",
          "",
          false,
          true,
        )
      } else {
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

    from := c.DefaultQuery("from", "/")
    state := SignValue(from)
    c.Set("loginLink", auth.Provider.LoginLink(state))

    c.Next()
  }
}
