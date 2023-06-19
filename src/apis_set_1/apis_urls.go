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
	// router.GET("all_users/", user_views.GetAllUserData)

	router.POST("register_build/", user_views.RegisterBuild)

	{
		protected_router := router.Group("user/", middlewares.ValidateToken(my_modules.AccessLevel.CUSTOMER))
		protected_router.GET("", my_modules.GetCachedResponse(user_views.GetUserData, "users", CACHE_TTL_DURATION, api_modules.ForUserPagination))
		protected_router.PUT("", user_views.UpdateUserData)
		protected_router.DELETE("", user_views.Deleteuser)
		protected_router.GET("active_sessions/", user_views.GetActiveSession)
		protected_router.POST("block_token/", user_views.BlockSession)
		protected_router.GET("logout/", user_views.Logout)

		// Todo, pending APIs

		protected_router.GET("get_playlists/", user_views.GetAvailablePlaylist)
		protected_router.POST("get_videos/", user_views.GetVideos)
		protected_router.POST("get_stream_key/", user_views.GetStreamKey)
		protected_router.GET("get_user_subscriptions/", user_views.GetUserPlaylistSubscriptionList)
		protected_router.GET("get_subscriptions/", user_views.GetAvailableSubscriptionPackages)
		protected_router.POST("enroll_to_course/", user_views.EnrollToCourse)
		protected_router.GET("confirm_payment_for_subscription/", user_views.PaymentConfirmationForSubscription)
		protected_router.GET("get_playlist_from_subscription/", user_views.GetPlaylistAvailableOnSubscription)

		protected_router.GET("study_materials_categories/", user_views.GetStudyMaterialsCategory)
		protected_router.GET("study_materials/", user_views.GetStudyMaterials)
		protected_router.POST("get_doc_key/", user_views.GetDocumentKey)
		protected_router.POST("enroll_to_study_material/", user_views.EnrollToStudyMaterial)
		protected_router.GET("confirm_payment_for_study_material_subscription/", user_views.PaymentConfirmationForSubscriptionForStudyMaterials)
		protected_router.GET("get_user_study_material_subscriptions/", user_views.GetUserStudyMaterialSubscriptionList)
	}

	{
		super_admin := router.Group("super_admin/", middlewares.ValidateToken(my_modules.AccessLevel.SUPER_ADMIN))
		super_admin.GET("user", super_admin_views.GetUsers)
		super_admin.POST("user", super_admin_views.AddAdminUsers)
		super_admin.PUT("user", super_admin_views.UpdateAdminUsersCredentials)
		super_admin.DELETE("user", super_admin_views.RemoveAdminUsers)
	}

	{
		admin_router := router.Group("admin/", middlewares.ValidateToken(my_modules.AccessLevel.ADMIN))
		admin_router.POST("upload_streaming_video/", admin_views.UploadVideo)
		admin_router.POST("generate_video_stream/", admin_views.GenerateVideoStream)
		admin_router.GET("upload_list/", admin_views.GetAllUploadedVideos)
		admin_router.DELETE("delete_streaming_video/", admin_views.RemoveVideos)

		admin_router.POST("upload_study_material/", admin_views.UploadStudyMaterials)
		admin_router.GET("doc_upload_list/", admin_views.GetAllUploadedStudyMaterials)
		admin_router.DELETE("delete_study_material/", admin_views.RemoveStudyMaterial)

		admin_router.GET("playlist/", admin_views.GetAllPlayLists)
		admin_router.POST("playlist/", admin_views.CreatePlayList)
		admin_router.PUT("playlist/", admin_views.UpdatePlayList)
		admin_router.DELETE("playlist/", admin_views.RemovePlayList)

		// lets use this in future to provide discounts/offer for playlist bundles
		admin_router.GET("subscription_package/", admin_views.GetAllSubscriptionPackages)
		admin_router.POST("subscription_package/", admin_views.CreateSubscriptionPackage)
		admin_router.PUT("subscription_package/", admin_views.UpdateSubscriptionPackage)
		admin_router.DELETE("subscription_package/", admin_views.RemoveSubscriptionPackage)

	}

}
