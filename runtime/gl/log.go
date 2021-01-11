/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Xxianglei
 */

package gl

var log Logger

type Logger interface {
	Debug(format interface{}, v ...interface{})
	Info(format interface{}, v ...interface{})
	Error(format interface{}, v ...interface{}) error
}

func SetLogger(l Logger) {
	log = l
}
