package app_cron_jobs

import (
	"cca/src/configs"
	"cca/src/database/database_connections"
	"context"
	"os/exec"

	"github.com/robfig/cron/v3"
	"go.mongodb.org/mongo-driver/bson"
)

var CRON_JOBS *cron.Cron

func InitCronJobs() {

	database_connections.MONGO_COLLECTIONS.VideoStreamGenerationQ.DeleteMany(context.Background(), bson.M{})
	if configs.EnvConfigs.APP_ENV != "development" {
		ffprobeCmd := exec.Command(
			"umount", "/home/sathyanitsme/cdn",
		)
		ffprobeCmd.Run()

		ffprobeCmd = exec.Command(
			"umount", "/home/sathyanitsme/storage",
		)
		ffprobeCmd.Run()

		ffprobeCmd = exec.Command(
			"systemctl", "daemon-reload",
		)
		ffprobeCmd.Run()

		ffprobeCmd = exec.Command(
			"systemctl", "restart", "local-fs.target",
		)
		ffprobeCmd.Run()
	}
	CRON_JOBS = cron.New()
	CRON_JOBS.AddFunc("*/15 * * * *", ClearExpiredToken)
	CRON_JOBS.AddFunc("*/1 * * * *", VideoStreamGenerationCron)
	CRON_JOBS.Start()
}
