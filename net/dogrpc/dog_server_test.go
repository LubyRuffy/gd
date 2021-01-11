/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Xxianglei
 */

package dogrpc_test

import (
	"github.com/Xxianglei/gd"
	de "github.com/Xxianglei/gd/derror"
	"github.com/Xxianglei/gd/dlog"
	"testing"
)

type TestReq struct {
	Data string
}

type TestResp struct {
	Ret string
}

func test(req *TestReq) (code uint32, message string, err error, ret *TestResp) {
	dlog.Debug("rpc sever req:%v", req)

	ret = &TestResp{
		Ret: "ok!!!",
	}

	return uint32(de.RpcSuccess), "ok", nil, ret
}

func TestDogServer(t *testing.T) {
	d := gd.Default()
	// Rpc
	d.RpcServer.AddDogHandler(1024, test)
	if err := d.RpcServer.DogRpcRegister(); err != nil {
		t.Logf("DogRpcRegister occur error:%s", err)
		return
	}

	err := d.RpcServer.Run(10241)
	if err != nil {
		t.Logf("Error occurs, error = %s", err.Error())
		return
	}
}
