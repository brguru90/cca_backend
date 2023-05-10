package my_modules

import (
	"travel_planner/src/app_cron_jobs"

	"github.com/robfig/cron/v3"
)

var CRON_JOBS *cron.Cron

func InitCronJobs() {
	CRON_JOBS = cron.New()
	CRON_JOBS.AddFunc("*/15 * * * *", app_cron_jobs.ClearExpiredToken)
	CRON_JOBS.Start()
}
