package app_cron_jobs

import (
	"cca/src/database/database_connections"

	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
)

var CRON_JOBS *cron.Cron

func InitCronJobs() {

	if err := database_connections.RedisPoolDel("video_stream_generation_in_progress"); err != nil {
		log.WithFields(log.Fields{
			"Error": err,
		}).Errorln("Unable to delete in redis pool")
	}

	CRON_JOBS = cron.New()
	CRON_JOBS.AddFunc("*/15 * * * *", ClearExpiredToken)
	CRON_JOBS.AddFunc("*/1 * * * *", VideoStreamGeneration)
	CRON_JOBS.Start()
}
