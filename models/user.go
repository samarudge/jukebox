package models

import(
  "github.com/jinzhu/gorm"
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
  Spotify       Spotify
  SpotifyID     uint64
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

func (u User) HasSpotify() bool{
  hasSpotify, _ := u.getSpotify()
  return hasSpotify
}

func (u User) GetSpotify() Spotify{
  _, spotify := u.getSpotify()
  return spotify
}

func (u User) getSpotify() (bool, Spotify){
  d := db.Db()
  s := Spotify{}
  d.Model(&u).Related(&s)
  return !d.NewRecord(s), s
}

func (u *User) LoginOrSignup(a Oauth2){
  d := db.Db()
  d.Where("oauth2_id = ?", a.ID).First(&u)
  d.Model(&u).Related(&a)

  if d.NewRecord(u) {
    u.Model = gorm.Model{}
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
      "name": a.Name,
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
}

func (u *User) LinkSpotify(s Spotify){
  d := db.Db()
  d.Model(&u).Related(&s)
  u.Spotify = s
  log.WithFields(log.Fields{
    "spotifyId": s.ID,
    "userId": u.ID,
  }).Info("Linking Spotify")
  d.Save(&u)
}

func (u User) ProfileLink() string{
  return fmt.Sprintf("/users/%d", u.ID)
}

func (u User) LastSeenStamp() string{
  return u.LastSeen.Format("Mon Jan 2 2006 15:04:05 MST")
}
