# inject
```text
go语言实现的依赖注入库

dependency injection library based on golang implementation
```
 

##安装
```text
go get -u github.com/hhaojin/bingo-inject
```

##使用

```go
package main

import (
	"fmt"
	"github.com/hhaojin/bingo-inject/inject"
)

func main() {
	injecter := inject.New()
	//快速使用
	ordersvc := &OrderSvc{}
	injecter.Apply(ordersvc)
	ordersvc.OrderItemSvc.GetOrderItem()
	ordersvc.UserSvc.GetUser()

	//也可以绑定接口
	injecter.Mapping((*IUser)(nil), &UserSvc{})

	//通过接口从容器里获取实体
	ni := injecter.Get((*IOrderItem)(nil), nil)
	fmt.Println(fmt.Sprintf("%#v", ni)) //&OrderItemSvc{}

	//也可以直接调用func
	injecter.Set("test string")
	injecter.Invoke(func(u *UserSvc, o IOrderItem, str string) {
		u.GetUser()
		o.(*OrderItemSvc).GetOrderItem()
		fmt.Println(str) //test string
	})
}

type IUser interface {
	GetUser()
}

type UserSvc struct{}

func (us *UserSvc) GetUser() {}

type IOrderItem interface {
	GetOrderItem()
}

type OrderItemSvc struct{}

func (os *OrderItemSvc) GetOrderItem() {}

type OrderSvc struct {
	UserSvc      *UserSvc      `inject:"-"` //打上inject tag
	OrderItemSvc *OrderItemSvc `inject:"-"`
}
```
