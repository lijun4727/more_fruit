package rpc

import (
	"encoding/json"
	"io/ioutil"
	"morefruit/common"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderRoute struct {
	rabbitRpc common.RabbitRpc
}

func (or *OrderRoute) Route(e *gin.Engine) {
	r := e.Group("/order")
	//r.Use(apicom.VerifyTokenMiddleHandle)
	r.POST("/create", or.CreateOrder)
	or.rabbitRpc.Connect(common.RabbitmqHandleOrderAddress)
}

func (or *OrderRoute) CreateOrder(c *gin.Context) {
	data, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}
	var orderInfo common.OrderInfo
	err = json.Unmarshal(data, &orderInfo)
	if err != nil || orderInfo.Quantity == 0 || len(orderInfo.Order_number) == 0 ||
		orderInfo.Goods_id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "json invalid",
		})
		return
	}

	if err := or.rabbitRpc.ResetConnect(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	response, err := or.rabbitRpc.CallRpc(common.RabbitmqHandleOrderQuene, string(data),
		common.RabbitmqHandleOrderTimeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	resCode, err := strconv.Atoi(string(response))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}
	or.replayResult(c, resCode)
}

func (or *OrderRoute) replayResult(c *gin.Context, resCode int) {
	switch resCode {
	case common.SUCCESS:
		c.JSON(http.StatusOK, gin.H{
			"message": "SUCCESS",
		})
	case common.ORDER_GOODS_ID_NO_EXIST:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "ORDER_GOODS_NO_EXIST",
		})
	case common.ORDER_GOODS_QUANTITY_SAMLL:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "ORDER_GOODS_QUANTITY_SAMLL",
		})
	case common.ORDER_TIMEOUT:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "ORDER_TIMEOUT",
		})
	case common.JSON_INVALID:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "JSON_INVALID",
		})
	case common.ORDER_NUM_EXIST:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "ORDER_NUM_EXIST",
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "INTER_ERROR",
		})
	}
}

func (or *OrderRoute) Clean() {
	or.rabbitRpc.Close()
}
