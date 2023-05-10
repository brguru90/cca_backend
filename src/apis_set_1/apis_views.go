package apis_set_1

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func test_api(c *gin.Context) {
	c.String(http.StatusOK, "the param sent %s", c.Param("id"))
}

func health_check(c *gin.Context) {
	// Todo: need to check DB connectivity, both mongo & redis
	c.Status(http.StatusOK)
}
