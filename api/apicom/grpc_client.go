package apicom

import (
	"context"
	"morefruit/base/jwt"
	"morefruit/common"
	"morefruit/third_party/grpc-lb/balancer"
	"morefruit/third_party/grpc-lb/registry/consul"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	con_api "github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type RpcClient struct {
	ServerName string
	Context    *gin.Context
	Adress     string
	rpcCon     *grpc.ClientConn
	mutex      sync.Mutex
}

func (rc *RpcClient) connectServer() error {
	if rc.rpcCon == nil || rc.rpcCon.GetState() != connectivity.Ready {
		rc.mutex.Lock()
		if rc.rpcCon != nil {
			rc.rpcCon.Close()
		}

		var err error
		if len(rc.ServerName) == 0 {
			rc.rpcCon, err = grpc.Dial(rc.Adress, grpc.WithBlock(),
				grpc.WithTimeout(common.RpcTimeout), grpc.WithInsecure())

		} else {
			consul.RegisterResolver("consul", &con_api.Config{Address: "http://127.0.0.1:8500"},
				rc.ServerName)
			rc.rpcCon, err = grpc.Dial(
				"consul:///",
				grpc.WithBlock(),
				grpc.WithTimeout(common.RpcTimeout),
				grpc.WithInsecure(),
				grpc.WithBalancerName(balancer.RoundRobin))
		}

		rc.mutex.Unlock()
		if err != nil {
			rc.Context.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return err
		}
	}
	return nil
}

func (rc *RpcClient) Call(c *gin.Context, rpcDataType reflect.Type, newClientFun reflect.Value,
	rcpMethodName string, createToken bool, serverName string) {
	rc.Context = c
	rc.ServerName = serverName
	if rc.connectServer() != nil {
		return
	}

	rcpData := reflect.New(rpcDataType).Interface()
	err := rc.Context.BindJSON(rcpData)
	if err != nil {
		rc.Context.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	newClienFunParam := []reflect.Value{reflect.ValueOf(rc.rpcCon)}
	retList := newClientFun.Call(newClienFunParam)

	ctx, cancel := context.WithTimeout(context.Background(), common.RpcTimeout)
	defer cancel()
	paramList := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(rcpData)}
	replayList := retList[0].MethodByName(rcpMethodName).Call(paramList)

	if rep := replayList[1].Interface(); rep != nil {
		err = rep.(error)
	} else {
		err = nil
	}
	err_code := replayList[0].Elem().Addr().Interface().(*common.ErrorCode)
	rc.replayResult(err, err_code, createToken)
}

func (rc *RpcClient) Close() {
	if rc.rpcCon != nil {
		rc.rpcCon.Close()
	}
}

func (rc *RpcClient) replayResult(err error, errCode *common.ErrorCode, createToken bool) {
	if err != nil {
		rc.Context.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	errMsg := common.ERROR_CODE_name[(int32)(errCode.ErrCode)]
	if errCode.ErrCode == common.ERROR_CODE_NONE {
		if createToken {
			token, err := jwt.CreateToken("lijun", "192.168.1.1", common.TokenExpireTime)
			if err != nil {
				rc.Context.JSON(http.StatusInternalServerError, gin.H{
					"message": err.Error(),
				})
			} else {
				rc.Context.JSON(http.StatusOK, gin.H{
					"message": errMsg,
					"token":   token,
				})
			}
		} else {
			rc.Context.JSON(http.StatusOK, gin.H{
				"message": errMsg,
			})
		}
	} else {
		rc.Context.JSON(http.StatusBadRequest, gin.H{
			"message": errMsg,
		})
	}
}
