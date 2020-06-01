package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"morefruit/base/balance"
	"morefruit/common"
	"morefruit/sever/sercom"
	"net"
	"strconv"

	"regexp"
	"unicode/utf8"

	"github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	ziu_mysql "github.com/ziutek/mymysql/mysql"
	"google.golang.org/grpc"
)

var x *xorm.Engine

var (
	nodeID        = flag.String("node", "node_0", "rpc cluster node ID")
	consulAddr    = flag.String("consulAddr", "http://127.0.0.1:8500", "consul agent node address")
	consulSrvName = flag.String("consulSrvName", "account_manage", "consul agent server name")
)

type AccountManageServer struct {
	common.AccountManageServer
}

func IsIdentityValid(idefine string) bool {
	var isIdentity bool
	// 15位身份证号码：15位全是数字
	if isIdentity, _ = regexp.MatchString(`^(\d{15})$`, idefine); isIdentity {
		return true
	}
	// 18位身份证：前17位为数字，第18位为校验位，可能是数字或X
	if isIdefine, _ := regexp.MatchString(`^(\d{17})([0-9]|X)$`, idefine); isIdefine {
		return true
	}
	return false
}

func (*AccountManageServer) CreateAccount(c context.Context, accountInfo *common.AccountInfo) (*common.ErrorCode, error) {
	if userNameLen := utf8.RuneCountInString(accountInfo.UserName); userNameLen > 20 || userNameLen < 5 {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_USERNAME_INVALID}, nil
	}
	if passwdLen := utf8.RuneCountInString(accountInfo.Password); passwdLen > 20 || passwdLen < 5 {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_PASSWORD_INVALID}, nil
	}
	if isDefine := IsIdentityValid(accountInfo.Identity); !isDefine {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_IDENTITY_INVALID}, nil
	}
	if isNum, _ := regexp.MatchString("^[0-9]*$", accountInfo.Phone); !isNum {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_PHONE_INVALID}, nil
	}

	account := sercom.AccountData{
		UserName: accountInfo.UserName,
		Password: accountInfo.Password,
		Identity: accountInfo.Identity,
		Phone:    accountInfo.Phone,
	}

	_, err := x.Insert(account)

	if err != nil {
		var mySqlErr *mysql.MySQLError
		if errors.As(err, &mySqlErr) && mySqlErr.Number == ziu_mysql.ER_DUP_ENTRY {
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_USERNAME_EXIST}, nil
		}
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
	}

	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}

func registerToConsul() (*balance.ConsulBalance, error) {
	port, _ := strconv.Atoi(common.AccountRpcPort[1:])

	config := balance.Config{
		ConsulAddr:    *consulAddr,
		ConsulSrvName: *consulSrvName,
		NodeID:        *nodeID,
		Port:          port,
		Weight:        "1",
		Ttl:           5,
	}
	cb := balance.ConsulBalance{}
	err := cb.RegisterServerIntoConsul(&config)

	return &cb, err
}

func main() {
	var err error
	cb, err := registerToConsul()
	if err != nil {
		log.Fatalf("failed to registerToConsul: %v", err)
		return
	}
	defer cb.UnRegister()

	x, err = xorm.NewEngine("mysql", common.MysqlConnCmd)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, connect cmd: %s", err, common.MysqlConnCmd)
		return
	}
	defer x.Close()

	if isTableExit, _ := x.IsTableExist(&sercom.AccountData{}); !isTableExit {
		err := x.Sync2(&sercom.AccountData{})
		if err != nil {
			log.Fatalf("failed to create table AccountData: %v", err)
			return
		}
	}

	lis, err := net.Listen("tcp", common.AccountRpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	} else {
		fmt.Printf("AccountManageServer listen port:%s\n", common.AccountRpcPort)
	}
	s := grpc.NewServer()
	common.RegisterAccountManageServer(s, &AccountManageServer{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
