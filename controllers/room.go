package controllers

import(
  "github.com/gin-gonic/gin"
  "github.com/samarudge/jukebox/helpers"
  "github.com/samarudge/jukebox/models"
  "github.com/samarudge/jukebox/db"
)

func RoomCreate(c *gin.Context){
  roomName := c.PostForm("roomName")

  if len(roomName) == 0{
    helpers.Render(c, "rooms/new.html", gin.H{"error":"You must enter a room name"})
    return
  }

  r := models.Room{}
  user, _ := c.Get("authUser")
  r.Create(user.(models.User), roomName)
  c.Redirect(302, "/rooms")
}

func loadRoom(c *gin.Context) (models.Room, bool){
  d := db.Db()
  r := models.Room{}
  r.ById(c.Param("roomId"))

  if d.NewRecord(r){
    c.Status(404)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Room Not Found",
      "errorDetails": "Room was not found",
    })
    c.Abort()
    return r, false
  }
  return r, true
}

func RoomJoin(c *gin.Context){
  room, found := loadRoom(c)
  if !found{
    return
  }

  u := c.MustGet("authUser").(models.User)

  err := room.JoinUser(u)
  if err != nil{
    c.Status(500)
    helpers.Render(c, "error.html", gin.H{
      "errorTitle": "Error joining room",
      "errorDetails": err,
    })
  } else {
    c.Redirect(302, "/")
  }
}

func RoomList(c *gin.Context){
  d := db.Db()

  var rooms []models.Room
  d.Order("name").Find(&rooms)
  helpers.Render(c, "rooms/list.html", gin.H{
    "rooms": rooms,
  })
}
