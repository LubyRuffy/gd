/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Xxianglei
 */

package main

import (
	"github.com/Xxianglei/gd/dlog"
	"github.com/Xxianglei/gd/runtime/gl"
	"strconv"
	"time"
)

func main() {
	var i chan int
	gl.Init()
	defer gl.Close()
	gl.Set(gl.LogId, strconv.FormatInt(time.Now().UnixNano(), 10))
	dlog.LoadConfiguration("log.xml")
	dlog.Debug("test:%s", "ok")
	<-i
}
