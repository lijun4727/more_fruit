package main

//明天额外任务，配置redis删除策略和内存淘汰策略

import (
	"context"
	"fmt"
	"log"
	"morefruit/common"
	"morefruit/sever/sercom"
	"net"
	"strings"

	"github.com/go-redis/redis/v7"
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"google.golang.org/grpc"
)

type AccountLoginServer struct {
	common.AccountLoginServer
}

func (*AccountLoginServer) Login(_ context.Context, account *common.Account) (*common.ErrorCode, error) {
	password, err := redisSlaveClient.Get(account.UserName).Result()
	if err != nil && err != redis.Nil {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
	}
	if password == "" {
		has, err := x.Table(&sercom.AccountData{}).Where("username=?", account.UserName).
			Cols("password").Get(&password)
		if err != nil {
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
		}
		if !has {
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_USERNAME_NO_EXIST}, nil
		}

		redisRes := redisMasterClient.Set(account.UserName, password, common.RedisAccountTimeOut)
		if redisRes.Err() != nil {
			log.Printf("redisWriteClient.SetXX fail:UserName=%s,err=%v", account.UserName, redisRes.Err())
		}
	} else {
		redisMasterClient.Expire(account.UserName, common.RedisAccountTimeOut)
	}

	if account.Password != password {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_PASSWORD_ERROR}, nil
	}

	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}

func InitRedisServer() {
	redisServer := redis.NewClient(&redis.Options{
		Addr: common.RedisSlaveAddress,
	})
	err := redisServer.ConfigSet("requirepass", common.RedisSlavePassword).Err()
	if err != nil && !strings.Contains(err.Error(), "NOAUTH") {
		log.Fatalln("InitRedisServer RedisSlavePassword faild ")
	}
	redisServer.Close()
	redisServer = redis.NewClient(&redis.Options{
		Addr: common.RedisMasterAddress,
	})
	err = redisServer.ConfigSet("requirepass", common.RedisMasterPassword).Err()
	if err != nil && !strings.Contains(err.Error(), "NOAUTH") {
		log.Fatalln("InitRedisServer RedisMasterPassword faild ")
	}
	redisServer.Close()
}

var (
	x                 *xorm.Engine
	redisSlaveClient  *redis.Client
	redisMasterClient *redis.Client
)

func main() {
	var err error
	x, err = xorm.NewEngine("mysql", common.MysqlConnCmd)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, connect cmd: %s", err, common.MysqlConnCmd)
		return
	}
	defer x.Close()

	InitRedisServer()

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

	// name, err := redisSlaveClient.Get("name").Result()
	// if err == nil {

	// }
	// fmt.Println(name)
	// var password string
	// has, err := x.Table(&common.AccountData{}).Where("username=?", "lijun").
	// 	Cols("password").Get(&password)
	// if err != nil {

	// }
	// if has {
	// 	fmt.Println(password)
	// }

	lis, err := net.Listen("tcp", common.LoginRpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	} else {
		fmt.Printf("AccountLoginServer listen port:%s\n", common.AccountRpcPort)
	}
	s := grpc.NewServer()
	common.RegisterAccountLoginServer(s, &AccountLoginServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
