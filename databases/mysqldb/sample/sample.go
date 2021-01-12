/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Xxianglei
 */

package main

import (
	"fmt"
	"github.com/Xxianglei/gd/databases/mysqldb"
	"github.com/Xxianglei/gd/databases/mysqldb/orm"
	"github.com/Xxianglei/gd/dlog"
	"time"
)

type TestDB struct {
	Id       uint64  `json:"id" mysqlField:"id"`
	Name     string  `json:"name" mysqlField:"name"`
	CardId   uint64  `json:"card_id" mysqlField:"card_id"`
	Sex      string  `json:"sex" mysqlField:"sex"`
	Birthday uint64  `json:"birthday" mysqlField:"birthday"`
	Status   uint8   `json:"status" mysqlField:"status"`
	CreateTs uint64  `json:"create_time" mysqlField:"create_time"`
	UpdateTs []uint8 `json:"update_time" mysqlField:"update_time"`
}

type HoneypotSysUser struct {
	Id            int       `orm:"column(user_id);auto"  json:"id" mysqlField:"user_id"`
	UserName      string    `orm:"column(user_name);size(64);null" json:"userName"  mysqlField:"user_name"`
	Password      string    `orm:"column(password);size(256);null"  json:"password"   mysqlField:"password"`
	RoleId        int8      `orm:"column(role_id);null" json:"roleId" mysqlField:"role_id"`
	LastLoginTime time.Time `orm:"column(last_login_time);type(datetime);null" json:"lastLoginTime"  mysqlField:"last_login_time"`
	CreateTime    time.Time `orm:"column(create_time);type(datetime);null"  json:"createTime" mysqlField:"create_time"`
	Status        bool      `orm:"column(status);type(datetime);null" json:"status" mysqlField:"status"`
}

func main() {
	var i chan struct{}
	o := mysqldb.MysqlClient{DataBases: "test"}
	if err := o.Start(); err != nil {
		dlog.Error("err:%s", err)
		fmt.Println(err)
		return
	}
	db := o.GetWriteDbs().DB
	if err := orm.RegisterDataBase("default", "mysql", db); err != nil {
		fmt.Println(err)
	} else {
		orm.RegisterModel(new(HoneypotSysUser))
		orm.RunSyncdb("default", false, true)
	}
	// Query
	query := "select ? from honeypot_sys_user where user_id = ?"
	data, err := o.Query((*HoneypotSysUser)(nil), query, 1)
	if err != nil {
		dlog.Error("err:%s", err)
		return
	}
	if data == nil {
		dlog.Error("err:%s", err)
		return
	}
	dlog.Debug("%v", data.(*TestDB))

	// Add
	insert := &TestDB{
		Name:     "chucks",
		CardId:   1026,
		Sex:      "male",
		Birthday: 19991010,
		Status:   1,
	}

	err = o.Add("test", insert, true)
	if err != nil {
		dlog.Error("%s", err)
	}

	// queryList
	query = "select ? from test where name = ? "
	retList, err := o.QueryList((*TestDB)(nil), query, "chucks")
	testList := make([]*TestDB, 0)
	for _, ret := range retList {
		product, _ := ret.(*TestDB)
		testList = append(testList, product)
	}
	dlog.Debug("%v", testList[0].CardId)

	// update
	where := make(map[string]interface{}, 0)
	where["id"] = 2
	err = o.Update("test", &TestDB{Sex: "female"}, where, []string{"sex"})
	if err != nil {
		dlog.Error("%s", err)
	}
	<-i
}
