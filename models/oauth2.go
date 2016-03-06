package models

import(
  "github.com/jinzhu/gorm"
  "time"
  "golang.org/x/oauth2"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/db"
  log "github.com/Sirupsen/logrus"
  "fmt"
)

type Oauth2 struct{
  gorm.Model
  Provider      string
  ProviderId    string
  AccessToken   string
  RefreshToken  string
  AuthValid     bool
  LastAuth      time.Time
  TokenExpires  time.Time
}

func (a *Oauth2) ExpiringToken() bool{
  return !(a.RefreshToken == "" && a.TokenExpires.IsZero())
}

func (a *Oauth2) EnsureAuth(token *oauth2.Token) (auth.UserData, error){
  d := db.Db()
  var userData auth.UserData

  a.Model = gorm.Model{}

  if a.ExpiringToken() && a.TokenExpires.Sub(time.Now().UTC()).Minutes() < 5{
    err := a.RenewAuthToken()
    token = a.CreateToken()

    if err != nil{
      return userData, nil
    }
  }

  providerId, userData, err := auth.Provider.GetUserData(token)
  d.Where("provider_id = ?", providerId).First(&a)

  if err != nil{
    if !d.NewRecord(a){
      a.AuthValid = false
      d.Save(&a)
    }
    return userData, err
  }

  a.ProviderId = providerId
  a.AuthValid = true
  a.AccessToken = token.AccessToken
  a.RefreshToken = token.RefreshToken
  a.TokenExpires = token.Expiry.UTC()
  a.LastAuth = time.Now().UTC()

  if d.NewRecord(a){
    a.Provider = auth.Provider.ProviderSlug()
  }

  d.Save(&a)
  return userData, nil
}

func (a *Oauth2) CreateToken() *oauth2.Token{
  t := oauth2.Token{}

  t.AccessToken = a.AccessToken
  t.RefreshToken = a.RefreshToken
  t.Expiry = a.TokenExpires
  log.WithFields(log.Fields{
    "authID": a.ID,
  }).Debug("Loaded token")
  return &t
}

func (a *Oauth2) RenewAuthToken() error{
  if !a.ExpiringToken(){
    return nil
  }

  if !a.AuthValid {
    return fmt.Errorf("Auth not valid")
  }

  tkn := a.CreateToken()
  tkn.Expiry = time.Now().Add(time.Minute*-5)

  c := auth.Provider.OauthConfig()
  tks := c.TokenSource(oauth2.NoContext, tkn)
  newToken, err := tks.Token()

  if err != nil{
    log.WithFields(log.Fields{
      "error": err,
    }).Warning("Could not refresh auth token")
    a.AuthValid = false
    d := db.Db()
    d.Save(&a)
    return err
  }

  log.WithFields(log.Fields{
    "auth": a.ID,
    "newExpiry": newToken.Expiry,
  }).Debug("Refreshed auth token")

  return nil
}

func readableExpiry(t time.Time) string{
  return fmt.Sprintf("%s (in %s)", t.Format("Mon Jan 2 2006 15:04:05 MST"), t.Sub(time.Now().UTC()).String())
}

func (a Oauth2) TokenExpiresIn() string{
  if a.ExpiringToken(){
    return readableExpiry(a.TokenExpires)
  } else {
    return "Never"
  }
}

func (a Oauth2) AuthExpiresIn() string{
  return readableExpiry(a.LastAuth.Add(auth.Provider.Provider().ReauthEvery))
}

func JobRenewAuth(){
  authFilter := time.Now().UTC().Add(auth.Provider.Provider().ReauthEvery*-1)
  auths := []Oauth2{}
  d := db.Db()
  var authCount int

  authQuery := d.Where("last_auth < ? and auth_valid = ?", authFilter, true)
  authQuery.Find(&auths).Count(&authCount)

  if authCount > 0{
    log.WithFields(log.Fields{
      "authCount":authCount,
    }).Debug("Renewing auth")

    for i,_ := range auths{
      a := auths[i]
      log.WithFields(log.Fields{
        "auth": a.ID,
        "provider": a.Provider,
        "expired": a.AuthExpiresIn(),
      }).Debug("Doing reauth")

      t := a.CreateToken()
      _, err := a.EnsureAuth(t)
      if err != nil{
        log.WithFields(log.Fields{
          "auth": a.ID,
          "err": err,
        }).Warning("Could not do reauth")
      }
    }
  }
}
