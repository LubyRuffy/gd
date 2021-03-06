/**
 * Copyright 2019 gd Author. All rights reserved.
 * Author: Xxianglei
 */

package mysqldb

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

const defaultCharSet = "utf8mb4"
const defaultLoc = "Local"

func (c *MysqlClient) initObjForMysqlDb(dbConfPath string) error {
	dbConfRealPath := dbConfPath
	if dbConfRealPath == "" {
		return errors.New("dbConf not set in g_cfg")
	}

	if !strings.HasSuffix(dbConfRealPath, ".ini") {
		return errors.New("dbConf not an ini file")
	}

	dbConf, err := ini.Load(dbConfRealPath)
	if err != nil {
		return err
	}

	if err = c.initDbs(dbConf, c.DataBases); err != nil {
		return err
	}
	return nil
}

func (c *MysqlClient) initDbs(f *ini.File, db string) error {
	m := f.Section(fmt.Sprintf("%s.%s", "Mysql", db))
	s := f.Section(fmt.Sprintf("%s.%s", "MysqlSlave", db))

	// master
	masterIp := m.Key("master_ip").String()
	masterPort := m.Key("master_port").String()
	userWrite := m.Key("user").String()
	passWrite := m.Key("password").String()
	masterProxy, _ := m.Key("master_is_proxy").Bool()

	// slave
	slaveIp := s.Key("slave_ip").String()
	slavePort := s.Key("slave_port").String()
	userRead := s.Key("user").String()
	passRead := s.Key("password").String()
	slaveProxy, _ := s.Key("slave_is_proxy").Bool()

	timeout := m.Key("timeout").String()
	if timeout == "" {
		timeout = "5s"
	} else if !strings.HasSuffix(timeout, "s") {
		timeout += "s"
	}

	parseTime, err := m.Key("parseTime").Bool()
	if err != nil {
		parseTime = true
	}

	loc := m.Key("loc").String()
	if timeout == "" {
		timeout = defaultLoc
	}

	connTimeout := m.Key("connTimeout").String()
	if connTimeout == "" {
		connTimeout = "1s"
	} else if !strings.HasSuffix(connTimeout, "s") {
		connTimeout += "s"
	}

	maxOpen, err := m.Key("max_open").Int()
	if err != nil {
		maxOpen = 100
	}
	maxIdle, err := m.Key("max_idle").Int()
	if err != nil {
		maxIdle = 1
	}

	enableSqlSafeUpdates, err := m.Key("enable_sql_safe_updates").Bool()
	if err != nil {
		enableSqlSafeUpdates = false
	}

	masterIps := strings.Split(masterIp, ",")
	connMasters := make([]string, 0)
	for _, masterIpVal := range masterIps {
		if masterIpVal == "" {
			continue
		}

		connMaster := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=%t&loc=%s", userWrite, passWrite, masterIp, masterPort, db, connTimeout, timeout, timeout, parseTime, loc)
		if enableSqlSafeUpdates {
			connMaster = connMaster + "&sql_safe_updates=1"
		}

		connMasters = append(connMasters, connMaster)
	}

	slaveIps := strings.Split(slaveIp, ",")
	connSlaves := make([]string, 0)
	for _, slaveIpVal := range slaveIps {
		if slaveIpVal == "" {
			continue
		}
		connSlaves = append(connSlaves, fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=%t&loc=%s", userRead, passRead, slaveIp, slavePort, db, connTimeout, timeout, timeout, parseTime, loc))
	}

	glSuffix := m.Key("glSuffix").String()
	to, _ := time.ParseDuration(timeout)
	return c.initMainDbsMaxOpen(connMasters, connSlaves, maxOpen, maxIdle, glSuffix, to, masterProxy, slaveProxy)
}

type CommonDbConf struct {
	DbName      string
	ConnTime    string // connect timeout
	ReadTime    string // read timeout
	WriteTime   string // write timeout
	MaxOpen     int    // connect pool
	MaxIdle     int    // max idle connect
	MaxLifeTime int64  // connect life time
	glSuffix    string
	Master      *DbConnectConf
	Slave       *DbConnectConf
}

type DbConnectConf struct {
	Addrs                []string
	User                 string
	Pass                 string
	CharSet              string // default utf8mb4
	ClientFoundRows      bool   // ??????update??????,???????????????????????????????????????,???clientFoundRows???false???,sql?????????????????????0;???clientFoundRows???true,sql??????????????????1
	IsProxy              bool
	EnableSqlSafeUpdates bool // (safe update mode)??????????????????????????????WHERE?????????????????????
	ParseTime            bool
	Loc                  string
}

func (c *MysqlClient) initDbsWithCommonConf(dbConf *CommonDbConf) error {
	if dbConf == nil {
		return errors.New("dbConf is nil")
	}
	if dbConf.Master == nil || len(dbConf.Master.Addrs) == 0 {
		return errors.New("masterAddr is nil")
	}
	if dbConf.DbName == "" {
		return errors.New("dbName is nil")
	}

	connTimeout := dbConf.ConnTime
	if connTimeout == "" {
		connTimeout = "200ms"
	}
	readTimeout := dbConf.ReadTime
	if readTimeout == "" {
		readTimeout = "500ms"
	}
	writeTimeout := dbConf.WriteTime
	if writeTimeout == "" {
		writeTimeout = "500ms"
	}
	maxOpen := dbConf.MaxOpen
	if maxOpen <= 0 {
		maxOpen = 100
	}
	maxIdle := dbConf.MaxIdle
	if maxIdle <= 0 {
		maxIdle = 1
	}

	connMasters, err := c.getReadWriteConnectString(dbConf.Master, connTimeout, readTimeout, writeTimeout, dbConf.DbName)
	if err != nil {
		return err
	}

	if len(connMasters) == 0 {
		return errors.New("no valid master ip found")
	}

	connSlave, err := c.getReadWriteConnectString(dbConf.Slave, connTimeout, readTimeout, writeTimeout, dbConf.DbName)
	if err != nil {
		return err
	}

	slaveIsProxy := false
	if dbConf.Slave != nil {
		slaveIsProxy = dbConf.Slave.IsProxy
	}

	to, err := time.ParseDuration(readTimeout)
	if err != nil {
		return fmt.Errorf("init mysqldb invalid duration %v", readTimeout)
	}

	return c.initMainDbsMaxOpen(connMasters, connSlave, maxOpen, maxIdle, dbConf.glSuffix, to, dbConf.Master.IsProxy, slaveIsProxy)
}

func (c *MysqlClient) getConnectString(conf *DbConnectConf, connTimeout, optTimeout int64, dbname string) ([]string, error) {
	if conf == nil || len(conf.Addrs) == 0 {
		return nil, nil
	}

	if conf.CharSet == "" {
		conf.CharSet = defaultCharSet
	}
	if conf.Loc == "" {
		conf.Loc = defaultLoc
	}

	conStrs := make([]string, 0, len(conf.Addrs))
	for _, host := range conf.Addrs {
		if host != "" {
			var constr string
			if conf.ClientFoundRows {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s&clientFoundRows=true&parseTime=%t&loc=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, optTimeout, optTimeout, conf.CharSet, conf.ParseTime, conf.Loc)
			} else {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%ds&readTimeout=%ds&writeTimeout=%ds&charset=%s&parseTime=%s&loc=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, optTimeout, optTimeout, conf.CharSet, conf.ParseTime, conf.Loc)
			}

			if conf.EnableSqlSafeUpdates {
				constr = constr + "&sql_safe_updates=1"
			}

			conStrs = append(conStrs, constr)
		}
	}
	return conStrs, nil
}

func (c *MysqlClient) getReadWriteConnectString(conf *DbConnectConf, connTimeout, readTimeout, writeTimeout string, dbname string) ([]string, error) {
	if conf == nil || len(conf.Addrs) == 0 {
		return nil, nil
	}

	if conf.CharSet == "" {
		conf.CharSet = defaultCharSet
	}

	if conf.Loc == "" {
		conf.Loc = defaultLoc
	}

	constrs := make([]string, 0, len(conf.Addrs))
	for _, host := range conf.Addrs {
		if host != "" {
			var constr string
			if conf.ClientFoundRows {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s&clientFoundRows=true&parseTime=%t&loc=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, readTimeout, writeTimeout, conf.CharSet, conf.ParseTime, conf.Loc)
			} else {
				constr = fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&charset=%s&parseTime=%s&loc=%s",
					conf.User, conf.Pass, host, dbname, connTimeout, readTimeout, writeTimeout, conf.CharSet, conf.ParseTime, conf.Loc)
			}

			if conf.EnableSqlSafeUpdates {
				constr = constr + "&sql_safe_updates=1"
			}

			constrs = append(constrs, constr)
		}
	}

	return constrs, nil
}
