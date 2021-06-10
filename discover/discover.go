package discover

import (
	"context"
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Hub struct {
	_Prefix string    // 前缀
	_Store  *sync.Map //
	_Cancel context.CancelFunc
	_Ctx    context.Context
	_Wg     *sync.WaitGroup
}

func New(ctx context.Context) *Hub {
	inst := new(Hub)
	inst._Store = new(sync.Map)
	inst._Wg = new(sync.WaitGroup)
	inst._Ctx, inst._Cancel = context.WithCancel(ctx)
	return inst
}

func (h *Hub) Run(client *clientv3.Client, prefix string) error {
	h._Prefix = prefix
	cWatchChan := client.Watch(h._Ctx, h._Prefix, clientv3.WithPrefix())
	// 查询当前已有的列表
	mGetRsp, mGetErr := client.Get(h._Ctx, h._Prefix, clientv3.WithPrefix())
	if mGetErr != nil {
		return mGetErr
	}

	for _, item := range mGetRsp.Kvs {
		h._Add(string(item.Key), string(item.Value))
	}
	h._Wg.Add(1)
	go func() {
		defer h._Wg.Done()
		for {
			select {
			case <-h._Ctx.Done():
				return
			case events, ok := <-cWatchChan:
				if !ok {
					cWatchChan = client.Watch(h._Ctx, h._Prefix)
					continue
				}
				for _, ev := range events.Events {
					switch ev.Type {
					case clientv3.EventTypePut:
						h._Add(string(ev.Kv.Key), string(ev.Kv.Value))
					case clientv3.EventTypeDelete:
						h._Delete(string(ev.Kv.Key))
					}
				}
			}
		}
	}()
	return nil
}

// List is used to return service list
func (h *Hub) List() []string {
	list := []string{}
	h._Store.Range(func(key, value interface{}) bool {
		list = append(list, value.(string))
		return true
	})
	return list

}

func (h *Hub) _Delete(key string) {
	h._Store.Delete(key)
}

func (h *Hub) _Add(key, val string) {
	h._Store.Store(key, val)
}

// Exit is used to resource backoff
func (h *Hub) Exit() {
	h._Cancel()
	h._Wg.Wait()
}
