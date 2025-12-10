package DB

import (
	"math/rand"
	"time"
)

func GetRand() *rand.Rand {
	return rand.New(rand.NewSource(time.Now().UnixNano()))
}

func GetSlave() *ConnDB {
	return Pool.Slave.getDB()
}

func GetMaster() *ConnDB {
	return Pool.Master.getDB()
}

func GetAlone() *ConnDB {
	return Pool.Alone.getDB()
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
	return Pool.Master.getOnlineDB()
}

func GetSlaveDB() *ConnDB {
	return Pool.Slave.getOnlineDB()
}

func (rolePool *PoolConn) getOnlineDB() *ConnDB {
	if Pool.Master.PoolNum == 0 {
		return nil
	}
	for i := range rolePool.DBConn {
		conn := &rolePool.DBConn[i]
		if conn.IsClose {
			continue
		}
		if ok := conn.checkOnline(); ok {
			return conn
		}
	}
	return nil
}

func (rolePool *PoolConn) getDB() *ConnDB {
	n := rolePool.PoolNum
	if n == 0 {
		return nil
	}

	// 收集可用连接的索引
	avail := make([]int, 0, n)
	for i := range n {
		conn := &rolePool.DBConn[i]
		if conn.IsClose {
			continue
		}
		avail = append(avail, i)
	}

	if len(avail) == 0 {
		return nil
	}

	// 随机选择一个可用连接
	idx := GetRand().Intn(len(avail))
	return &rolePool.DBConn[avail[idx]]
}
