/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package main

import (
	"fmt"
	"github.com/chuck1024/gd/databases/mysqldb"
	"github.com/chuck1024/gd/databases/mysqldb/orm"
	"github.com/chuck1024/gd/dlog"
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
	Id            int       `orm:"column(user_id);auto" description:"用户ID" json:"id" valid:"Min(-1)"`
	UserName      string    `orm:"column(user_name);size(64);null" description:"用户名称" json:"userName" valid:"Match(/^[a-z0-9_-]{2,50}$/)"`
	Password      string    `orm:"column(password);size(256);null" description:"登录密码" json:"password" valid:"Match(/^[a-z0-9]{1,200}$/)"`
	RoleId        int8      `orm:"column(role_id);null" description:"角色ID" json:"roleId" valid:"Max(999)"`
	LastLoginTime time.Time `orm:"column(last_login_time);type(datetime);null" description:"最后一次登录时间" json:"lastLoginTime"`
	CreateTime    time.Time `orm:"column(create_time);type(datetime);null" description:"创建时间" json:"createTime"`
	CreateTime2   time.Time `orm:"column(create_time2);type(datetime);null" description:"创建时间" json:"createTime2"`
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
	query := "select ? from test where id = ?"
	data, err := o.Query((*TestDB)(nil), query, 2)
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
