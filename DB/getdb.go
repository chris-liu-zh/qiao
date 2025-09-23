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

func alone() *ConnDB {
	if Pool.Alone.PoolNum == 0 {
		return nil
	}
	id := rand.Intn(Pool.Alone.PoolNum)
	return &Pool.Alone.DBConn[id]
}

func (mapper *Mapper) Read() *ConnDB {
	if mapper.Part == "alone" {
		return alone()
	}
	if dbconn := GetSlave(); dbconn != nil {
		return dbconn
	}
	return GetMaster()
}

func (mapper *Mapper) Write() *ConnDB {
	if mapper.Part == "alone" {
		return alone()
	}
	if dbconn := GetMaster(); dbconn != nil {
		return dbconn
	}
	return GetSlave()
}

func GetNewPool(part string) (conn *ConnDB) {
	switch part {
	case "master":
		if conn = GetMasterDB(); conn != nil {
			return
		}
		if conn = GetSlaveDB(); conn != nil {
			return
		}
	case "slave":
		if conn = GetSlaveDB(); conn != nil {
			return
		}
		if conn = GetMasterDB(); conn != nil {
			return
		}
	case "alone":
		if conn = GetAloneDB(); conn != nil {
			return
		}
	}
	return nil
}

func GetMasterDB() *ConnDB {
	for _, conn := range Pool.Master.DBConn {
		if db, ok := online(&conn); ok {
			return db
		}
	}
	Pool.Master.PoolNum = 0
	return nil
}

func GetSlaveDB() *ConnDB {
	for _, conn := range Pool.Slave.DBConn {
		if db, ok := online(&conn); ok {
			return db
		}
	}
	Pool.Slave.PoolNum = 0
	return nil
}

func GetAloneDB() *ConnDB {
	for _, conn := range Pool.Alone.DBConn {
		if db, ok := online(&conn); ok {
			return db
		}
	}
	Pool.Alone.PoolNum = 0
	return nil
}
