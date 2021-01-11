/**
 * Copyright 2018 gd Author. All Rights Reserved.
 * Author: Xxianglei
 */

package dogrpc_test

import (
	"github.com/Xxianglei/gd"
	"github.com/Xxianglei/gd/utls/network"
	"testing"
	"time"
)

func TestRpcClient(t *testing.T) {
	d := gd.Default()
	c := d.NewRpcClient(time.Duration(500*time.Millisecond), 0)
	c.AddAddr(network.GetLocalIP() + ":10241")

	body := []byte("How are you?")

	code, rsp, err := c.Invoke(1024, body)
	if err != nil {
		t.Logf("Error when sending request to server: %s", err)
	}

	t.Logf("code=%d, resp=%s", code, string(rsp))
}
