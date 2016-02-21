package helpers

import(
  "github.com/gin-gonic/gin"
  "jukebox/auth"
)

var AuthProvider = auth.Google

func Auth() gin.HandlerFunc{
  /*
    Process authentication data
  */

  return func(c *gin.Context) {
    c.Set("authProvider", AuthProvider)

    from := c.DefaultQuery("from", "/")
    c.Set("loginLink", AuthProvider.LoginLink(from))

    c.Next()
  }
}
