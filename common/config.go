package common

import "time"

const (
	TokenExpireTime            = 24 * time.Hour
	AccountRpcPort             = ":50051"
	LoginRpcPort               = ":50052"
	ShopManagePort             = ":50053"
	ImageMangePort             = ":50054"
	MysqlConnCmd               = "root:l@tcp(127.0.0.1:3306)/more_fruit?charset=utf8"
	AccountManageRpcAddress    = "127.0.0.1" + AccountRpcPort
	LoginRpcAddress            = "127.0.0.1" + LoginRpcPort
	ShopRpcAddress             = "127.0.0.1" + ShopManagePort
	ApiPort                    = ":8080"
	ApiReadTimeOut             = 30 * time.Second
	ApiWriteTimeOut            = 30 * time.Second
	RpcTimeout                 = 20 * time.Second
	RedisSlaveAddress          = "127.0.0.1:6379"
	RedisSlavePassword         = "l"
	RedisMasterAddress         = "127.0.0.1:6379"
	RedisMasterPassword        = "l"
	RedisReadTimeOut           = 10 * time.Second
	RedisWriteTimeOut          = 10 * time.Second
	RedisAccountTimeOut        = time.Hour * 24 * 7
	ShopNameMaxChar            = 10
	ShopDescMaxChar            = 50
	RabbitmqHandleOrderAddress = "amqp://lijun:l@192.168.1.6:5672/"
	RabbitmqHandleOrderQuene   = "create_order"
	RabbitmqHandleOrderTimeout = time.Minute
	MongoConnCmd               = "mongodb://lion:123456@192.168.1.6:27017"
	AccountManageSvrName       = "account_manage"
)
