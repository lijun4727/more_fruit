package rpc

import (
	"morefruit/api/apicom"
	"morefruit/common"
	"reflect"

	"github.com/gin-gonic/gin"
)

type AccountRoute struct {
	rpcAccountManageClient apicom.RpcClient
	rpcAccountLoginClient  apicom.RpcClient
}

func (sr *AccountRoute) CreateAccount(c *gin.Context) {
	sr.rpcAccountManageClient.Call(c, reflect.TypeOf(common.AccountInfo{}),
		reflect.ValueOf(common.NewAccountManageClient), "CreateAccount", true)
}

func (sr *AccountRoute) Login(c *gin.Context) {
	sr.rpcAccountLoginClient.Call(c, reflect.TypeOf(common.Account{}),
		reflect.ValueOf(common.NewAccountLoginClient),
		"Login", true)
}

func (sr *AccountRoute) Route(e *gin.Engine) {
	sr.rpcAccountManageClient.Adress = common.AccountManageRpcAddress
	sr.rpcAccountLoginClient.Adress = common.LoginRpcAddress

	r := e.Group("/account")
	//r.Use(AccountMiddle)
	r.POST("/register", sr.CreateAccount)
	r.POST("/login", sr.Login)
}

func (sr *AccountRoute) Clean() {
	sr.rpcAccountManageClient.Close()
	sr.rpcAccountLoginClient.Close()
}
