package main

import (
	"flag"
	"fmt"
	"morefruit/api/apicom"
	"morefruit/api/rpc"
	"morefruit/common"

	"github.com/fvbock/endless"
	"github.com/gin-gonic/gin"
)

func main() {
	var test = flag.String("test", "test_var", "test var ")
	flag.Parse()
	// token, _ := apicom.CreateToken("lijun", "123", common.TokenExpireTime)
	// fmt.Println(token)
	e := gin.Default()
	e.GET("/test", func(c *gin.Context) {
		testStr := fmt.Sprintf("test=%s", *test)
		c.String(200, testStr)
	})
	apicom.AppendRoute(e, &rpc.ShopRoute{})
	apicom.AppendRoute(e, &rpc.AccountRoute{})
	apicom.AppendRoute(e, &rpc.OrderRoute{})
	defer apicom.CleanAllRoute()
	endless.DefaultReadTimeOut = common.ApiReadTimeOut
	endless.DefaultWriteTimeOut = common.ApiWriteTimeOut
	endless.DefaultMaxHeaderBytes = 1 << 20
	//endless.ListenAndServeTLS(common.ApiPort, "server.crt", "server.key", e)
	endless.ListenAndServe(common.ApiPort, e)
}
