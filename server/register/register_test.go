/**
 * Copyright 2018 Author. All rights reserved.
 * Author: Xxianglei
 */

package register_test

import (
	"github.com/Xxianglei/gd/server/register"
	"testing"
	"time"
)

//func TestEtcd(t *testing.T) {
//	var r register.DogRegister
//	r = &register.EtcdRegister{}
//	r.NewRegister([]string{"localhost:2379"}, "/root/", "stagging", "github", "gd")
//
//	r.Run("127.0.0.1", 10240, 10)
//	time.Sleep(3 * time.Second)
//	r.Close()
//}

func TestZk(t *testing.T) {
	var r register.DogRegister
	r = &register.ZkRegister{}
	r.NewRegister([]string{"localhost:2181"}, "/root/", "stagging", "github", "gd")
	r.Run("127.0.0.1", 10240, 10)
	time.Sleep(10 * time.Second)
	r.Close()
}
