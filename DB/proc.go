/*
 * @Author: Chris
 * @Date: 2024-03-16 18:13:37
 * @LastEditors: Chris
 * @LastEditTime: 2025-03-05 18:50:29
 * @Description: 请填写简介
 */
package DB

import (
	"database/sql"
	"fmt"
)

type proc struct {
	declare string
	set     string
	name    string
	in      string
	output  string
	selects string
	m       *Mapper
}

func (mapper *Mapper) Proc2(name string) *proc {
	p := &proc{
		name: name,
		m:    mapper,
	}
	return p
}

func (p *proc) Declare(declare map[string]any) *proc {
	if len(declare) > 0 {
		for k, v := range declare {
			p.declare += fmt.Sprintf("declare %v %s;", k, v)
		}
	}
	return p
}

func (p *proc) Set(set map[string]any) *proc {
	if len(set) > 0 {
		for k, v := range set {
			p.set += fmt.Sprintf("set %s=N'%v';", k, v)
		}
	}
	return p
}

func (p *proc) InPut(in string) *proc {
	if in != "" {
		p.in = fmt.Sprintf("%v", in)
	}
	return p
}

func (p *proc) OutPut(output string) *proc {
	if output != "" {
		p.output = fmt.Sprintf(",%v OUTPUT;", output)
	}
	return p
}

func (p *proc) Select(Select string) *proc {
	if Select != "" {
		p.selects = fmt.Sprintf("select %v;", Select)
	}
	return p
}

func (p *proc) getSql() *proc {
	if p.declare != "" {
		p.m.Complete.Sql = p.declare
	}
	if p.set != "" {
		p.m.Complete.Sql = p.m.Complete.Sql + p.set
	}

	p.m.Complete.Sql = fmt.Sprintf("%v exec %v ", p.m.Complete.Sql, p.name)

	if p.in != "" {
		p.m.Complete.Sql = p.m.Complete.Sql + p.in
	}

	if p.output != "" {
		p.m.Complete.Sql = p.m.Complete.Sql + p.output
	}

	if p.selects != "" {
		p.m.Complete.Sql = p.m.Complete.Sql + p.selects
	}
	return p
}

func (p *proc) Query() (rows *sql.Rows, err error) {
	p = p.getSql()
	p.m.debug("Query")
	if rows, err = p.m.Write().Query(p.m.Complete.Sql); err != nil {
		return
	}
	return
}

func (mapper *Mapper) ProcQuery(procSql string, args ...any) (rows *sql.Rows, err error) {
	mapper.Complete = SqlComplete{Sql: procSql, Args: args}
	mapper.debug("ProcQuery")
	if rows, err = mapper.Write().Query(procSql, args...); err != nil {
		return
	}
	return
}

func (p *proc) Exec() (r sql.Result, err error) {
	p = p.getSql()
	if r, err = QiaoDB().ExecSql(p.m.Complete.Sql); err != nil {
		return
	}
	return
}
