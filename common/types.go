package common

type OrderInfo struct {
	Order_number       string `form:"id" json:"id" xml:"id" binding:"required"`
	Goods_id int64  `form:"goods_id" json:"goods_id" xml:"goods_id" binding:"required"`
	Quantity int32  `form:"quantity" json:"quantity" xml:"quantity" binding:"required"`
}