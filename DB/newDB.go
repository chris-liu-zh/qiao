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
	PoolCount         int           `json:"PoolCount"`
	SwitchRole        bool          `json:"SwitchRole"`        //是否启用主从切换
	ReconnectNum      int           `json:"ReconnectNum"`      //重连次数
	ReconnectInterval time.Duration `json:"ReconnectInterval"` //重连间隔时间
	Master            *PoolConn
	Slave             *PoolConn
	Alone             *PoolConn
}

type PoolConn struct {
	PoolNum int `json:"PoolNum"`
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
	drive   string
	Sign    string `json:"Sign"`
	Err     error  `json:"Err"`
	IsClose bool   `json:"IsClose"` //连接是否关闭
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

func init() {
	if Pool.Master == nil {
		Pool.Master = &PoolConn{}
	}
	if Pool.Slave == nil {
		Pool.Slave = &PoolConn{}
	}
	if Pool.Alone == nil {
		Pool.Alone = &PoolConn{}
	}
}

func PrintPool() {
	fmt.Println("数据库连接池信息")
	fmt.Printf("总连接数: %d\n", Pool.PoolCount)
	fmt.Printf("主库连接数: %d\n", Pool.Master.PoolNum)
	for _, v := range Pool.Master.DBConn {
		fmt.Printf("主库连接 id: %d;是否关闭:%v;DSN: %s\n", v.Conf.ID, v.IsClose, v.Conf.Dsn)
	}
	fmt.Printf("从库连接数: %d\n", Pool.Slave.PoolNum)
	for _, v := range Pool.Slave.DBConn {
		fmt.Printf("从库连接 id: %d;是否关闭:%v;DSN: %s\n", v.Conf.ID, v.IsClose, v.Conf.Dsn)
	}
	fmt.Printf("单库连接数: %d\n", Pool.Alone.PoolNum)
	for _, v := range Pool.Alone.DBConn {
		fmt.Printf("单库连接 id: %d;是否关闭:%v;DSN: %s\n", v.Conf.ID, v.IsClose, v.Conf.Dsn)
	}
}

func Reconnect(role string, id int) {
	var dbs []ConnDB
	switch role {
	case "master":
		dbs = Pool.Master.DBConn
	case "slave":
		dbs = Pool.Slave.DBConn
	case "alone":
		dbs = Pool.Alone.DBConn
	default:
		fmt.Println("role只支持master,slave,alone")
	}
	for i := range dbs {
		if dbs[i].Conf.ID == id {
			if ok := dbs[i].checkOnline(); !ok {
				fmt.Println("reconnect failed")
				return
			}
			fmt.Println("reconnect success")
			return
		}
	}
	fmt.Println("没有找到对应的数据库连接")
}

func PGpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf("%s LIMIT %d OFFSET %d", Select, size, page)
	return mapper
}

func MYpage(mapper *Mapper, size, page int) *Mapper {
	mapper.SqlTpl = fmt.Sprintf("%s LIMIT %d , %d", Select, page-1, size)
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

	if !conndb.Conf.Open {
		return
	}

	if conndb.Conf.Role == "" {
		conndb.Conf.Role = "alone"
	}

	if conndb.Conf.Type == "" {
		return errors.New("数据库类型不能为空")

	}

	switch conndb.Conf.Type {
	case "pgsql":
		conndb.Sign = "$"
		conndb.DBFunc.Page = PGpage
		conndb.DBFunc.AddReturnId = QiaoDB().PgsqlAddReturnId
		conndb.drive = "postgres"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable connect_timeout=%d", conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	case "mysql":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = QiaoDB().MysqlAddReturnId
		conndb.drive = "mysql"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?timeout=%ds&parseTime=true&loc=Local", conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	case "sqlite":
		conndb.Sign = "?"
		conndb.DBFunc.Page = MYpage
		conndb.DBFunc.AddReturnId = QiaoDB().MysqlAddReturnId
		conndb.drive = "sqlite3"
	case "mssql":
		conndb.Sign = "@p"
		conndb.DBFunc.Page = MSpage
		conndb.DBFunc.AddReturnId = QiaoDB().MssqlAddReturnId
		conndb.drive = "sqlserver"
		if conndb.Conf.Dsn == "" {
			conndb.Conf.Dsn = fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&dial+timeout=%d&encrypt=disable&parseTime=true", conndb.Conf.User, conndb.Conf.Pwd, conndb.Conf.Host, conndb.Conf.Port, conndb.Conf.DBName, conndb.Conf.TimeOut)
		}
	}
	conn, err := conndb.openSql()
	if err != nil {
		return err
	}
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

func (db *ConnDB) openSql() (*sql.DB, error) {
	sqlDB, err := sql.Open(db.drive, db.Conf.Dsn)
	if err != nil {
		db.log("get sql error", db.Conf.Dsn).logERROR(err)
		return nil, err
	}
	sqlDB.SetMaxOpenConns(db.Conf.MaxOpen)
	sqlDB.SetMaxIdleConns(db.Conf.MaxIdle)

	var duration time.Duration
	if db.Conf.MaxIdleTime == "" {
		duration = 7 * time.Hour
	}

	//设置最大空闲超时
	if duration == 0 {
		if duration, err = time.ParseDuration(db.Conf.MaxIdleTime); err != nil {
			db.log("format error, MaxIdleTime will default to 7 hours", "").logWARNING()
			duration = 7 * time.Hour
		}
	}

	sqlDB.SetConnMaxIdleTime(duration)
	return sqlDB, nil
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
