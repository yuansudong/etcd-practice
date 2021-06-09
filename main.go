package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/yuansudong/etcd-practice/discover"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var wg sync.WaitGroup

func main() {

	rootctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		fmt.Println(err.Error())
	}
	defer cli.Close()
	//register.Do(rootctx, cli, "/api.hfdy.com/user", "127.0.0.1:8089")
	inst := discover.New(rootctx)
	if err = inst.Run(cli, "/api.hfdy.com/user"); err != nil {
		fmt.Println(err.Error())
	}
	for i := 0; i < 100; i++ {
		fmt.Println("当前的节点信息是：", inst.List())
		time.Sleep(time.Second)
	}
	c := make(chan os.Signal, 10)
	//监听指定信号 ctrl+c kill
	signal.Notify(c, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM,
		syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	fmt.Println("wait single come in")
	for s := range c {
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			log.Println("receive singnal ")
			cancel()
			return
		case syscall.SIGUSR1:
			fmt.Println("usr1 signal", s)
		case syscall.SIGUSR2:
			fmt.Println("usr2 signal", s)
		default:
			fmt.Println("other signal", s)
		}
	}

}
