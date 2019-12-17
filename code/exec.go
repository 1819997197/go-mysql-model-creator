package code

import (
	"errors"
	"flag"
	"fmt"
	"github.com/laixyz/utils"
	"os"
	"strings"

	"github.com/laixyz/sqlxyz"
)

//Exec 执行函数
func Exec() {
	defer func() {
		if info := recover(); info != nil {
			fmt.Println("[error]", info)
		} else {
			fmt.Println("[succ]")
		}
	}()
	var arg = os.Args
	ProjectCommand := strings.ToLower(ProjectName)
	if len(arg) == 1 || arg[1] == "version" {
		fmt.Println("\n " + ProjectCommand + " v " + Version)
		fmt.Println("\n 使用:")
		fmt.Println(" 生成所有表:\n " + ProjectCommand + " -conf=./test.conf -dist=../model -connect=default")
		fmt.Println(" 只生成members表:\n " + ProjectCommand + " -conf=./test.conf -dist=../model -connect=default -table=members")
		fmt.Println(" 只生成members和members_messages表:\n " + ProjectCommand + " -conf=./test.conf -dist=../model -connect=default -table=members,members_messages")
		fmt.Println("\n 配置文件范例test.conf:")
		fmt.Println("[mysql]")
		fmt.Println("host=localhost")
		fmt.Println("user=test")
		fmt.Println("password=test")
		fmt.Println("db=test")
		fmt.Println("port=3306")
		fmt.Println("charset=utf8")
		return
	}
	var configFile string
	flag.StringVar(&configFile, "conf", "./config.conf", "请使用指定配置文件目录")
	var distPath string
	flag.StringVar(&distPath, "dist", "./dist", "指定输出文件目录")

	var ConnectID string
	flag.StringVar(&ConnectID, "connect", "default", "指定使用数据库的配置的ConnectID")

	var destTableName string
	flag.StringVar(&destTableName, "table", "", "表名, 缺省时则生成所有表, 指定表名则只生成指定表的文件,多个表时以半角逗号隔开")

	flag.BoolVar(&DebugMode, "debug", false, "是否开启调试模式，调试模式不会生成文件")

	flag.Parse()
	config, err := IniFileLoad(configFile)
	if err != nil {
		panic(err.Error())
		return
	}
	if ConnectID == "" {
		ConnectID = "default"
	}
	var destTable []string
	if destTableName != "" {
		tableNames := strings.Split(destTableName, ",")
		for _, t := range tableNames {
			t = strings.TrimSpace(t)
			if t != "" {
				destTable = StringArrayAppend(destTable, t)
			}
		}
	}
	var packageName string
	packageName, err = utils.GetPackageName(distPath)
	if err != nil {
		Panic("发生错误: ", err.Error())
		return
	}

	var cfg sqlxyz.MySQLConfig
	if err = config.Cfg.Section("mysql").MapTo(&cfg); err != nil {
		Panic("发生错误: ", err.Error())
		return
	}
	var modelInit ModelInit
	if err = config.Cfg.Section("model").MapTo(&modelInit); err != nil {
		modelInit.State = []string{"State"}
		modelInit.Created = []string{"Created"}
		modelInit.Updated = []string{"Updated"}
		modelInit.Deleted = []string{"Deleted", "Deprecated"}
	}
	cfg.ParseTime = true
	err = sqlxyz.Register("default", cfg)
	if err != nil {
		Panic("sqlxyz.register : ", err.Error())
		return
	}
	db, err := sqlxyz.Using("default")
	if err != nil {
		Panic("sqlxyz.Using: ", err.Error())
		return
	}

	tmpTables, err := Tables(db, ConnectID, destTable, modelInit)
	if err != nil {
		Panic("Tables: ", err.Error())
		return
	}
	for _, t := range tmpTables {
		doc := TableDoc(t)
		header := Header(t, packageName)
		err = utils.FileWrite(distPath+"/"+t.TableAlias+".go", header+doc)
		if err != nil {
			fmt.Println("FileWrite "+distPath+"/"+t.TableAlias+".go"+" 发生错误: ", err.Error())
		}
		//Debug2Json(t.FieldNames)
	}

	//创建const.go
	err = ConstFileCreate(packageName, ConnectID, distPath)
	if err != nil {
		Panic("ConstFileCreate", err.Error())
		return
	}

}

// ConstFileCreate 创建const.go
func ConstFileCreate(packageName, ConnectID string, distPath string) (err error) {
	var fileContent = ""
	fileContent += "package " + packageName + "\n\n"
	fileContent += "var MySQLConnectID = \"" + ConnectID + "\""
	err = utils.FileWrite(distPath+"/const.go", fileContent)
	if err != nil {
		return errors.New("FileWrite " + distPath + "/const.go" + "发生错误: " + err.Error())
	}
	return nil
}
