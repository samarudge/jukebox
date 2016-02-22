package models

import(
  "github.com/jinzhu/gorm"
  "golang.org/x/oauth2"
  "jukebox/auth"
  "jukebox/db"
  log "github.com/Sirupsen/logrus"
  "fmt"
)

type User struct{
  gorm.Model
  auth.UserData
  Provider      string
  AccessToken   string
}

func (u *User) ById(userId string){
  d := db.Db()
  d.Where("id = ?", userId).First(&u)
}

func (u *User) CreateOrUpdateFromToken(token *oauth2.Token){
  userData, _ := auth.AuthProvider.UserData(token)
  u.Model = gorm.Model{}
  u.UserData = userData

  d := db.Db()

  d.Where("provider_id = ? AND provider = ?", u.ProviderId, auth.AuthProvider.Name).First(&u)

  if d.NewRecord(u) {
    u.Provider = auth.AuthProvider.Name

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

func (u User) ProfileLink() string{
  return fmt.Sprintf("/users/%d", u.ID)
}
