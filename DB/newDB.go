package DB

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/glebarez/go-sqlite"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

var Pool DBPool

type DBPool struct {
	PoolCount         int
	SwitchRole        bool
	ReconnectNum      int           //重连次数
	ReconnectInterval time.Duration //重连间隔时间
	Master            PoolConn
	Slave             PoolConn
	Alone             PoolConn
}

type PoolConn struct {
	PoolNum int
	DBConn  []ConnDB
}

type Config struct {
	ID          int    `json:"ID"`
	Title       string `json:"title"`
	Type        string `json:"Type"`
	Open        bool   `json:"Open"`
	Role        string `json:"Role"`
	Host        string `json:"Host"`
	Port        int    `json:"Port"`
	User        string `json:"User"`
	Pwd         string `json:"Pwd"`
	DBName      string `json:"DBName"`
	Dsn         string `json:"Dsn"`
	MaxIdle     int    `json:"MaxIdle"`
	MaxOpen     int    `json:"MaxOpen"`
	TimeOut     int    `json:"TimeOut"`
	MaxIdleTime string `json:"MaxIdleTime"`
}

type ConnDB struct {
	Conf    Config
	Sign    string
	Err     error
	IsClose bool //连接是否关闭
	DBFunc  dbFunc
}

type dbFunc struct {
	Page
	Return
	Conn *sql.DB
}

type Return struct {
	AddReturnId GetReturn
	Affected    GetReturn
}

type Page func(*Mapper, int, int) *Mapper

type GetReturn func(string, ...any) (int64, error)

func PGpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(Select+"LIMIT %d OFFSET %d", size, page)
	return mapper
}

func MYpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(Select+"LIMIT %d , %d", page-1, size)
	return mapper
}

func MSpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(`select top %d ${field} from (select row_number() over(${order}) as rownumber,${joinfield} from ${table} ${join} ${where}) temp_row where rownumber > %d;`, size, page*size-size)
	return mapper
}

func InitDB(switchRole bool, ReconnectNum int, ReconnectInterval time.Duration, conf ...Config) error {
	for _, v := range conf {
		if v.Open {
			if err := v.NewDB(); err != nil {
				slog.Error(err.Error())
			}
		}
	}

	if Pool.PoolCount == 0 {
		return errors.New("没有打开的数据库")
	}
	Pool.ReconnectInterval = ReconnectInterval
	Pool.ReconnectNum = ReconnectNum
	Pool.SwitchRole = switchRole
	return nil
}

func (conf Config) NewDB() (err error) {
	conndb := ConnDB{
		Conf: conf,
	}
	var drive string
	switch conndb.Conf.Type {
	case "pgsql":
		conndb.Sign = "$"
		conndb.DBFunc.Page = PGpage
		conndb.DBFunc.AddReturnId = QiaoDB().PgsqlAddReturnId
		drive = "postgres"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d", conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	case "mysql":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = QiaoDB().MysqlAddReturnId
		drive = "mysql"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&parseTime=true&loc=Local", conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	case "sqlite":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = QiaoDB().MysqlAddReturnId
		drive = "sqlite3"
	case "mssql":
		conndb.Sign = "@p"
		conndb.DBFunc.Page = MSpage
		conndb.DBFunc.AddReturnId = QiaoDB().MssqlAddReturnId
		drive = "sqlserver"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&dial+timeout=%d&encrypt=disable&parseTime=true", conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	}

	conn, err := sql.Open(drive, conndb.Conf.Dsn)
	if err != nil {
		conndb.log("get sql error", conndb.Conf.Dsn).logERROR(err)
		return
	}
	if err = conn.Ping(); err != nil {
		conndb.log("get sql error", conndb.Conf.Dsn).logERROR(err)
		return
	}
	conn.SetMaxOpenConns(conndb.Conf.MaxOpen)
	conn.SetMaxIdleConns(conndb.Conf.MaxIdle)

	var duration time.Duration
	if conndb.Conf.MaxIdleTime == "" {
		duration = 7 * time.Hour
	}

	//设置最大空闲超时
	if duration == 0 {
		if duration, err = time.ParseDuration(conndb.Conf.MaxIdleTime); err != nil {
			conndb.log("format error, MaxIdleTime will default to 7 hours", "").logWARNING()
			duration = 7 * time.Hour
		}
	}

	conn.SetConnMaxIdleTime(duration)

	conndb.DBFunc.Conn = conn
	switch conndb.Conf.Role {
	case "master":
		Pool.Master.DBConn = append(Pool.Master.DBConn, conndb)
		Pool.Master.PoolNum = len(Pool.Master.DBConn)
	case "slave":
		Pool.Slave.DBConn = append(Pool.Slave.DBConn, conndb)
		Pool.Slave.PoolNum = len(Pool.Slave.DBConn)
	default:
		Pool.Alone.DBConn = append(Pool.Alone.DBConn, conndb)
		Pool.Alone.PoolNum = len(Pool.Alone.DBConn)
	}
	Pool.PoolCount = Pool.Master.PoolNum + Pool.Slave.PoolNum + Pool.Alone.PoolNum
	return
}

func Stop() {
	Pool.PoolCount = 0
	roles := []string{"master", "slave", "alone"}
	for _, role := range roles {
		switch role {
		case "master":
			close(Pool.Master.DBConn)
			Pool.Master.PoolNum = 0
			Pool.Master.DBConn = nil
		case "slave":
			close(Pool.Slave.DBConn)
			Pool.Slave.PoolNum = 0
			Pool.Slave.DBConn = nil
		case "alone":
			close(Pool.Alone.DBConn)
			Pool.Alone.PoolNum = 0
			Pool.Alone.DBConn = nil
		}
	}
}

func close(db []ConnDB) {
	for _, conn := range db {
		conn.DBFunc.Conn.Close()
	}
}
