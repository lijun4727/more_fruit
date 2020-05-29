package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"morefruit/common"
	"morefruit/sever/sercom"
	"net"
	"unicode/utf8"

	"github.com/go-redis/redis"
	"github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	ziu_mysql "github.com/ziutek/mymysql/mysql"

	"google.golang.org/grpc"
)

var (
	x                 *xorm.Engine
	goods_view_name   = "goods_all_info"
	redisMasterClient *redis.Client
)

type ShopManageService struct {
	common.ShopManageServer
}

func (*ShopManageService) Build(_ context.Context, shop *common.Shop) (*common.ErrorCode, error) {
	if nameLen := utf8.RuneCountInString(shop.Name); nameLen <= 0 || nameLen > 10 {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_SHOP_NAME_INVALID}, nil
	}
	if utf8.RuneCountInString(shop.Desc) > common.ShopDescMaxChar {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_SHOP_DESC_OVER_LEN}, nil
	}

	shopInfo := sercom.ShopInfo{
		Name:      shop.Name,
		Desc:      shop.Desc,
		AccountId: shop.AccountId,
	}
	_, err := x.Insert(shopInfo)
	if err != nil {
		var mySqlErr *mysql.MySQLError
		if errors.As(err, &mySqlErr) && mySqlErr.Number == ziu_mysql.ER_NO_REFERENCED_ROW_2 {
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_ACCOUNT_ID_NO_EXIST}, nil
		}
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
	}

	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}
func (*ShopManageService) SetLogo(context.Context, *common.Logo) (*common.ErrorCode, error) {
	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}

func (sms *ShopManageService) AddImage(_ context.Context, shopImg *common.ShopImages) (*common.ErrorCode, error) {
	session := x.NewSession()
	defer session.Close()
	session.Begin()
	for _, path := range shopImg.ImgPaths {
		imgInfo := sercom.ImageInfoInShop{
			Path: path,
		}
		_, err := session.Insert(&imgInfo)
		if err != nil {
			var mySqlErr *mysql.MySQLError
			if errors.As(err, &mySqlErr) && mySqlErr.Number == ziu_mysql.ER_DUP_ENTRY {
				return &common.ErrorCode{ErrCode: common.ERROR_CODE_IMAGE_EXIST}, nil
			}
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
		}

		shopToImg := sercom.ShopImage{
			ShopId:  shopImg.ShopId,
			ImageId: imgInfo.Id,
		}
		_, err = session.Insert(shopToImg)
		if err != nil {
			var mySqlErr *mysql.MySQLError
			if errors.As(err, &mySqlErr) && mySqlErr.Number == ziu_mysql.ER_NO_REFERENCED_ROW_2 {
				return &common.ErrorCode{ErrCode: common.ERROR_CODE_SHOP_ID_NO_EXIST}, nil
			}
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
		}
	}
	err := session.Commit()
	if err != nil {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
	}
	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}

func (sms *ShopManageService) AddGoods(_ context.Context, goodses *common.Goodses) (*common.ErrorCode, error) {
	session := x.NewSession()
	defer session.Close()
	session.Begin()
	for _, goods := range goodses.Goodses {
		goodsInfo := sercom.GoodsInfo{
			Name:     goods.Name,
			Desc:     goods.Desc,
			Quantity: goods.Quantity,
		}
		_, err := session.Insert(&goodsInfo)
		if err != nil {
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
		}
		shopGoods := sercom.ShopGoods{
			ShopId:  goods.ShopId,
			GoodsId: goodsInfo.Id,
		}
		_, err = session.Insert(shopGoods)
		if err != nil {
			var mySqlErr *mysql.MySQLError
			if errors.As(err, &mySqlErr) && mySqlErr.Number == ziu_mysql.ER_NO_REFERENCED_ROW_2 {
				return &common.ErrorCode{ErrCode: common.ERROR_CODE_SHOP_ID_NO_EXIST}, nil
			}
			return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
		}
	}
	err := session.Commit()
	if err != nil {
		return &common.ErrorCode{ErrCode: common.ERROR_CODE_INTERNAL_ERROR}, err
	}
	return &common.ErrorCode{ErrCode: common.ERROR_CODE_NONE}, nil
}

func ResetGoodsOnlineWithLua(goodsNum int) error {
	var goodAllInfo []sercom.GoodsAllInfo
	err := x.Table(goods_view_name).Desc("goods_id").Limit((int)(goodsNum)).Find(&goodAllInfo)
	if err != nil {
		return err
	}

	pipe := redisMasterClient.Pipeline()
	defer pipe.Close()
	var keys []string
	var args []interface{}
	for _, goods := range goodAllInfo {
		keys = append(keys, fmt.Sprintf("goods_info_%d", goods.Goods_id))
		args = append(args, "shop_id")
		args = append(args, goods.Shop_id)
		args = append(args, "name")
		args = append(args, goods.Name)
		args = append(args, "desc")
		args = append(args, goods.Desc)
		args = append(args, "quantity")
		args = append(args, goods.Quantity)
	}

	luaData, err := ioutil.ReadFile("reset_goods_online.lua")
	if err != nil {
		return err
	}

	script := redis.NewScript(string(luaData))
	_, err = script.Run(redisMasterClient, keys, args...).Result()
	return err
}

// func ResetGoodsOnline(goodsNum int) error {
// 	var goodAllInfo []sercom.GoodsAllInfo
// 	err := x.Table(goods_view_name).Desc("goods_id").Limit((int)(goodsNum)).Find(&goodAllInfo)
// 	if err != nil {
// 		return err
// 	}

// 	pipe := redisMasterClient.Pipeline()
// 	defer pipe.Close()
// 	pipe.FlushAll()

// 	for _, goods := range goodAllInfo {
// 		key := fmt.Sprintf("goods_info_%d", goods.Goods_id)
// 		pipe.HSet(key, "goods_id", goods.Goods_id)
// 		pipe.HSet(key, "shop_id", goods.Shop_id)
// 		pipe.HSet(key, "name", goods.Name)
// 		pipe.HSet(key, "desc", goods.Desc)
// 		pipe.HSet(key, "quantity", goods.Quantity)
// 	}
// 	_, err = pipe.Exec()
// 	return err
// }

func InitShopDatabase() error {
	session := x.NewSession()
	defer session.Close()
	session.Begin()
	if isTableExit, _ := x.IsTableExist(&sercom.ShopInfo{}); !isTableExit {

		err := session.Sync2(&sercom.ShopInfo{})
		if err != nil {
			log.Fatalf("failed to create table ShopInfo: %v", err)
			return err
		}

		shopName := x.TableName(&sercom.ShopInfo{})
		sqlStr := fmt.Sprintf(`alter table %s add constraint FK_shopinfo_accountid foreign key (account_id) 
		references account_data(id) ON DELETE CASCADE ON UPDATE CASCADE`, shopName)
		_, err = session.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shopinfo_accountid: %v", err)
			return err
		}
	}
	err := session.Commit()
	return err
}

func InitShopImgDatabase() error {
	if isTableExit, _ := x.IsTableExist(&sercom.ImageInfoInShop{}); !isTableExit {
		err := x.Sync2(&sercom.ImageInfoInShop{})
		if err != nil {
			log.Fatalf("failed to create table ImageInfo: %v", err)
			return err
		}
	}

	session := x.NewSession()
	defer session.Close()
	session.Begin()
	if isTableExit, _ := x.IsTableExist(&sercom.ShopImage{}); !isTableExit {
		err := session.Sync2(&sercom.ShopImage{})
		if err != nil {
			log.Fatalf("failed to create table ShopImage: %v", err)
			return err
		}

		tableName := x.TableName(&sercom.ShopImage{})
		sqlStr := fmt.Sprintf(`alter table %s add constraint FK_shopimage_imageid foreign key (image_id) 
		references image_info_in_shop(id) ON DELETE CASCADE ON UPDATE CASCADE`, tableName)
		_, err = session.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shop_image_id: %v", err)
			return err
		}
		sqlStr = fmt.Sprintf(`alter table %s add constraint FK_shopimage_shopid foreign key (shop_id) 
		references shop_info(id) ON DELETE CASCADE ON UPDATE CASCADE`, tableName)
		_, err = session.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shop_id: %v", err)
			return err
		}
	}
	err := session.Commit()
	return err
}

func InitShopGoods() error {
	if isTableExit, _ := x.IsTableExist(&sercom.GoodsInfo{}); !isTableExit {
		err := x.Sync2(&sercom.GoodsInfo{})
		if err != nil {
			log.Fatalf("failed to create table ImageInfo: %v", err)
			return err
		}
	}

	session := x.NewSession()
	defer session.Close()
	session.Begin()
	if isTableExit, _ := x.IsTableExist(&sercom.ShopGoods{}); !isTableExit {
		err := session.Sync2(&sercom.ShopGoods{})
		if err != nil {
			log.Fatalf("failed to create table ShopImage: %v", err)
			return err
		}

		tableName := x.TableName(&sercom.ShopGoods{})
		sqlStr := fmt.Sprintf(`alter table %s add constraint FK_shopgoods_goods_id foreign key (goods_id) 
		references goods_info(id) ON DELETE CASCADE ON UPDATE CASCADE`, tableName)
		_, err = session.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shop_image_id: %v", err)
			return err
		}
		sqlStr = fmt.Sprintf(`alter table %s add constraint FK_shopgoods_sho_id foreign key (shop_id) 
		references shop_info(id) ON DELETE CASCADE ON UPDATE CASCADE`, tableName)
		_, err = session.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shop_id: %v", err)
			return err
		}
	}
	err := session.Commit()
	return err
}

func InitGoodsAllinfoDatabase() error {
	if isTableExit, _ := x.IsTableExist("goods_all_info"); !isTableExit {
		shopGoods := x.TableName(&sercom.ShopGoods{})
		goods_info := x.TableName(&sercom.GoodsInfo{})
		sqlStr := fmt.Sprintf(`create or replace view %s as 
		select sg.shop_id,sg.goods_id,gi.name,gi.desc,gi.quantity
		from %s as sg,%s as gi where sg.goods_id = gi.id`,
			goods_view_name, shopGoods, goods_info)
		_, err := x.Exec(sqlStr)
		if err != nil {
			log.Fatalf("failed to create foreign key FK_shop_id: %v", err)
			return err
		}
	}
	return nil
}

func InitDatabase() bool {
	if InitShopDatabase() != nil || InitShopImgDatabase() != nil || InitShopGoods() != nil ||
		InitGoodsAllinfoDatabase() != nil {
		return false
	}
	return true
}

func main() {
	var err error
	x, err = xorm.NewEngine("mysql", common.MysqlConnCmd)
	if err != nil {
		log.Fatalf("failed to connect mysql: %v, connect cmd: %s", err, common.MysqlConnCmd)
		return
	}
	defer x.Close()

	if !InitDatabase() {
		return
	}

	redisMasterClient = redis.NewClient(&redis.Options{
		Addr:         common.RedisMasterAddress,
		Password:     common.RedisMasterPassword,
		ReadTimeout:  common.RedisReadTimeOut,
		WriteTimeout: common.RedisWriteTimeOut,
	})
	defer redisMasterClient.Close()

	ResetGoodsOnlineWithLua(10)

	lis, err := net.Listen("tcp", common.ShopManagePort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return
	} else {
		fmt.Printf("ShopManageService listen port:%s\n", common.ShopManagePort)
	}
	s := grpc.NewServer()
	common.RegisterShopManageServer(s, &ShopManageService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
