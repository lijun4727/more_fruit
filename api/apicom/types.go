package apicom

import "github.com/gin-gonic/gin"

type ApiRoute interface {
	Route(e *gin.Engine)
	Clean()
}
