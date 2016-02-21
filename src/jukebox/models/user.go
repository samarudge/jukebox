package models

import(
  "github.com/jinzhu/gorm"
  "golang.org/x/oauth2"
  "jukebox/auth"
  "jukebox/helpers"
  "jukebox/db"
  log "github.com/Sirupsen/logrus"
)

type User struct{
  gorm.Model
  auth.UserData
  Provider      string
  AccessToken   string
}

func (u *User) CreateOrUpdateFromToken(token *oauth2.Token){
  userData, _ := helpers.AuthProvider.UserData(token)
  u.Model = gorm.Model{}
  u.UserData = userData

  d := db.Db()

  d.Where("provider_id = ? AND provider = ?", u.ProviderId, helpers.AuthProvider.Name).First(&u)

  if d.NewRecord(u) {
    u.Provider = helpers.AuthProvider.Name

    d.Create(&u)

    log.WithFields(log.Fields{
      "userId": u.ID,
      "name": u.Name,
      "providerId": u.ProviderId,
    }).Info("New User")
  }

  u.AccessToken = token.AccessToken
  d.Save(&u)

  log.WithFields(log.Fields{
    "userId": u.ID,
    "name": u.Name,
    "providerId": u.ProviderId,
  }).Info("Login")
}
