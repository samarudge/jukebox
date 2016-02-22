package helpers

import(
  "github.com/gin-gonic/gin"
  "jukebox/auth"
  "jukebox/models"
  log "github.com/Sirupsen/logrus"
)

func Auth() gin.HandlerFunc{
  /*
    Process authentication data
  */

  return func(c *gin.Context) {
    c.Set("authProvider", auth.AuthProvider)

    authUserCookie, err := c.Cookie("jukebox_user")
    if err == nil{
      authUserId, err := VerifyValue(authUserCookie)
      if err != nil{
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
        u := models.User{}
        u.ById(authUserId)

        c.Set("authUser", u)
      }
    }

    from := c.DefaultQuery("from", "/")
    c.Set("loginLink", auth.AuthProvider.LoginLink(from))

    c.Next()
  }
}
