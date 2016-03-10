package models

import(
  "github.com/jinzhu/gorm"
  "github.com/samarudge/jukebox/auth"
  "github.com/samarudge/jukebox/db"
  log "github.com/Sirupsen/logrus"
)

type Spotify struct{
  gorm.Model
  auth.UserData
  Oauth2        Oauth2
  Oauth2ID      uint64
  SystemAccount bool
}

func (s *Spotify) ByAuth(a *Oauth2){
  d := db.Db()
  d.Where("oauth2_id = ?", a.ID).First(&s)
}

func (s Spotify) Auth() Oauth2{
  d := db.Db()
  a := Oauth2{}
  d.Model(&s).Related(&a)
  return a
}

func (s Spotify) Exists() bool{
  d := db.Db()
  return !d.NewRecord(s)
}

func (s *Spotify) CreateOrUpdate(a Oauth2){
  d := db.Db()
  d.Where("oauth2_id = ?", a.ID).First(&s)

  if d.NewRecord(s) {
    d.Model(&s).Related(&a)
    s.Model = gorm.Model{}
    s.Oauth2 = a
    d.Create(&s)
    log.WithFields(log.Fields{
      "authId": a.ID,
      "spotifyId": s.ID,
    }).Debug("Created new spotify")
  } else {
    d.Save(&s)
    log.WithFields(log.Fields{
      "authId": a.ID,
      "spotifyId": s.ID,
    }).Debug("Loaded spotify")
  }
}

func (s *Spotify) MakeSystem(){
  d := db.Db()
  t := d.Begin()
  t.Model(Spotify{}).UpdateColumn("system_account", false)
  s.SystemAccount = true
  t.Save(&s)
  t.Commit()
  log.WithFields(log.Fields{
    "spotifyId": s.ID,
  }).Debug("Made system spotify account")
}

func (s *Spotify) LoadSystem(){
  d := db.Db()
  d.Where("system_account = ?", true).First(&s)
}
