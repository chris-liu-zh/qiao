package DB

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var Pool PoolPart

type PoolPart struct {
	PoolCount int
	Master    PoolConn
	Slave     PoolConn
	Alone     PoolConn
}

type PoolConn struct {
	PoolNum int
	DBConn  []ConnDB
}

type DBList struct {
	List []DBNew
}

type DBNew struct {
	ID          int    `json:"ID"`
	Title       string `json:"title"`
	Type        string `json:"Type"`
	Open        bool   `json:"Open"`
	Part        string `json:"Part"`
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
	ID      int
	Title   string
	Part    string //角色，主从，独立
	Sign    string //占位符
	Dsn     string
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

type GetReturn func(*ConnDB, string, ...any) (int64, error)

func PGpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(Select+"LIMIT %d OFFSET %d", size, page)
	return mapper
}

func MYpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(Select+"LIMIT %d , %d", size, page)
	return mapper
}

func MSpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf(`select top %d ${field} from (select row_number() over(${order}) as rownumber,${joinfield} from ${table} ${join} ${where}) temp_row where rownumber > %d;`, size, page*size-size)
	return mapper
}

func (list DBList) InitDB() error {
	for _, v := range list.List {
		return v.NewDB()
	}
	return nil
}

func (db DBNew) NewDB() (err error) {
	conndb := ConnDB{
		ID:    db.ID,
		Title: fmt.Sprintf("%s:%d", db.Title, db.ID),
		Dsn:   db.Dsn,
	}
	var drive string
	switch db.Type {
	case "pgsql":
		conndb.Sign = "$"
		conndb.DBFunc.Page = PGpage
		conndb.DBFunc.AddReturnId = PgsqlAddReturnId
		drive = "postgres"
		if db.Dsn == "" {
			conndb.Dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d", db.Host, db.Port, db.User, db.Pwd, db.DBName, db.TimeOut)
		}
	case "mysql":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = MyqlAddReturnId
		drive = "mysql"
		if db.Dsn == "" {
			conndb.Dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&parseTime=true&loc=Local", db.User, db.Pwd, db.Host, db.Port, db.DBName, db.TimeOut)
		}
	case "sqlite":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = MyqlAddReturnId
		drive = "sqlite3"
		conndb.Dsn = db.Dsn
	case "mssql":
		conndb.Sign = "@p"
		conndb.DBFunc.Page = MSpage
		conndb.DBFunc.AddReturnId = MssqlAddReturnId
		drive = "sqlserver"
		if db.Dsn == "" {
			conndb.Dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&dial+timeout=%d&encrypt=disable&parseTime=true", db.User, db.Pwd, db.Host, db.Port, db.DBName, db.TimeOut)
		}
	}

	conn, err := sql.Open(drive, conndb.Dsn)
	if err != nil {
		conndb.log(err.Error(), "").logERROR()
		return
	}
	if err = conn.Ping(); err != nil {
		conndb.log(err.Error(), "").logERROR()
		return
	}
	conn.SetMaxOpenConns(db.MaxOpen)
	conn.SetMaxIdleConns(db.MaxIdle)

	var duration time.Duration
	if db.MaxIdleTime == "" {
		duration = 7 * time.Hour
	}

	//设置最大空闲超时
	if duration == 0 {
		if duration, err = time.ParseDuration(db.MaxIdleTime); err != nil {
			conndb.log("format error, MaxIdleTime will default to 7 hours", "").logWARNING()
			duration = 7 * time.Hour
		}
	}

	conn.SetConnMaxIdleTime(duration)

	conndb.DBFunc.Conn = conn
	conndb.Part = db.Part
	switch db.Part {
	case "master":
		Pool.Master.DBConn = append(Pool.Master.DBConn, conndb)
		Pool.Master.PoolNum = len(Pool.Master.DBConn)
	case "slave":
		Pool.Slave.DBConn = append(Pool.Slave.DBConn, conndb)
		Pool.Slave.PoolNum = len(Pool.Slave.DBConn)
	case "alone":
		Pool.Alone.DBConn = append(Pool.Alone.DBConn, conndb)
		Pool.Alone.PoolNum = len(Pool.Alone.DBConn)
	}
	Pool.PoolCount = Pool.Master.PoolNum + Pool.Slave.PoolNum + Pool.Alone.PoolNum
	return
}

func Stop() {
	Pool.PoolCount = 0
	parts := []string{"master", "slave", "alone"}
	for _, part := range parts {
		switch part {
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
