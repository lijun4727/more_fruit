package apicom

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"morefruit/base/jwt"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	AllRoute [](ApiRoute)
)

func AppendRoute(e *gin.Engine, apiRoute ApiRoute) {
	AllRoute = append(AllRoute, apiRoute)
	apiRoute.Route(e)
}

func CleanAllRoute() {
	for _, r := range AllRoute {
		r.Clean()
	}
}

func VerifyTokenMiddleHandle(c *gin.Context) {
	data, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	type Token struct {
		Data string `json:"token"`
	}
	var token Token
	json.Unmarshal([]byte(data), &token)
	_, err = jwt.VerifyToken(token.Data)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "token invalid," + err.Error(),
		})
	}
}
