package apicom

import (
	"context"
	"morefruit/base/jwt"
	"morefruit/common"
	"net/http"
	"reflect"
	"sync"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

type RpcClient struct {
	Context *gin.Context
	Adress  string
	rpcCon  *grpc.ClientConn
	mutex   sync.Mutex
}

func (rc *RpcClient) connectServer() error {
	if rc.rpcCon == nil || rc.rpcCon.GetState() != connectivity.Ready {
		rc.mutex.Lock()
		if rc.rpcCon.GetState() == connectivity.Ready {
			rc.mutex.Unlock()
			return nil
		}

		if rc.rpcCon != nil {
			rc.rpcCon.Close()
		}

		var err error
		rc.rpcCon, err = grpc.Dial(rc.Adress, grpc.WithBlock(),
			grpc.WithTimeout(common.RpcTimeout), grpc.WithInsecure())
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
	rcpMethodName string, createToken bool) {
	rc.Context = c
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
