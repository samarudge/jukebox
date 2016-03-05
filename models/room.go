package models

import(
  "github.com/jinzhu/gorm"
  "github.com/samarudge/jukebox/db"
  "fmt"
  log "github.com/Sirupsen/logrus"
)

type Room struct{
  gorm.Model
  Name      string
  Active    bool
  Creator   User
  CreatorID uint64
  Members   []User
}

func (r *Room) ById(roomId string){
  d := db.Db()
  d.Where("id = ?", roomId).First(&r)
}

func (r *Room) Create(user User, roomName string){
  d := db.Db()

  d.Model(&r).Related(&user)

  r.Model = gorm.Model{}
  r.Name = roomName
  r.Creator = user
  r.Active = true

  d.Create(&r)
  log.WithFields(log.Fields{
    "roomName": r.Name,
    "roomId": r.ID,
  }).Debug("Created room")
}

func (r *Room) JoinUser(u User) error{
  d := db.Db()
  d.Model(&r).Related(&u)

  r.Members = append(r.Members, u)

  d.Save(&r)
  return nil
}

func (r Room) Url() string{
  return fmt.Sprintf("/rooms/%d", r.ID)
}
