package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"morefruit/base/rabrpc"
	"morefruit/common"
	"morefruit/sever/sercom"
	"time"

	"github.com/go-redis/redis"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

//mongo go官方库有bug，用户名和密码长度必须大于1,bson.D类型导致函数出bug，用bson.M代替

var (
	x                 *xorm.Engine
	redisMasterClient *redis.Client
	redisSlaveClient  *redis.Client
	mongoClient       *mongo.Client
)

var (
	OrderExistError = errors.New("order num have existed")
)

func updateRedisGoods(orderInfo *common.OrderInfo) int {
	var keys []string
	keys = append(keys, fmt.Sprintf("goods_info_%d", orderInfo.Goods_id))
	luaScript, _ := ioutil.ReadFile("update_goods_online.lua")
	script := redis.NewScript(string(luaScript))
	if script == nil {
		return common.INTER_ERROR
	}

	resCode, err := script.Run(redisMasterClient, keys, orderInfo.Quantity).Result()
	if err != nil {
		return common.INTER_ERROR
	}
	if val, ok := resCode.(int64); ok && val < 0 {
		return common.ORDER_GOODS_QUANTITY_SAMLL
	} else {
		return common.SUCCESS
	}
}

func updateMysqlGoods(orderInfo *common.OrderInfo) int {
	var cur_qiantity int
	sqlCmd := fmt.Sprintf("call reponse_order_proc(%d,%d)", orderInfo.Goods_id, orderInfo.Quantity)
	_, err := x.SQL(sqlCmd).Get(&cur_qiantity)
	if err != nil {
		return common.INTER_ERROR
	}

	if cur_qiantity >= 0 {
		return common.SUCCESS
	}

	var resCode int
	switch cur_qiantity {
	case -1:
		resCode = common.ORDER_GOODS_QUANTITY_SAMLL
	case -2:
		resCode = common.ORDER_GOODS_ID_NO_EXIST
	default:
		resCode = common.INTER_ERROR
	}
	return resCode
}

func VerifyOrder(orderInfo *common.OrderInfo) int {
	coll := mongoClient.Database(sercom.MongoOrderDBName).Collection(sercom.MongoOrderColName)
	var result bson.M
	orderData := bson.M{"order_num": orderInfo.Order_number}
	err := coll.FindOne(context.TODO(), orderData).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return common.SUCCESS
		}
		return common.INTER_ERROR
	}
	return common.ORDER_NUM_EXIST
}

func createOrderInMongo(orderInfo *common.OrderInfo) int {
	coll := mongoClient.Database(sercom.MongoOrderDBName).Collection(sercom.MongoOrderColName)
	orderData := bson.M{
		"order_num": orderInfo.Order_number,
		"goods_id":  orderInfo.Goods_id,
		"quantity":  orderInfo.Quantity,
	}
	_, err := coll.InsertOne(context.TODO(), orderData)

	if err != nil {
		switch err.(type) {
		case mongo.WriteException:
			wrEx := err.(mongo.WriteException)
			for _, wrEr := range wrEx.WriteErrors {
				if wrEr.Code == 11000 {
					return common.ORDER_NUM_EXIST
				}
			}
			return common.INTER_ERROR
		default:
			return common.INTER_ERROR
		}
	}
	return common.SUCCESS
}

func callback(context *context.Context, jsonParam string) string {
	var orderInfo common.OrderInfo
	err := json.Unmarshal([]byte(jsonParam), &orderInfo)
	if err != nil {
		return fmt.Sprintf("%d", common.JSON_INVALID)
	}

	resCode := VerifyOrder(&orderInfo)
	if resCode != common.SUCCESS {
		return fmt.Sprintf("%d", resCode)
	}

	resCode = createOrderInMongo(&orderInfo)
	if resCode != common.SUCCESS {
		return fmt.Sprintf("%d", resCode)
	}

	resCode = updateRedisGoods(&orderInfo)
	if resCode != common.SUCCESS {
		return fmt.Sprintf("%d", resCode)
	}
	resCode = updateMysqlGoods(&orderInfo)

	//以上出错时，根据业务是否支持数据回滚

	return fmt.Sprintf("%d", resCode)
}

func initMongoOrder() error {
	dbOrder := mongoClient.Database(sercom.MongoOrderDBName)
	if dbOrder == nil {
		return nil
	}

	coll := dbOrder.Collection(sercom.MongoOrderColName)
	models := []mongo.IndexModel{
		{
			Keys:    bson.M{"order_num": 1},
			Options: options.Index().SetName("testIndex").SetUnique(true),
		},
	}

	// Specify the MaxTime option to limit the amount of time the operation can run on the server
	opts := options.CreateIndexes().SetMaxTime(2 * time.Second)
	_, err := coll.Indexes().CreateMany(context.TODO(), models, opts)
	if err != nil {
		log.Fatal(err)
	}
	return err
}

//执行该模块前先在mysql中执行response_order_proc.sql文件，程序中执行时显示语法错误
func main() {
	var err error
	mongoClient, err = mongo.NewClient(options.Client().ApplyURI(common.MongoConnCmd))
	if err != nil {
		log.Fatalf("fail to mongo NewClient: %v, connect cmd: %s", err, common.MongoConnCmd)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = mongoClient.Connect(ctx)
	if err != nil {
		log.Fatalf("fail to connect mongo: %v, connect cmd: %s", err, common.MongoConnCmd)
		return
	}
	//err = mongoClient.Ping(ctx, nil)
	defer mongoClient.Disconnect(ctx)

	if err = initMongoOrder(); err != nil {
		log.Fatalf("fail to initMongoOrder: %v", err)
		return
	}

	// //var client *mongo.Client // assume client is configured with write concern majority and read preference primary

	// // Specify the DefaultReadConcern option so any transactions started through the session will have read concern
	// // majority.
	// // The DefaultReadPreference and DefaultWriteConcern options aren't specified so they will be inheritied from client
	// // and be set to primary and majority, respectively.
	// opts := options.Session().SetDefaultReadConcern(readconcern.Majority())
	// sess, err := mongoClient.StartSession(opts)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer sess.EndSession(context.TODO())

	// // Specify the ReadPreference option to set the read preference to primary preferred for this transaction.
	// txnOpts := options.Transaction().SetReadPreference(readpref.PrimaryPreferred())
	// result, err := sess.WithTransaction(context.TODO(), func(sessCtx mongo.SessionContext) (interface{}, error) {
	// 	// Use sessCtx as the Context parameter for InsertOne and FindOne so both operations are run in a
	// 	// transaction.

	// 	coll := mongoClient.Database(sercom.MongoOrderDBName).Collection(sercom.MongoOrderColName)
	// 	orderData := bson.M{"order_num": "abdc1"}
	// 	var result bson.M
	// 	err := coll.FindOne(sessCtx, orderData).Decode(&result)
	// 	fmt.Println(err)
	// 	if err != nil {
	// 		_, err = coll.InsertOne(sessCtx, orderData)
	// 	}
	// 	return result, err
	// }, txnOpts)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("result: %v\n", result)

	x, err = xorm.NewEngine("mysql", common.MysqlConnCmd)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, connect cmd: %s", err, common.MysqlConnCmd)
		return
	}
	defer x.Close()

	// sqlScript, err := ioutil.ReadFile("response_order_proc.sql")
	// if err != nil {
	// 	return
	// }
	// fmt.Println(string(sqlScript))
	// var cur_qiantity int
	// _, err = x.SQL(string(sqlScript)).Get(&cur_qiantity)
	// _, err = x.ImportFile("response_order_proc.sql")

	redisSlaveClient = redis.NewClient(&redis.Options{
		Addr:         common.RedisSlaveAddress,
		Password:     common.RedisSlavePassword,
		ReadTimeout:  common.RedisReadTimeOut,
		WriteTimeout: common.RedisWriteTimeOut,
	})
	defer redisSlaveClient.Close()

	redisMasterClient = redis.NewClient(&redis.Options{
		Addr:         common.RedisMasterAddress,
		Password:     common.RedisMasterPassword,
		ReadTimeout:  common.RedisReadTimeOut,
		WriteTimeout: common.RedisWriteTimeOut,
	})
	defer redisMasterClient.Close()

	var rabRpc rabrpc.RabbitRpc
	err = rabRpc.Connect(common.RabbitmqHandleOrderAddress)
	if err != nil {
		log.Fatalf("rabRpc connect failed:%v", err)
		return
	}
	err = rabRpc.ResponceRpc(common.RabbitmqHandleOrderQuene, callback)
	if err != nil {
		log.Fatalf("common.RabbitmqHandleOrderQuene failed:%v", err)
		return
	}
}
