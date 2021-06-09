package register

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Do 用于注册当前的进程到ETCD中。
func Do(
	ctx context.Context,
	client *clientv3.Client,
	appid string,
	srvAddr string,
) {

	lease, err := client.Grant(ctx, 5)
	if err != nil {
		fmt.Println(err.Error())
	}
	key := filepath.Join(appid, uuid.New().String())
	fmt.Println("rand key is ", key)
	client.Put(ctx, key, srvAddr, clientv3.WithLease(lease.ID))
	keepChan, err := client.KeepAlive(ctx, lease.ID)
	if err != nil {
		fmt.Println(err.Error())
	}
	clientv3.WithIgnoreLease()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case rsp := <-keepChan:
				fmt.Println(rsp.ID, rsp.MemberId, rsp.TTL, rsp.Revision)
			}
		}
	}()

}
