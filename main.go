package main

import "github.com/laixyz/go-mysql-model-creator/code"

/*

功能: 针对mysql数据库的所有表创建golang需要的结构体及查询全表全字段的SQL语句常量

使用:

	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default  生成所有表
	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default -table=members 只生成members表
	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default -table=members,members_messages 只生成members和members_messages表

数据库配置文件格式

[mysql]
host=localhost
user=test
password=test
db=test
port=3306
charset=utf8
*/

func main() {
	code.Exec()
}
