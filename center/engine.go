package center

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

const (
	SystemKey    string = "/system"
	BussinessKey string = "/bussiness"
	SwitchKey    string = "/switch"
)

// Engine 处理更新等操作
type Engine struct {
	_Business *Bussiness
	_Switch   *OnOff
	_System   *System
	_Ctx      context.Context
	_Cancel   context.CancelFunc
	_Prefix   string
	_Wg       sync.WaitGroup
}

// New is used to instance configure
func New(ctx context.Context, prefix string) *Engine {
	inst := new(Engine)
	inst._Prefix = prefix
	inst._Business = new(Bussiness)
	inst._Switch = new(OnOff)
	inst._System = new(System)
	inst._Ctx, inst._Cancel = context.WithCancel(ctx)
	return inst
}

// _LoadFromEtcd is used to get key's value,and parse data to struct
func (e *Engine) _LoadFromEtcd(client *clientv3.Client) error {

	mGetRsp, mGetErr := client.Get(e._Ctx, e._Prefix, clientv3.WithPrefix())
	if mGetErr != nil {
		return mGetErr
	}
	for _, item := range mGetRsp.Kvs {
		e._DealKV(item.Key, item.Value)
	}
	return nil
}

// GetSystem is used to system's configure
func (e *Engine) GetSystem() *System {
	return (*System)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&e._System))))
}

// GetBussiness is used to bussiness's configure
func (e *Engine) GetBussiness() *Bussiness {
	return (*Bussiness)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&e._Business))))
}

// GetSwitch is used to switch's configure
func (e *Engine) GetSwitch() *OnOff {
	return (*OnOff)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&e._Switch))))
}

// _DealKV is used to bussiness's configure
func (e *Engine) _DealKV(bKey, bVal []byte) error {
	key := string(bKey)
	key = strings.ReplaceAll(key, e._Prefix, "")
	switch key {
	case SystemKey:
		inst := new(System)
		if err := yaml.Unmarshal(bVal, inst); err != nil {
			return err
		}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&e._System)), unsafe.Pointer(inst))
	case SwitchKey:
		inst := new(OnOff)
		if err := yaml.Unmarshal(bVal, inst); err != nil {
			return err
		}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&e._Switch)), unsafe.Pointer(inst))
	case BussinessKey:
		inst := new(Bussiness)
		if err := yaml.Unmarshal(bVal, inst); err != nil {
			return err
		}
		atomic.StorePointer((*unsafe.Pointer)(unsafe.Pointer(&e._System)), unsafe.Pointer(inst))
	default:
		fmt.Println("unknown key", key)
	}

	return nil

}

// Exit is used to resource backoff
func (e *Engine) Exit() {
	e._Cancel()
	e._Wg.Wait()
}

// Watch is used to listen etcd's change
func (e *Engine) Watch(client *clientv3.Client) error {

	//	step 1: watch the prefix key of etcd
	cWatchChan := client.Watch(e._Ctx, e._Prefix, clientv3.WithPrefix())
	if err := e._LoadFromEtcd(client); err != nil {
		return err
	}
	e._Wg.Add(1)
	go func() {
		defer e._Wg.Done()
		for {
			select {
			case <-e._Ctx.Done():
				return
			case mEvent, ok := <-cWatchChan:
				if !ok {
					cWatchChan = client.Watch(e._Ctx, e._Prefix, clientv3.WithPrefix())
					continue
				}
				for _, event := range mEvent.Events {
					if event.Type == clientv3.EventTypePut {
						if err := e._DealKV(event.Kv.Key, event.Kv.Value); err != nil {
							fmt.Println("this is error:", err.Error())
						} else {
							fmt.Println(e.GetBussiness(), e.GetSwitch(), e.GetSystem())
						}

					}
				}
			}
		}
	}()

	return nil
}
