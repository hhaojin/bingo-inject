package tests

import (
	"github.com/hhaojin/bingo-inject/inject"
	"gopkg.in/go-playground/assert.v1"
	"testing"
)

type IUser interface {
	GetUser() string
}

type UserSvc struct{}

func (us *UserSvc) GetUser() string {
	return "user"
}

type IOrderItem interface {
	GetOrderItem() string
}

type OrderItemSvc struct {
	OrderNum string
}

func (os *OrderItemSvc) GetOrderItem() string {
	return os.OrderNum
}

type OrderSvc struct {
	UserSvc      *UserSvc      `inject:"-"` //打上inject tag
	OrderItemSvc *OrderItemSvc `inject:"-"`
}

type TestConfig struct {
	OrderSvc IOrderItem `inject:"-"`
}

//方法名任意
func (tc *TestConfig) Xxx() IUser {
	return &UserSvc{}
}

func Test_DefaultInjecter_Apply(t *testing.T) {
	t.Run("apply-依赖注入", func(t *testing.T) {
		inejcter := inject.New()
		order := &OrderSvc{}
		inejcter.Apply(order)
		order.OrderItemSvc.OrderNum = "order_num"
		order2 := &OrderSvc{}
		inejcter.Apply(order2)

		assert.Equal(t, order2.OrderItemSvc.OrderNum, "order_num")
		assert.Equal(t, order.OrderItemSvc.GetOrderItem(), order2.OrderItemSvc.GetOrderItem())
	})
}

func Test_DefaultInjecter_Configs(t *testing.T) {
	t.Run("Configs测试", func(t *testing.T) {
		inejcter := inject.New()
		order := &OrderSvc{}
		inejcter.Apply(order)
		order.OrderItemSvc.OrderNum = "order_num"

		c := &TestConfig{}
		inejcter.Configs(c)
		iu := inejcter.Get((*IUser)(nil), nil)
		assert.Equal(t, c.OrderSvc.GetOrderItem(), "order_num")
		assert.Equal(t, iu.(*UserSvc).GetUser(), "user")
	})
}

func Test_Mapping(t *testing.T) {
	t.Run("mapping", func(t *testing.T) {
		inejcter := inject.New()
		b := &OrderItemSvc{
			OrderNum: "order_num",
		}
		inejcter.Mapping((*IOrderItem)(nil), b)
		oi := inejcter.Get((*IOrderItem)(nil), nil)
		assert.Equal(t, oi.(*OrderItemSvc).GetOrderItem(), b.GetOrderItem())
	})
}

func Test_Set(t *testing.T) {
	t.Run("set", func(t *testing.T) {
		inejcter := inject.New()
		var a = "aaa"
		var i = 123
		inejcter.Set(a, i)
		assert.Equal(t, inejcter.Get(a, nil).(string), a)
		assert.Equal(t, inejcter.Get(i, nil).(int), i)
	})
}

func Test_Invoke(t *testing.T) {
	t.Run("Invoke", func(t *testing.T) {
		inejcter := inject.New()
		orderItem := &OrderItemSvc{OrderNum: "order_num"}
		testStr := "set string"
		testInt := 444
		inejcter.Apply(orderItem)
		inejcter.Set(testStr, testInt)
		ret, err := inejcter.Invoke(func(s string, i int, orderItem IOrderItem) (string, int, string, string) {
			return s, i, orderItem.GetOrderItem(), orderItem.(*OrderItemSvc).GetOrderItem()
		})
		if err != nil {
			t.Error(err)
		}
		assert.Equal(t, len(ret), 4)
		assert.Equal(t, ret[0], testStr)
		assert.Equal(t, ret[1], testInt)
		assert.Equal(t, ret[2], orderItem.GetOrderItem())
		assert.Equal(t, ret[3], orderItem.OrderNum)
	})
}

func Test_GetInterface(t *testing.T) {
	t.Run("interface", func(t *testing.T) {
		inejcter := inject.New()
		oi := &OrderItemSvc{
			OrderNum: "order_num",
		}
		inejcter.Set(oi)
		ib := inejcter.Get((*IOrderItem)(nil), nil)
		assert.Equal(t, ib.(IOrderItem).GetOrderItem(), oi.GetOrderItem())
	})
}
