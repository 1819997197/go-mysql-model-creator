package code

import (
	"github.com/go-ini/ini"
)

// IniCfg Ini 结构体
type IniCfg struct {
	Cfg *ini.File
}

// IniConfigInit 初始化一个配置文件
func IniConfigInit(confPath string) (cfg *ini.File, err error) {
	cfg, err = ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, confPath)
	return
}

// IniFileLoad 加载一个ini配置
func IniFileLoad(confPath string) (c IniCfg, err error) {
	cfg, err := ini.LoadSources(ini.LoadOptions{AllowBooleanKeys: true}, confPath)
	if err != nil {
		return c, err
	}
	c.Cfg = cfg
	return c, nil
}

// MysqlCfg 初始化一个mysql的Ini配置
func (c *IniCfg) MysqlCfg(SectionName string, v interface{}) (err error) {
	err = c.Cfg.Section(SectionName).MapTo(v)
	return
}
