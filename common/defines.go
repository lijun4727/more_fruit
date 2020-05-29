package common

/*
 *和api模块交互的server返回码，主要用于rabbitmq通讯方式，gorpc通讯方式返回码直接定义
 *business.proto文件中
 */

const (
	INTER_ERROR                = -1
	SUCCESS                    = 0
	ORDER_GOODS_QUANTITY_SAMLL = 1
	ORDER_GOODS_ID_NO_EXIST    = 2
	ORDER_TIMEOUT              = 3
	JSON_INVALID               = 4
	ORDER_NUM_EXIST            = 5
)
