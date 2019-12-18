# go-mysql-model-creator
快速将mysql指定数据库或表生成golang struct和表的常用方法集代码的小工具


命令行调用范例

	# 生成所有表
	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default  
	# 生成单表
	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default -table=member 
	# 生成多表
	go-mysql-model-creator -conf=./test.conf -dist=../model -connect=default -table=member,member_message 

命令行参数说明
   
    conf : 指定配置文件
    dist : 代码生成存放路径
    connect: 给数据库复用连接指一个唯一标准，不指定，默认为 default
    table : 不指定时为所有表进行生成。指定为表名，多个表以半角逗号隔开 
   
配置文件范例

    [mysql]
    host=localhost
    user=ba
    password=babababa
    db=ba
    port=3306
    charset=utf8
    [model]
    state=State
    created=Created
    updated=Updated
    deleted=Deleted, Deprecated
    
  

