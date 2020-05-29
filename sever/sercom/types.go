package sercom

type AccountData struct {
	Id          int64
	UserName    string `xorm:"unique notnull VARCHAR(20)"`
	Password    string `xorm:"notnull VARCHAR(20)"`
	Identity    string `xorm:"index notnull CHAR(18)"`
	Phone       string `xorm:"notnull CHAR(11)"`
	LogoImageId int64  //外键=>ImageInfo.Id
	Email       string
	Name        string
}

type ShopInfo struct {
	Id          int64
	Name        string `xorm:"notnull VARCHAR(10)"`
	Desc        string `xorm:"VARCHAR(50)"`
	AccountId   int64  `xorm:"notnull BIGINT(11)"` //外键=>AccountData.Id
	LogoImageId int64  //外键=>ImageInfo.Id
	Asset       int
}

type ImageInfoInShop struct {
	Id   int64  `xorm:"pk notnull autoincr BIGINT(11)"`
	Path string `xorm:"unique notnull VARCHAR(200)"`
}

type ShopImage struct {
	ShopId  int64 `xorm:"pk notnull BIGINT(11)"` //外键=>ShopInfo.Id
	ImageId int64 `xorm:"pk notnull BIGINT(11)"` //外键=>ImageInfo.Id
}

type AccountImage struct {
	AccountId int64 `xorm:"pk notnull BIGINT(11)"` //外键=>AccountData.Id
	ImageId   int64 `xorm:"pk notnull BIGINT(11)"` //外键=>ImageInfo.Id
}

type GoodsInfo struct {
	Id       int64
	Name     string `xorm:"notnull VARCHAR(10)"`
	Desc     string `xorm:"VARCHAR(10)"`
	Quantity int32  `xorm:"notnull INT(4)"`
}

type ShopGoods struct {
	ShopId  int64 `xorm:"pk notnull BIGINT(11)"` //外键=>ShopInfo.Id
	GoodsId int64 `xorm:"pk notnull BIGINT(11)"` //外键=>GoodsInfo.Id
}

type GoodsAllInfo struct {
	Shop_id  int64
	Goods_id int64
	Name     string
	Desc     string
	Quantity int32
}
