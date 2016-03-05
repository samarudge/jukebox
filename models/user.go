package models

import(
  "github.com/jinzhu/gorm"
  "golang.org/x/oauth2"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/db"
  log "github.com/Sirupsen/logrus"
  "fmt"
  "time"
)

type User struct{
  gorm.Model
  auth.UserData
  Oauth2        Oauth2
  Oauth2ID      uint64
  RoomID        uint64
  LastSeen      time.Time
  IsAdmin       bool
}

func (u *User) ById(userId string){
  d := db.Db()
  d.Where("id = ?", userId).First(&u)
}

func (u *User) ByAuth(a *Oauth2){
  d := db.Db()
  d.Where("oauth2_id = ?", a.ID).First(&u)
}

func (u User) Auth() Oauth2{
  d := db.Db()
  a := Oauth2{}
  d.Model(&u).Related(&a)
  return a
}

func (u *User) LoginOrSignup(token *oauth2.Token) error{
  a := Oauth2{}
  UserData, err := a.EnsureAuth(token)
  if err != nil{
    return err
  }

  d := db.Db()
  d.Where("oauth2id = ?", a.ID).First(&u)
  d.Model(&u).Related(&a)

  u.Model = gorm.Model{}
  u.UserData = UserData

  if d.NewRecord(u) {
    u.Oauth2 = a

    userCount := 0
    d.Find(&User{}).Count(&userCount)
    if userCount == 0{
      log.Info("First user, promoting to admin")
      u.IsAdmin = true
    }

    d.Create(&u)

    log.WithFields(log.Fields{
      "userId": u.ID,
      "name": u.Name,
      "providerId": a.ProviderId,
      "provider": a.Provider,
      "authId": a.ID,
      "authKey": u.Oauth2ID,
    }).Debug("New User")
  } else {
    d.Save(&u)

    log.WithFields(log.Fields{
      "userId": u.ID,
      "name": u.Name,
      "providerId": a.ProviderId,
      "authId": a.ID,
    }).Debug("Login")
  }

  return nil
}

func (u User) ProfileLink() string{
  return fmt.Sprintf("/users/%d", u.ID)
}

func (u User) LastSeenStamp() string{
  return u.LastSeen.Format("Mon Jan 2 2006 15:04:05 MST")
}
