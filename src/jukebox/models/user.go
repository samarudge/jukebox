package models

import(
  "github.com/jinzhu/gorm"
  "golang.org/x/oauth2"
  "jukebox/auth"
  "jukebox/db"
  log "github.com/Sirupsen/logrus"
  "fmt"
  "time"
)

type User struct{
  gorm.Model
  auth.UserData
  Provider      string
  AccessToken   string
  RefreshToken  string
  AuthValid     bool
  LastAuth      time.Time
  TokenExpires  time.Time
  LastSeen      time.Time
}

func (u *User) ById(userId string){
  d := db.Db()
  d.Where("id = ?", userId).First(&u)
}

func (u *User) CreateOrUpdateFromToken(token *oauth2.Token) error{
  d := db.Db()
  userData, err := auth.Provider.GetUserData(token)
  if err != nil{
    u.AuthValid = false
    d.Save(&u)
    return err
  }

  u.Model = gorm.Model{}
  u.UserData = userData

  d.Where("provider_id = ? AND provider = ?", u.ProviderId, auth.Provider.Provider().Name).First(&u)

  if d.NewRecord(u) {
    u.Provider = auth.Provider.Provider().Name

    d.Create(&u)

    log.WithFields(log.Fields{
      "userId": u.ID,
      "name": u.Name,
      "providerId": u.ProviderId,
    }).Debug("New User")
  }

  u.AuthValid = true
  u.AccessToken = token.AccessToken
  u.RefreshToken = token.RefreshToken
  u.TokenExpires = token.Expiry.UTC()
  u.LastAuth = time.Now().UTC()
  d.Save(&u)

  log.WithFields(log.Fields{
    "userId": u.ID,
    "name": u.Name,
    "providerId": u.ProviderId,
  }).Debug("Login")

  return nil
}

func (u *User) CreateToken() *oauth2.Token{
  t := oauth2.Token{}
  t.AccessToken = u.AccessToken
  t.RefreshToken = u.RefreshToken
  t.Expiry = u.TokenExpires
  log.WithFields(log.Fields{
    "user":u.ID,
  }).Debug("Loaded token")
  return &t
}

func (u *User) RenewAuthToken() error{
  if !u.AuthValid {
    return fmt.Errorf("Auth not valid")
  }

  if time.Now().UTC().Sub(u.LastSeen).Hours() > 24*14 {
    return fmt.Errorf("User not seen for 14 days")
  }

  tkn := u.CreateToken()
  tkn.Expiry = time.Now().Add(time.Minute*-5)

  c := auth.Provider.OauthConfig()
  tks := c.TokenSource(oauth2.NoContext, tkn)
  newToken, err := tks.Token()

  if err != nil{
    log.WithFields(log.Fields{
      "error": err,
    }).Warning("Could not refresh auth token")
    u.AuthValid = false
    d := db.Db()
    d.Save(&u)
    return err
  }

  log.WithFields(log.Fields{
    "user": u.ID,
    "newExpiry": newToken.Expiry,
  }).Debug("Refreshed auth token")

  return u.CreateOrUpdateFromToken(newToken)
}

func (u *User) CheckAuth() error{
  tkn := u.CreateToken()
  err := u.CreateOrUpdateFromToken(tkn)

  if err != nil{
    log.WithFields(log.Fields{
      "error": err,
    }).Warning("Could not verify auth")
  }

  return err
}

func (u User) ProfileLink() string{
  return fmt.Sprintf("/users/%d", u.ID)
}

func (u User) LastSeenStamp() string{
  return u.LastSeen.Format("Mon Jan 2 2006 15:04:05 MST")
}

func readableExpiry(t time.Time) string{
  return fmt.Sprintf("%s (in %s)", t.Format("Mon Jan 2 2006 15:04:05 MST"), t.Sub(time.Now().UTC()).String())
}

func (u User) TokenExpiresIn() string{
  return readableExpiry(u.TokenExpires)
}

func (u User) AuthExpiresIn() string{
  return readableExpiry(u.LastAuth.Add(auth.Provider.Provider().ReauthEvery))
}
