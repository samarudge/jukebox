package db

import(
  "github.com/jinzhu/gorm"
  _ "github.com/mattn/go-sqlite3"
  log "github.com/Sirupsen/logrus"
)

// From https://gist.github.com/bnadland/2e4287b801a47dcfcc94
type GormLogger struct {}

func (*GormLogger) Print(v ...interface{}) {
  if v[0] == "sql" {
    log.WithFields(log.Fields{"module": "gorm", "type": "sql", "params":v[4],}).Debug(v[3])
  }
  if v[0] == "log" {
    log.WithFields(log.Fields{"module": "gorm", "type": "log"}).Info(v[2])
  }
}

func Db() (gorm.DB){
  db, _ := gorm.Open("sqlite3", "storage.db")
  db.SetLogger(&GormLogger{})
  db.LogMode(true)

  return db
}
