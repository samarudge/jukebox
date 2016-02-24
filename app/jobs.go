package app

import(
  "github.com/jasonlvhit/gocron"
  "github.com/samarudge/jukebox/models"
)

func loadJobs(){
  gocron.Every(30).Seconds().Do(models.JobRenewUserAuth)

  gocron.Start()
}
