package inject

import (
	"fmt"
	"reflect"
)

type (
	GetterFunc func(rt reflect.Type) interface{}

	Mapper interface {
		//Get 从容器里获取对象
		//key可以是 reflect.Type，也可以是某种类型或者接口和变量
		//f, 单容器里面没有找到相应的值的时候，会尝试调用f
		Get(key interface{}, f GetterFunc) interface{}

		//Set 把对象放入容器里，对象可以是任意类型
		//注意，同一种类型只会有一个值
		Set(objs ...interface{})

		//Mapping 设置映射关系，用于接口类型和实体的映射
		//key必须是指针类型的接口，例如：(*MyInterface)(nil)
		//val必须是指针类型的结构体
		Mapping(key, val interface{})
	}

	Injector interface {
		//Apply 对对象进行递归注入，递归调用 Apply 方法， obj必须是指针结构体
		//需要进行注入的字段，必须打上tag: `inject:"-"`
		//会优先从容器里获取注入对象，没有则实例一个，并把实例对象放入容器保存
		//所有的注入对象都是单例，使用的时候需要注意
		Apply(obj interface{})

		//Config 会对对象进行依赖注入
		//并且， 会自动调用对象的所有方法，如果对象方法有且仅有一个返回值，那么返回值会保存再 Mapper 里面
		//对象返回的返回值可以是容器或者指针struct
		Configs(objs ...interface{})
	}

	Invoker interface {
		//Invoke 尝试把f当做func执行，如果f不是func会panic；
		//在容器里尝试获取f的参数对应的值，作为参数传入执行f，如果没有相应的值会执行失败，返回error
		//f的参数值可以通过 Mapper.Set 初始化
		Invoke(f interface{}) ([]interface{}, error)
	}

	Injecter struct {
		mapper BeanMapper
	}
)

const (
	InjectTag    = "inject"
	InjectTagVal = "-"
)

func New() *Injecter {
	return &Injecter{mapper: make(BeanMapper)}
}

func (injecter *Injecter) Get(key interface{}, f GetterFunc) (ni interface{}) {
	var rt reflect.Type
	if kt, ok := key.(reflect.Type); ok {
		rt = kt
	} else {
		rt = reflect.TypeOf(key)
	}
	rv := injecter.mapper.get(rt)
	if rv.IsValid() {
		return rv.Interface()
	}
	if f != nil {
		return f(rt)
	}
	return
}

func (injecter *Injecter) Set(objs ...interface{}) {
	if objs == nil || len(objs) == 0 {
		return
	}
	for _, obj := range objs {
		injecter.mapper.set(reflect.TypeOf(obj), reflect.ValueOf(obj))
	}
}

func (injecter *Injecter) Apply(obj interface{}) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("the obj must be ptr: %#v", obj))
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		panic(fmt.Sprintf("the obj must be struct: %v", rv.Kind()))
	}
	for i := 0; i < rv.NumField(); i++ {
		field := rv.Type().Field(i)
		if !rv.Field(i).CanSet() || field.Tag.Get(InjectTag) != InjectTagVal {
			continue
		}
		kind := rv.Field(i).Type().Kind()
		if kind != reflect.Ptr && kind != reflect.Interface {
			panic("the field type must be ptr or interface")
		}
		if getV := injecter.Get(field.Type, func(rt reflect.Type) interface{} {
			if rt.Kind() != reflect.Ptr {
				return nil
			}
			n := reflect.New(rt.Elem())
			if n.IsValid() {
				return n.Interface()
			}
			return nil
		}); getV != nil {
			injecter.Apply(getV)
			injecter.Set(getV)
			rv.Field(i).Set(reflect.ValueOf(getV))
		}
	}
	injecter.Set(obj)
}

func (injecter *Injecter) Configs(objs ...interface{}) {
	for _, obj := range objs {
		injecter.config(obj)
	}
}

func (injecter *Injecter) config(obj interface{}) {
	injecter.Apply(obj)
	rv := reflect.ValueOf(obj)
	for i := 0; i < rv.NumMethod(); i++ {
		ret := rv.Method(i).Call(nil)
		if ret != nil && len(ret) == 1 {

			injecter.Set(ret[0].Interface())
		}
	}
}

func (injecter *Injecter) Mapping(key, val interface{}) {
	t := reflect.TypeOf(key).Elem()
	if t.Kind() != reflect.Interface {
		panic("key must be interface")
	}
	if !reflect.ValueOf(val).Type().Implements(t) {
		panic("val must be implement key")
	}
	injecter.Apply(val)
	injecter.mapper.set(reflect.TypeOf(key), reflect.ValueOf(val))
}

func (injecter *Injecter) Invoke(f interface{}) ([]interface{}, error) {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		panic("f must be func")
	}
	var in = make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		val := injecter.mapper.get(argType)
		if !val.IsValid() {
			return nil, fmt.Errorf("Value not found for type %v", argType)
		}
		in[i] = val
	}
	res := reflect.ValueOf(f).Call(in)
	var result = make([]interface{}, len(res))
	for i, val := range res {
		result[i] = val.Interface()
	}
	return result, nil
}
