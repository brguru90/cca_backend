package apis_set_1

import (
	"time"
	"travel_planner/src/apis_set_1/api_modules"
	"travel_planner/src/apis_set_1/views/admin_views"
	"travel_planner/src/apis_set_1/views/user_views"
	"travel_planner/src/configs"
	"travel_planner/src/middlewares"
	"travel_planner/src/my_modules"

	"github.com/gin-gonic/gin"
)

const one_sec = 1000000000

// only the functions whose initial letter is upper case only those can be exportable from package
func InitApiTest(router *gin.RouterGroup) {
	var CACHE_TTL_DURATION = time.Duration(one_sec * configs.EnvConfigs.RESPONSE_CACHE_TTL_IN_SECS)

	router.Use(middlewares.ApiSpecificMiddleware())

	router.GET("test/:id", test_api)
	router.GET("health_check", health_check)
	router.POST("sign_up/", user_views.SignUp)
	router.POST("login/", user_views.Login)
	router.POST("login_mobile/", user_views.LoginWithMobile)
	router.POST("verify_social_auth/", user_views.VerifySocialAuth)
	router.POST("login_status/", user_views.LoginStatus)
	router.GET("all_users/", user_views.GetAllUserData)

	{
		protected_router := router.Group("user/", middlewares.ValidateToken())
		protected_router.GET("", my_modules.GetCachedResponse(user_views.GetUserData, "users", CACHE_TTL_DURATION, api_modules.ForUserPagination))
		protected_router.PUT("", user_views.UpdateUserData)
		protected_router.DELETE("", user_views.Deleteuser)
		protected_router.GET("active_sessions/", user_views.GetActiveSession)
		protected_router.POST("block_token/", user_views.BlockSession)
		protected_router.GET("logout/", user_views.Logout)
	}

	{
		admin_router := router.Group("admin/")
		admin_router.POST("upload_streaming_video/", admin_views.UploadVideo)
	}

}
