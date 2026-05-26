package generated

import (
	"time"
)

type 订单 struct {
	商品 []商品 `json:"商品"`
	用户 []用户 `json:"用户"`
}

type 商品 struct {
	库存 []库存 `json:"库存"`
}

type 用户 struct {
	地址 []地址 `json:"地址"`
}

type 电子产品 struct {
	商品 string `json:"商品"`
}

