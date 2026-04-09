package db

import (
	"errors"
	"github.com/team-ide/framework"
)

func TableCreate(dbService IService, moduleName, version, tableName string, table *Table) (err error) {
	info := "module [" + moduleName + "] install version [" + version + "] table [" + tableName + "]"
	framework.Info(info + " create start")

	var tableCheckExists = dbService.TableCheckExists(dbService.GetSqlConn(), "", "", tableName)
	if !tableCheckExists {
		framework.Info(info + " not exists create start")
		name := table.Name
		table.Name = tableName
		defer func() {
			table.Name = name
		}()
		err = dbService.TableCreate(dbService.GetSqlConn(), dbService.GetDDLHandler(), "", "", table)
		if err != nil {
			err = errors.New(info + " create error:" + err.Error())
			framework.Error(err.Error())
			return
		}
		framework.Info(info + " create success")
	} else {
		framework.Info(info + " exists")
	}

	framework.Info(info + " create end")

	return
}

func TableColumnAdd(dbService IService, moduleName, version, tableName string, column *Column) (err error) {
	info := "module [" + moduleName + "] install version [" + version + "] table [" + tableName + "] column [" + column.Name + "]"
	framework.Info(info + " add start")

	var tableCheckExists = dbService.ColumnCheckExists(dbService.GetSqlConn(), "", "", tableName, column.Name)
	if !tableCheckExists {
		framework.Info(info + " not exists add start")

		err = dbService.ColumnAdd(dbService.GetSqlConn(), dbService.GetDDLHandler(), "", "", tableName, column)
		if err != nil {
			err = errors.New(info + " add error:" + err.Error())
			framework.Error(err.Error())
			return
		}
		framework.Info(info + " add success")
	} else {
		framework.Info(info + " exists")
	}

	framework.Info(info + " add end")

	return
}

func TableColumnDrop(dbService IService, moduleName, version, tableName string, columnName string) (err error) {
	info := "module [" + moduleName + "] install version [" + version + "] table [" + tableName + "] column [" + columnName + "]"
	framework.Info(info + " drop start")
	if columnName == "" {
		err = errors.New(info + " drop error: column name is empty")
		framework.Error(err.Error())
		return
	}
	var tableCheckExists = dbService.ColumnCheckExists(dbService.GetSqlConn(), "", "", tableName, columnName)
	if tableCheckExists {
		framework.Info(info + " exists drop start")

		err = dbService.ColumnDelete(dbService.GetSqlConn(), dbService.GetDDLHandler(), "", "", tableName, columnName)
		if err != nil {
			err = errors.New(info + " drop error:" + err.Error())
			framework.Error(err.Error())
			return
		}
		framework.Info(info + " drop success")
	} else {
		framework.Info(info + " not exists")
	}

	framework.Info(info + " drop end")

	return
}

func TableIndexAdd(dbService IService, moduleName, version, tableName string, index *Index) (err error) {
	info := "module [" + moduleName + "] install version [" + version + "] table [" + tableName + "] index [" + index.Name + "]"
	framework.Info(info + " add start")

	var tableCheckExists = dbService.IndexCheckExists(dbService.GetSqlConn(), "", "", tableName, index)
	if !tableCheckExists {
		framework.Info(info + " not exists add start")

		err = dbService.IndexAdd(dbService.GetSqlConn(), dbService.GetDDLHandler(), "", "", tableName, index)
		if err != nil {
			err = errors.New(info + " add error:" + err.Error())
			framework.Error(err.Error())
			return
		}
		framework.Info(info + " add success")
	} else {
		framework.Info(info + " exists")
	}

	framework.Info(info + " add end")

	return
}

func TableIndexDrop(dbService IService, moduleName, version, tableName string, indexName string) (err error) {
	info := "module [" + moduleName + "] install version [" + version + "] table [" + tableName + "] index [" + indexName + "]"
	framework.Info(info + " drop start")
	if indexName == "" {
		err = errors.New(info + " drop error: index name is empty")
		framework.Error(err.Error())
		return
	}

	var tableCheckExists = dbService.IndexCheckExists(dbService.GetSqlConn(), "", "", tableName, &Index{Name: indexName})
	if tableCheckExists {
		framework.Info(info + " exists drop start")

		err = dbService.IndexDelete(dbService.GetSqlConn(), dbService.GetDDLHandler(), "", "", tableName, indexName)
		if err != nil {
			err = errors.New(info + " drop error:" + err.Error())
			framework.Error(err.Error())
			return
		}
		framework.Info(info + " drop success")
	} else {
		framework.Info(info + " not exists")
	}

	framework.Info(info + " drop end")

	return
}
