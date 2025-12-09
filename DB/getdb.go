package DB

import (
	"math/rand"
	"time"
)

func GetRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetSlave() *ConnDB {
	if Pool.Slave.PoolNum == 0 {
		return nil
	}
	id := GetRand().Intn(Pool.Slave.PoolNum)
	return &Pool.Slave.DBConn[id]
}

func GetMaster() *ConnDB {
	if Pool.Master.PoolNum == 0 {
		return nil
	}
	id := GetRand().Intn(Pool.Master.PoolNum)
	return &Pool.Master.DBConn[id]
}

func GetAlone() *ConnDB {
	if Pool.Alone.PoolNum == 0 {
		return nil
	}
	id := rand.Intn(Pool.Alone.PoolNum)
	return &Pool.Alone.DBConn[id]
}

func (mapper *Mapper) Read() *ConnDB {
	if mapper.Role == "alone" {
		return GetAlone()
	}
	if dbconn := GetSlave(); dbconn != nil {
		return dbconn
	}
	return GetMaster()
}

func (mapper *Mapper) Write() *ConnDB {
	if mapper.Role == "alone" {
		return GetAlone()
	}
	if dbconn := GetMaster(); dbconn != nil {
		return dbconn
	}
	return GetSlave()
}

func GetNewPool(Role string) (conn *ConnDB) {
	switch Role {
	case "master":
		if conn = GetMasterDB(); conn != nil {
			return
		}
		if Pool.SwitchRole {
			if conn = GetSlaveDB(); conn != nil {
				return
			}
		}
	case "slave":
		if conn = GetSlaveDB(); conn != nil {
			return
		}
		if conn = GetMasterDB(); conn != nil {
			return
		}
	case "alone":
		if conn = GetAlone(); conn != nil {
			return
		}
	}
	return nil
}

func GetMasterDB() *ConnDB {
	if Pool.Master.PoolNum == 0 {
		return nil
	}
	for _, conn := range Pool.Master.DBConn {
		if ok := conn.checkOnline(); ok {
			return &conn
		}
	}
	Pool.Master.PoolNum = 0
	return nil
}

func GetSlaveDB() *ConnDB {
	if Pool.Slave.PoolNum == 0 {
		return nil
	}
	for _, conn := range Pool.Slave.DBConn {
		if ok := conn.checkOnline(); ok {
			return &conn
		}
	}
	Pool.Slave.PoolNum = 0
	return nil
}
