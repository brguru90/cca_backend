package app_cron_jobs

import (
	"cca/src/database/database_connections"
	"context"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
)

var CRON_JOBS *cron.Cron

func InitCronJobs() {

	database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.DeleteMany(context.Background(), bson.M{})
	CRON_JOBS = cron.New()
	CRON_JOBS.AddFunc("*/15 * * * *", ClearExpiredToken)
	CRON_JOBS.AddFunc("*/1 * * * *", VideoStreamGenerationCron)
	CRON_JOBS.Start()
}
