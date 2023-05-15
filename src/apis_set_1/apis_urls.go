package apis_set_1

import (
	"cca/src/apis_set_1/api_modules"
	"cca/src/apis_set_1/views/admin_views"
	"cca/src/apis_set_1/views/super_admin_views"
	"cca/src/apis_set_1/views/user_views"
	"cca/src/configs"
	"cca/src/middlewares"
	"cca/src/my_modules"
	"time"

	"github.com/gin-gonic/gin"
)

const one_sec = 1000000000

var only_customer = []my_modules.AccessLevelType{my_modules.AccessLevel.CUSTOMER}
var all_users = []my_modules.AccessLevelType{my_modules.AccessLevel.SUPER_ADMIN, my_modules.AccessLevel.ADMIN, my_modules.AccessLevel.CUSTOMER}

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
	router.GET("login_status/", user_views.LoginStatus)
	router.GET("all_users/", user_views.GetAllUserData)

	{
		protected_router := router.Group("user/", middlewares.ValidateToken(my_modules.AccessLevel.CUSTOMER))
		protected_router.GET("", my_modules.GetCachedResponse(user_views.GetUserData, "users", CACHE_TTL_DURATION, api_modules.ForUserPagination))
		protected_router.PUT("", user_views.UpdateUserData)
		protected_router.DELETE("", user_views.Deleteuser)
		protected_router.GET("active_sessions/", user_views.GetActiveSession)
		protected_router.POST("block_token/", user_views.BlockSession)
		protected_router.GET("logout/", user_views.Logout)

		// Todo, pending APIs

		protected_router.GET("get_subscriptions/", user_views.GetAvailableSubscriptionPackages)
		protected_router.GET("enroll_to_subscription/", user_views.EnrollToSubscriptionPackage)
		protected_router.GET("confirm_payment_for_subscription/", user_views.PaymentConfirmationForSubscription)
		protected_router.GET("get_playlist_from_subscription/", user_views.GetPlaylistAvailableOnSubscription)
		protected_router.GET("get_videos_from_playlist/", user_views.GetPlaylistVideos)
	}

	{
		super_admin := router.Group("super_admin/", middlewares.ValidateToken(my_modules.AccessLevel.SUPER_ADMIN))
		super_admin.GET("", super_admin_views.GetUsers)
		super_admin.POST("", super_admin_views.AddAdminUsers)
		super_admin.PUT("", super_admin_views.UpdateAdminUsers)
		super_admin.DELETE("", super_admin_views.RemoveAdminUsers)
	}

	{
		admin_router := router.Group("admin/", middlewares.ValidateToken(my_modules.AccessLevel.ADMIN))
		admin_router.POST("upload_streaming_video/", admin_views.UploadVideo)
		admin_router.POST("generate_video_stream/", admin_views.GenerateVideoStream)
		admin_router.GET("upload_list/", admin_views.GetAllUploadedVideos)
		admin_router.POST("get_stream_key/", admin_views.GetStreamKey)

		// Todo, pending APIs
		admin_router.GET("playlist/", admin_views.GetAllPlayLists)
		admin_router.POST("playlist/", admin_views.CreatePlayList)
		admin_router.PUT("playlist/", admin_views.UpdatePlayList)
		admin_router.GET("subscription_package/", admin_views.CreateSubscriptionPackage)
		admin_router.POST("subscription_package/", admin_views.CreateSubscriptionPackage)
		admin_router.PUT("subscription_package/", admin_views.UpdateSubscriptionPackage)

	}

}
