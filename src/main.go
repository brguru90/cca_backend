package main

import (
	"cca/src/apis_set_1"
	"cca/src/configs"
	"cca/src/database"
	"cca/src/database/triggers"
	"cca/src/middlewares"
	"cca/src/my_modules"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"

	docs "cca/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

var SERVER_PORT string = "8000"

func main() {
	configs.InitEnv()
	my_modules.InitLogger()
	// database.ConnectPostgres()
	database.InitMongoDB()
	database.ConnectRedis()
	my_modules.InitFirebase()
	database.InitRedisPool()
	go triggers.TriggerForUsersModification()
	go my_modules.InitCronJobs()

	// init with default middlewares
	var all_router *gin.Engine = gin.Default()

	if configs.EnvConfigs.DISABLE_COLOR {
		gin.DisableConsoleColor()
	} else {
		gin.ForceConsoleColor()
	}

	if configs.EnvConfigs.GIN_MODE == "release" {
		// init without any middlewares
		all_router = gin.New()
		// but adding this
		all_router.Use(gin.Recovery())
	}

	// https://github.com/gin-gonic/gin

	// all_router = gin.New()
	// all_router.Use(static.Serve("/", static.LocalFile("./src/static", true)))
	all_router.Use(static.Serve("/", static.LocalFile("../frontend/build", true)))
	all_router.Use(static.Serve("/cdn", static.LocalFile("./uploads/public/video/multi_bitrate/", true)))
	if configs.EnvConfigs.GIN_MODE != "release" {
		all_router.Use(cors.Default())
	}
	docs.SwaggerInfo.BasePath = "/api"

	// !warning, the use of middleware may applicable to all further extended routes, so grouping will fix the issue, since middleware within the groups will not applicable to above routes from where its grouped

	{
		// just grouping, to make it more readable, to make middleware specific to groups
		api_router := all_router.Group("/api")
		api_router.Use(middlewares.FindUserAgentMiddleware()) // an example for global middleware on api_router
		api_router.Use(middlewares.HeaderHandlerFunc).GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "hi")
		})
		apis_set_1.InitApiTest(api_router) // more apis imported

		api_router.Use(func(c *gin.Context) {
			if c.Request.RequestURI == "/api/swagger" || c.Request.RequestURI == "/api/swagger/" {
				c.Redirect(http.StatusTemporaryRedirect, c.Request.RequestURI+"/index.html")
			} else {
				c.Next()
			}
		}).GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	}

	bind_to_host := fmt.Sprintf(":%d", configs.EnvConfigs.SERVER_PORT) //formatted host string
	// all_router.Run(bind_to_host)
	srv := &http.Server{
		Addr:         bind_to_host,
		Handler:      all_router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Debugf("http://127.0.0.1:%d", configs.EnvConfigs.SERVER_PORT)
	log.Debugf("http://127.0.0.1:%d/api/swagger", configs.EnvConfigs.SERVER_PORT)
	srv.ListenAndServe()
}
