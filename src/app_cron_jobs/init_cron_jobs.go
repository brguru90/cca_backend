package app_cron_jobs

import (
	"github.com/robfig/cron/v3"
)

var CRON_JOBS *cron.Cron

func InitCronJobs() {
	CRON_JOBS = cron.New()
	CRON_JOBS.AddFunc("*/15 * * * *", ClearExpiredToken)
	CRON_JOBS.AddFunc("*/1 * * * *", VideoStreamGenerationCron)
	CRON_JOBS.Start()
}
