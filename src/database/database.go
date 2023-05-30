package database

import (
	"cca/src/database/database_connections"
	"cca/src/database/mongo_modals"
)

func InitDataBases() {
	database_connections.InitMongoDB()
	database_connections.ConnectRedis()
	database_connections.InitRedisPool()

	mongo_modals.InitUserCollection()
	mongo_modals.InitActiveSessionCollection()
	mongo_modals.InitAppBuildRegistrationCollection()
	mongo_modals.InitVideoUploadCollection()
	mongo_modals.InitVideoPlayListCollection()
	mongo_modals.InitVideoPlayListSubscriptionPackageCollection()
	mongo_modals.InitVideoPlayListUserSubscriptionCollection()
	mongo_modals.InitPaymentOrderCollection()
	mongo_modals.InitStudyMaterialCollection()
	mongo_modals.InitStudyMaterialUserSubscriptionCollection()
	mongo_modals.InitStudyMaterialCategoryModalCollection()
}
