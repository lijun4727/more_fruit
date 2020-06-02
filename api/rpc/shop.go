package rpc

import (
	"fmt"
	"morefruit/api/apicom"
	"morefruit/common"
	"os"
	"reflect"
	"sync"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/gin-gonic/gin"
	"github.com/streadway/amqp"
)

type ShopRoute struct {
	rpcClient apicom.RpcClient
	amqpConn  *amqp.Connection
	amqpCh    *amqp.Channel
	mutex     sync.Mutex
}

func (sr *ShopRoute) Build(c *gin.Context) {
	sr.rpcClient.Call(c, reflect.TypeOf(common.Shop{}),
		reflect.ValueOf(common.NewShopManageClient), "Build", false, "")
}

func (sr *ShopRoute) AddImage(c *gin.Context) {
	// 创建OSSClient实例。
	client, err := oss.New("<yourEndpoint>", "<yourAccessKeyId>", "<yourAccessKeySecret>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 获取存储空间。
	bucket, err := client.Bucket("<yourBucketName>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}

	// 读取本地文件。
	fd, err := os.Open("<yourLocalFile>")
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
	defer fd.Close()

	// 上传文件流。
	err = bucket.PutObject("<yourObjectName>", fd)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(-1)
	}
}

func (sr *ShopRoute) AddGoods(c *gin.Context) {
	sr.rpcClient.Call(c, reflect.TypeOf(common.Goodses{}),
		reflect.ValueOf(common.NewShopManageClient), "AddGoods", false, "")
}

func (sr *ShopRoute) Route(e *gin.Engine) {
	sr.rpcClient.Adress = common.ShopRpcAddress
	r := e.Group("/shop")
	r.Use(apicom.VerifyTokenMiddleHandle)
	r.POST("/build", sr.Build)
	r.POST("/addimage", sr.AddImage)
	r.POST("/addgoods", sr.AddGoods)
}

func (sr *ShopRoute) Clean() {
	sr.rpcClient.Close()
}
