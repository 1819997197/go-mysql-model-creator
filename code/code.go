package code

import "strings"

// Header 生成代码
func Header(table MysqlTable, packageName string) string {
	var fileContent = ""
	fileContent += "/*\n"
	fileContent += "\t " + table.TableAlias + " 针对数据库表 " + table.Table + " 结构体的定义及常用方法\n"
	fileContent += "\t由" + ProjectName + "工具自动生成, 详细使用请查看: " + ProjectURL + "\n"
	fileContent += "*/\n\n"
	fileContent += "package " + packageName + "\n\n"
	fileContent += "import (\n"
	fileContent += "\t\"database/sql\"\n"
	fileContent += "\t\"strings\"\n"
	if table.HasTime {
		fileContent += "\t\"time\"\n"
	}
	fileContent += "\t\"github.com/laixyz/sqlxyz\"\n"
	if table.HasExtendType {
		fileContent += "\t\"github.com/laixyz/sqlxyz/typexyz\"\n"
	}
	fileContent += ")\n"
	fileContent += "\n"
	return fileContent
}

// TableDoc 生成表的代码
func TableDoc(table MysqlTable) string {
	var sqlQuery = ""
	var structDoc = ""
	var structComment = ""
	var sqlInsert = ""
	structDoc = "\tsqlxyz.SchemaModel\n"
	var fieldDoc string

	for _, fieldname := range table.FieldNames {
		if field, ok := table.Fields[fieldname]; ok {
			fieldDoc += "\t" + field.FieldAlias + "\tbool\t`db:\"" + field.FieldName + "\"`"
			fieldDoc += "\t// " + field.FieldName + " \n"
			structDoc += "\t" + field.FieldAlias + " "
			structDoc += "\t" + field.FieldType
			structDoc += "\t`db:\"" + field.FieldName + "\"`"
			structDoc += "\t// " + field.FieldTitle + " 类型: " + field.FieldType
			if field.IsPrimary {
				structDoc += " 主健字段（Primary Key）"
			}
			if field.IsAutoIncrement {
				structDoc += " 自增长字段 "
			}
			if field.FieldDescription != "" {
				structDoc += " 说明: " + field.FieldDescription
			}
			if field.FieldDefault != "" {
				structDoc += " 默认值: " + field.FieldDefault
			}

			structDoc += "\n"
		}
	}
	var tableNameDoc string = "// " + table.TableAlias + "TableName 数据库表名\nconst " + table.TableAlias + "TableName = \"`" + table.Table + "`\"\n\n"
	structDoc = "type " + table.TableAlias + " struct {\n" + structDoc + "}\n"
	fieldDoc = "\n// " + table.TableAlias + "Fields 字段结构体 \ntype " + table.TableAlias + "Fields struct{\n" + fieldDoc + "}\n"
	sqlInsert = "\tINSERT INTO `" + table.Table + "` SET " + strings.Join(table.FieldNames, "=?, ") + "=? \n"
	sqlInsert += "\tUPDATE `" + table.Table + "` SET " + strings.Join(table.FieldNames, "=?, ") + "=? \n"
	sqlInsert += "\tDELETE FROM `" + table.Table + "` WHERE \n"
	sqlQuery = "\tSELECT " + strings.Join(table.FieldNames, ", ") + " FROM `" + table.Table + "`\n"
	fileContent := ""

	structComment = "常用SQL:\n" + sqlQuery + sqlInsert
	var structDescription string = "\n// " + table.TableAlias + " 针对数据库表 " + table.Table + " 的结构体定义\n"

	return tableNameDoc + "/*" + fileContent + "\n表结构：\n\n" + table.SQLCreate + "\n\n" + structComment + "*/\n" + structDescription + structDoc + GetStructMethod(table) + fieldDoc //+ getFieldMethodCode(table)
}

// GetStructMethod 获取表结构
func GetStructMethod(table MysqlTable) string {
	var strMethod string
	var tmp string
	var tmpInsert string
	var tmpSelect string

	whereSQL, paramInitCode, paramListCode := table.Primary2Code()

	strMethod += "\n// New" + table.TableAlias + " 新建一个" + table.TableAlias + " 对像，并指定默认值\n"
	strMethod += "func New" + table.TableAlias + "() " + table.TableAlias + " { \n"
	strMethod += "\tvar t " + table.TableAlias + "\n"
	strMethod += "\tt.ConnectID = MySQLConnectID\n"
	for _, fieldName := range table.FieldNames {
		if field, ok := table.Fields[fieldName]; ok {
			if tmpSelect != "" {
				tmpSelect += ", "
			}
			tmpSelect += "&t." + field.FieldAlias
			if field.IsPrimary && field.IsAutoIncrement {
				continue
			}
			if tmpInsert != "" {
				tmpInsert += ", "
			}
			tmpInsert += "t." + fieldName

			if field.IsInteger || field.IsExtendType || field.IsTime {
				tmp += "\tt." + field.FieldAlias + " = " + field.Default() + "\n"
			} else {
				if field.Default() == "" {
					tmp += "\tt." + field.FieldAlias + " = \"\"\n"
				}
			}
		}
	}
	strMethod += tmp
	strMethod += "\treturn t\n"
	strMethod += "}\n"
	strMethod += "\n// Ping 检查数据库连接是否正常\n"
	strMethod += "func (t *" + table.TableAlias + ") Ping() (err error) {\n"
	strMethod += "\tt.ConnectID = MySQLConnectID\n"
	strMethod += "\treturn t.SchemaModel.Ping()\n"
	strMethod += "}\n"

	tmp = "db *sqlx.DB"
	strMethod += "\n// Find 根据条件查找一条记录\n"
	strMethod += "func Find" + table.TableAlias + "(vars ...string) (t " + table.TableAlias + ", exists bool, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn t, false, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) > 0 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlOrder string\n"
	strMethod += "\tif len(vars) > 1 {\n"
	strMethod += "\t\tsqlOrder = \" ORDER BY  \"+vars[1]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"SELECT " + strings.Join(table.FieldNames, ", ") + " FROM `" + table.Table + "`\" + sqlWhere + sqlOrder\n"
	strMethod += "\terr = t.DB.QueryRow(query + \" LIMIT 1\").Scan(" + tmpSelect + ")\n"
	strMethod += "\tif err == nil { \n"
	strMethod += "\t\treturn t, true, nil\n"
	strMethod += "\t} else if err == sql.ErrNoRows {\n"
	strMethod += "\t\treturn t, false, nil\n"
	strMethod += "\t} else {\n"
	strMethod += "\t\treturn t, false, err\n"
	strMethod += "\t}\n"
	strMethod += "}\n"
	strMethod += "\n// Find 根据条件查找一条记录\n"
	strMethod += "\n// 条件实例: Find(\"`State`!=-1\") 等于 select * from table where `state`!=1 LIMIT 1\n"
	strMethod += "\n// 条件实例: Find(\"`State`!=-1\",\"ID DESC\") 等于 select * from table where `state`!=1 ORDER BY ID DESC LIMIT 1\n"
	strMethod += "func (t *" + table.TableAlias + ") Find(vars ...string) (exists bool, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn false, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) > 0 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlOrder string\n"
	strMethod += "\tif len(vars) > 1 {\n"
	strMethod += "\t\tsqlOrder = \" ORDER BY  \"+vars[1]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"SELECT " + strings.Join(table.FieldNames, ", ") + " FROM `" + table.Table + "`\" + sqlWhere + sqlOrder\n"
	strMethod += "\terr = t.DB.QueryRow(query + \" LIMIT 1\").Scan(" + tmpSelect + ")\n"
	strMethod += "\tif err == nil { \n"
	strMethod += "\t\treturn true, nil\n"
	strMethod += "\t} else if err == sql.ErrNoRows {\n"
	strMethod += "\t\treturn false, nil\n"
	strMethod += "\t} else {\n"
	strMethod += "\t\treturn false, err\n"
	strMethod += "\t}\n"
	strMethod += "}\n"

	strMethod += "\n// Count 根据条件查询一个分页结果集, 条件实例: Count(\"`State`!=-1\")\n"
	strMethod += "func (t *" + table.TableAlias + ") Count(vars ...string) (total int, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn 0, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) >= 1 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlTotal = \"SELECT count(*) as Total FROM `" + table.Table + "`\" + sqlWhere\n"
	strMethod += "\terr = t.DB.QueryRow(sqlTotal).Scan(&total)\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn 0, err\n"
	strMethod += "\t}\n"
	strMethod += "\treturn total, nil\n"
	strMethod += "}\n"

	if table.HasPrimary {
		strMethod += "\n// FindByPrimaryID 根据条件查找一条记录, 条件实例: FindByPrimaryID(1000)\n"
		strMethod += "func (t *" + table.TableAlias + ") FindByPrimaryID(" + paramInitCode + ") (exists bool, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn false, err\n"
		strMethod += "\t}\n"

		strMethod += "\tvar query = \"SELECT " + strings.Join(table.FieldNames, ", ") + " FROM `" + table.Table + " " + whereSQL + "\"\n"
		strMethod += "\terr = t.DB.QueryRow(query, " + paramListCode + ").Scan(" + tmpSelect + ")\n"
		strMethod += "\tif err == nil { \n"
		strMethod += "\t\treturn true, nil\n"
		strMethod += "\t} else if err == sql.ErrNoRows {\n"
		strMethod += "\t\treturn false, nil\n"
		strMethod += "\t} else {\n"
		strMethod += "\t\treturn false, err\n"
		strMethod += "\t}\n"
		strMethod += "}\n"
	}

	strMethod += "\n// " + table.TableAlias + "FindAll 查询所有记录\n"
	strMethod += "func " + table.TableAlias + "FindAll(vars ...string) (data []" + table.TableAlias + ", total int, err error) { \n"
	strMethod += "\tvar this " + table.TableAlias + "\n"
	strMethod += "\treturn this.FindAll(vars...)\n"
	strMethod += "}\n"

	strMethod += "\n// FindAll 根据条件查询一个结果集, 条件实例: FindAll(\"`State`!=-1\")\n"
	strMethod += "func (t *" + table.TableAlias + ") FindAll(vars ...string) (data []" + table.TableAlias + ", total int, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn data, 0, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) >= 1 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlOrder string\n"
	strMethod += "\tif len(vars) >= 2 {\n"
	strMethod += "\t\tsqlOrder = \" ORDER BY  \"+vars[1]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"SELECT " + strings.Join(table.FieldNames, ", ") + " FROM `" + table.Table + "`\" + sqlWhere + sqlOrder\n"
	strMethod += "\terr = t.DB.Select(&data, query)\n"
	strMethod += "\tif err == nil { \n"
	strMethod += "\t\treturn data, len(data), nil\n"
	strMethod += "\t} else if err == sql.ErrNoRows {\n"
	strMethod += "\t\treturn data, 0, nil\n"
	strMethod += "\t} else {\n"
	strMethod += "\t\treturn data, 0, err\n"
	strMethod += "\t}\n"
	strMethod += "}\n"
	if table.SearchDisabled == false {
		strMethod += "\n// " + table.TableAlias + "Pager 分页查询\n"
		strMethod += "func " + table.TableAlias + "Pager(Fields []string, Where string, OrderBy string, Page, PageSize int64) (p sqlxyz.Pager, total int, err error) {\n"
		strMethod += "\tvar this " + table.TableAlias + "\n"
		strMethod += "\treturn this.Pager(Fields, Where, OrderBy, Page, PageSize)\n"
		strMethod += "}\n"
		strMethod += "\n// Pager 根据条件查询一个分页结果集, 条件实例: Pager(\"`State`!=-1\", \"ID DESC\", 1, 50)\n"
		strMethod += "func (t *" + table.TableAlias + ") Pager(Fields []string, Where string, OrderBy string, Page, PageSize int64) (p sqlxyz.Pager, total int, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn p, 0, err\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlWhere string\n"
		strMethod += "\tif Where!=\"\" {\n"
		strMethod += "\t\tsqlWhere=\" Where \"+Where\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlOrderBy string\n"
		strMethod += "\tif OrderBy!=\"\" {\n"
		strMethod += "\t\tsqlOrderBy=\" Order BY \"+OrderBy\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlTotal = \"SELECT count(*) as Total FROM `" + table.Table + "`\" + sqlWhere\n"
		strMethod += "\tvar RecordCount int64\n"
		strMethod += "\terr = t.DB.QueryRow(sqlTotal).Scan(&RecordCount)\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn p, 0, err\n"
		strMethod += "\t}\n"
		strMethod += "\tp = sqlxyz.NewPager(Page, RecordCount, PageSize)\n"
		strMethod += "\tvar Data []" + table.TableAlias + "\n"
		strMethod += "\tif RecordCount > 0 {\n"
		strMethod += "\t\tvar fieldStr string\n"
		strMethod += "\t\tif len(Fields) > 0 {\n"
		strMethod += "\t\t\tfieldStr=strings.Join(Fields, \", \")\n"
		strMethod += "\t\t} else {\n"
		strMethod += "\t\t\tfieldStr=\"" + strings.Join(table.FieldNames, ", ") + "\"\n"
		strMethod += "\t\t}\n"
		strMethod += "\t\tvar query = \"SELECT \" + fieldStr + \" FROM `" + table.Table + "`\" + sqlWhere + sqlOrderBy\n"
		strMethod += "\t\terr = t.DB.Select(&Data, query+\" LIMIT ?, ?\", p.Offset, p.PageSize)\n"
		strMethod += "\t\tif err == sql.ErrNoRows {\n"
		strMethod += "\t\t\treturn p, 0, nil\n"
		strMethod += "\t\t} else if err != nil {\n"
		strMethod += "\t\t\treturn p, 0, err\n"
		strMethod += "\t\t}\n"
		strMethod += "\t\tp.Data = Data\n"

		strMethod += "\t}\n"
		strMethod += "\treturn p, len(Data), nil\n"
		strMethod += "}\n"
	}
	if table.PagingDisabled == false {
		strMethod += "\n// " + table.TableAlias + "Search 条件查询并分页\n"
		strMethod += "func " + table.TableAlias + "Search(Fields []string, Where string, OrderBy string, Page, PageSize int) (data []" + table.TableAlias + ", total int, err error) { \n"
		strMethod += "\tvar this " + table.TableAlias + "\n"
		strMethod += "\treturn this.Search(Fields, Where, OrderBy, Page, PageSize)\n"
		strMethod += "}\n"

		strMethod += "\n// Search 根据条件查询一个分页结果集, 条件实例: Pager(\"`State`!=-1\", \"ID DESC\", 1, 50)\n"
		strMethod += "func (t *" + table.TableAlias + ") Search(Fields []string, Where string, OrderBy string, Page, PageSize int) (data []" + table.TableAlias + ", total int, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn data, 0, err\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlWhere string\n"
		strMethod += "\tif Where!=\"\" {\n"
		strMethod += "\t\tsqlWhere=\" Where \"+Where\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlTotal = \"SELECT count(*) as Total FROM `" + table.Table + "`\" + sqlWhere\n"
		strMethod += "\terr = t.DB.QueryRow(sqlTotal).Scan(&total)\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn data, 0, err\n"
		strMethod += "\t}\n"
		strMethod += "\tif PageSize==-1 {\n"
		strMethod += "\t\tPageSize = total\n"
		strMethod += "\t}\n"
		strMethod += "\tvar Offset int\n"
		strMethod += "\tPage, _, Offset = sqlxyz.SearchPager(Page, total, PageSize)\n"

		strMethod += "\tif total > 0 {\n"
		strMethod += "\t\tvar sqlOrder string\n"
		strMethod += "\t\tif OrderBy !=\"\" {\n"
		strMethod += "\t\t\tsqlOrder = \" ORDER BY \" + OrderBy\n"
		strMethod += "\t\t}\n"
		strMethod += "\t\tvar fieldStr string\n"
		strMethod += "\t\tif len(Fields) > 0 {\n"
		strMethod += "\t\t\tfieldStr=strings.Join(Fields, \", \")\n"
		strMethod += "\t\t} else {\n"
		strMethod += "\t\t\tfieldStr=\"" + strings.Join(table.FieldNames, ", ") + "\"\n"
		strMethod += "\t\t}\n"
		strMethod += "\t\tvar query = \"SELECT \" + fieldStr + \" FROM `" + table.Table + "`\" + sqlWhere + sqlOrder\n"
		strMethod += "\t\terr = t.DB.Select(&data, query+\" LIMIT ?, ?\", Offset, PageSize)\n"
		strMethod += "\t\tif err == sql.ErrNoRows {\n"
		strMethod += "\t\t\treturn data, 0, nil\n"
		strMethod += "\t\t} else if err != nil {\n"
		strMethod += "\t\t\treturn data, 0, err\n"
		strMethod += "\t\t}\n"

		strMethod += "\t}\n"
		strMethod += "\treturn data, total, nil\n"
		strMethod += "}\n"
	}
	strMethod += "\n// FindLimit 根据条件查询一个分页结果集, 条件实例: Pager(\"`State`!=-1\", \"ID DESC\", 1, 50)\n"
	strMethod += "func (t *" + table.TableAlias + ") FindLimit(Fields []string, Where string, OrderBy string, Offset, PageSize int) (data []" + table.TableAlias + ", total int, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn data, 0, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif Where!=\"\" {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+Where\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlOrder string\n"
	strMethod += "\tif OrderBy !=\"\" {\n"
	strMethod += "\t\tsqlOrder = \" ORDER BY \" + OrderBy\n"
	strMethod += "\t}\n"
	strMethod += "\tvar fieldStr string\n"
	strMethod += "\tif len(Fields) > 0 {\n"
	strMethod += "\t\tfieldStr=strings.Join(Fields, \", \")\n"
	strMethod += "\t} else {\n"
	strMethod += "\t\tfieldStr=\"" + strings.Join(table.FieldNames, ", ") + "\"\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"SELECT \" + fieldStr + \" FROM `" + table.Table + "`\" + sqlWhere + sqlOrder\n"
	strMethod += "\terr = t.DB.Select(&data, query+\" LIMIT ?, ?\", Offset, PageSize)\n"
	strMethod += "\tif err == sql.ErrNoRows {\n"
	strMethod += "\t\treturn data, 0, nil\n"
	strMethod += "\t} else if err != nil {\n"
	strMethod += "\t\treturn data, 0, err\n"
	strMethod += "\t}\n"
	strMethod += "\treturn data, total, nil\n"
	strMethod += "}\n"

	strMethod += GetSaveFunc(table)
	strMethod += GetUpdateFunc(table)

	strMethod += GetDeleteFunc(table)
	strMethod += GetSettingFunc(table)
	return strMethod
}

//GetSaveFunc 保存函数代码
func GetSaveFunc(table MysqlTable) string {
	var strMethod string
	var tmpInsert string
	var arrFieldName []string
	for _, fieldname := range table.FieldNames {
		if field, ok := table.Fields[fieldname]; ok {
			if table.IsOnlyPrimary == true && field.IsPrimary && field.IsAutoIncrement {
				//跳过自增长主键更新
				continue
			}
			arrFieldName = append(arrFieldName, field.FieldName)
			if tmpInsert != "" {
				tmpInsert += ", "
			}
			tmpInsert += "t." + field.FieldAlias
		}
	}
	strMethod += "\n// Save 写入一条完整记录\n"
	strMethod += "func (t *" + table.TableAlias + ") Save() (result sql.Result, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn result, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"INSERT INTO `" + table.Table + "` SET " + strings.Join(arrFieldName, "=?, ") + "=?\"\n"
	strMethod += "\tresult, err = t.DB.Exec(query, " + tmpInsert + ")\n"
	strMethod += "\treturn result, err\n"
	strMethod += "}\n"
	return strMethod
}

//GetUpdateFunc 更新函数
func GetUpdateFunc(table MysqlTable) string {
	var strMethod string
	var tmpInsert string
	var arrFieldName []string
	whereSQL, paramInitCode, paramListCode := table.Primary2Code()
	for _, fieldname := range table.FieldNames {
		if field, ok := table.Fields[fieldname]; ok {
			if table.IsOnlyPrimary == true && field.IsPrimary {
				//跳过唯一主键更新
				continue
			}
			if field.FieldName == table.CreatedFieldName || field.FieldName == table.DeletedFieldName {
				continue
			}
			arrFieldName = append(arrFieldName, field.FieldName)
			if tmpInsert != "" {
				tmpInsert += ", "
			}
			tmpInsert += "t." + field.FieldAlias
		}
	}
	strMethod += "\n// Update 更新一条完整记录，如果是单一主键会自动忽略主键值的更新\n"
	strMethod += "func (t *" + table.TableAlias + ") Update(vars ...string) (result sql.Result, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn result, err\n"
	strMethod += "\t}\n"
	if table.HasUpdated {
		if f, ok := table.Fields[table.UpdatedFieldName]; ok {
			if f.FieldType == "time.Time" {
				strMethod += "\tt." + f.FieldAlias + " = time.Now()\n"
			} else if f.FieldType == "typexyz.Timestamp" {
				strMethod += "\tt." + f.FieldAlias + " = typexyz.Now()\n"
			} else {
				strMethod += "\tt." + f.FieldAlias + " = time.Now().Unix()\n"
			}
		}
	}
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) >= 1 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + strings.Join(arrFieldName, "=?, ") + "=?" + "\" + sqlWhere\n"
	strMethod += "\tresult, err = t.DB.Exec(query, " + tmpInsert + ")\n"
	strMethod += "\treturn result, err\n"
	strMethod += "}\n"
	if table.HasPrimary {
		strMethod += "\n// UpdateByPrimaryID 以主键为条件更新一条完整记录，如果是单一主键会自动忽略主键值的更新\n"
		strMethod += "func (t *" + table.TableAlias + ") UpdateByPrimaryID(" + paramInitCode + ") (result sql.Result, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn result, err\n"
		strMethod += "\t}\n"

		if table.HasUpdated {
			if f, ok := table.Fields[table.UpdatedFieldName]; ok {
				if f.FieldType == "time.Time" {
					strMethod += "\tt." + f.FieldAlias + " = time.Now()\n"
				} else if f.FieldType == "typexyz.Timestamp" {
					strMethod += "\tt." + f.FieldAlias + " = typexyz.Now()\n"
				} else {
					strMethod += "\tt." + f.FieldAlias + " = time.Now().Unix()\n"
				}
			}
		}
		strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + strings.Join(arrFieldName, "=?, ") + "=? " + whereSQL + "\"\n"
		strMethod += "\tresult, err = t.DB.Exec(query, " + tmpInsert + "," + paramListCode + ")\n"
		strMethod += "\treturn result, err\n"
		strMethod += "}\n"
	}
	return strMethod
}

//GetDeleteFunc 删除函数
func GetDeleteFunc(table MysqlTable) string {
	var strMethod string
	whereSQL, paramInitCode, paramListCode := table.Primary2Code()
	if table.HasDeleted && table.HasState {
		strMethod += "\n// Delete 标注记录删除状态及时间 State=-1 作为删除状态, 如未指定参数，将对全表数据进行该操作\n"
		strMethod += "func (t *" + table.TableAlias + ") Delete(vars ...string) (result sql.Result, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn result, err\n"
		strMethod += "\t}\n"
		strMethod += "\tvar sqlWhere string\n"
		strMethod += "\tif len(vars) >= 1 {\n"
		strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
		strMethod += "\t}\n"
		f, ok := table.Fields[table.DeletedFieldName]
		if ok {
			if f.FieldType == "time.Time" {
				strMethod += "\tt." + f.FieldAlias + " = time.Now()\n"
			} else if f.FieldType == "typexyz.Timestamp" {
				strMethod += "\tt." + f.FieldAlias + " = typexyz.Now()\n"
			} else {
				strMethod += "\tt." + f.FieldAlias + " = time.Now().Unix()\n"
			}
		}
		if table.HasDeleted {
			strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + table.StateFieldName + "=-1, " + table.DeletedFieldName + "=?\" + sqlWhere\n"
			strMethod += "\tresult, err = t.DB.Exec(query, t." + f.FieldAlias + ")\n"
		} else {
			strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + table.StateFieldName + "=-1\" + sqlWhere\n"
			strMethod += "\tresult, err = t.DB.Exec(query)\n"
		}
		strMethod += "\treturn result, err\n"
		strMethod += "}\n"
	}
	if table.HasPrimary {

		if table.HasUpdated && table.HasState {
			strMethod += "\n// DeleteByPrimaryID 指定主键标注删除一条记录,未指定参数则操作当前记录\n"
			strMethod += "func (t *" + table.TableAlias + ") DeleteByPrimaryID(" + paramInitCode + ") (result sql.Result, err error) { \n"
			strMethod += "\terr = t.Ping()\n"
			strMethod += "\tif err != nil {\n"
			strMethod += "\t\treturn result, err\n"
			strMethod += "\t}\n"
			f, ok := table.Fields[table.DeletedFieldName]
			if ok {
				if f.FieldType == "time.Time" {
					strMethod += "\tt." + f.FieldAlias + " = time.Now()\n"
				} else if f.FieldType == "typexyz.Timestamp" {
					strMethod += "\tt." + f.FieldAlias + " = typexyz.Now()\n"
				} else {
					strMethod += "\tt." + f.FieldAlias + " = time.Now().Unix()\n"
				}
			}
			if table.HasDeleted {
				strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + table.StateFieldName + "=-1, " + f.FieldName + "=? " + whereSQL + "\"\n"
				strMethod += "\tresult, err = t.DB.Exec(query, t." + f.FieldAlias + "," + paramListCode + ")\n"
			} else {
				strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + table.StateFieldName + "=-1 " + whereSQL + "\"\n"
				strMethod += "\tresult, err = t.DB.Exec(query, " + paramListCode + ")\n"
			}
			strMethod += "\treturn result, err\n"
			strMethod += "}\n"
		}

		strMethod += "\n// PhysicallyDeleteByPrimaryID 指定主键物理删除一条记录,未指定参数为删除当前对像\n"
		strMethod += "func (t *" + table.TableAlias + ") PhysicallyDeleteByPrimaryID(" + paramInitCode + ") (result sql.Result, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn result, err\n"
		strMethod += "\t}\n"

		strMethod += "\tvar query = \"DELETE FROM `" + table.Table + "` " + whereSQL + "\"\n"
		strMethod += "\tresult, err = t.DB.Exec(query, " + paramListCode + ")\n"
		strMethod += "\treturn result, err\n"
		strMethod += "}\n"
	}
	strMethod += "\n// PhysicallyDelete 根据条件物理删除一条记录，删除后无法恢复, 如未指定条件等于删除全表数据\n"
	strMethod += "func (t *" + table.TableAlias + ") PhysicallyDelete(vars ...string) (result sql.Result, err error) { \n"
	strMethod += "\terr = t.Ping()\n"
	strMethod += "\tif err != nil {\n"
	strMethod += "\t\treturn result, err\n"
	strMethod += "\t}\n"
	strMethod += "\tvar sqlWhere string\n"
	strMethod += "\tif len(vars) >= 1 {\n"
	strMethod += "\t\tsqlWhere=\" Where \"+vars[0]\n"
	strMethod += "\t}\n"
	strMethod += "\tvar query = \"DELETE FROM `" + table.Table + "`\" + sqlWhere\n"
	strMethod += "\tresult, err = t.DB.Exec(query)\n"
	strMethod += "\treturn result, err\n"
	strMethod += "}\n"
	return strMethod
}

//GetSettingFunc 设置状态函数
func GetSettingFunc(table MysqlTable) string {
	var strMethod string
	whereSQL, paramInitCode, paramListCode := table.Primary2Code()
	s, ok := table.Fields[table.StateFieldName]
	if !ok {
		return ""
	}
	if table.HasUpdated && table.HasState && table.HasPrimary {
		strMethod += "\n// Setting 指定主键设置字段State及Updated字段\n"
		strMethod += "func (t *" + table.TableAlias + ") Setting(" + paramInitCode + ", " + s.FieldAlias + " " + s.FieldType + ") (result sql.Result, err error) { \n"
		strMethod += "\terr = t.Ping()\n"
		strMethod += "\tif err != nil {\n"
		strMethod += "\t\treturn result, err\n"
		strMethod += "\t}\n"
		f, ok := table.Fields[table.UpdatedFieldName]
		if ok {
			if f.FieldType == "time.Time" {
				strMethod += "\tt." + f.FieldAlias + " = time.Now()\n"
			} else if f.FieldType == "typexyz.Timestamp" {
				strMethod += "\tt." + f.FieldAlias + " = typexyz.Now()\n"
			} else {
				strMethod += "\tt." + f.FieldAlias + " = time.Now().Unix()\n"
			}
		}
		strMethod += "\tvar query = \"UPDATE `" + table.Table + "` SET " + s.FieldName + "=?, " + f.FieldName + "=? " + whereSQL + "\"\n"
		strMethod += "\tresult, err = t.DB.Exec(query, " + s.FieldAlias + ", t." + f.FieldAlias + "," + paramListCode + ")\n"
		strMethod += "\treturn result, err\n"
		strMethod += "}\n"
	}
	return strMethod
}
