package djson

import (
	"fmt"
	"testing"
)

func TestFileUnmarshal(t *testing.T){
	scripts := []Script{
			Script{Tag:"GO:", ScriptFunc:CallBack, Args:"test"},
	}
	config := Config{}
	if err := FileUnmarshal("./test.json", &config, scripts); err != nil{
		panic(err)
	}
	fmt.Println(config)
}


func CallBack(content string, args interface{}) (interface{}, error){
	r := fmt.Sprintf("%s%s", content, args)
	return r, nil
}


type BaseField struct{
	Name			string		`json:"name" require:"true"`
	Title 			string 		`json:"title"`
	Info 			string		`json:"info"`
	Multiple 		bool		`json:"multiple"`
	Default			interface{}	`json:"default"`
}


type Field struct {
	BaseField
	Memo			string	`json:"memo"`
	Show			bool	`json:"show" default:"true"`
}


type SearchField 	struct {
	BaseField
	Values			[]string	`json:"values"`
}

type Kv struct {
	KeyFields   []string `json:"key_fields"`   // 主键（必填）
	KeySep      string   `json:"value_sep"`    // 多关键字段分隔符（默认_)
}

type Attr struct {
	Attr1		string		`json:"attr1"`
	Attr2		string		`json:"attr2"`
}


//自定义模型整体配置对象
type Config struct {
	Name				string						`json:"-"`
	ConnName			string						`json:"conn_name" default:"default"`
	DbName				string						`json:"db_name"`
	Table				string						`json:"table"`
	AutoIncrement 		bool						`json:"auto_increment" default:"true"`
	Fields				[]Field						`json:"fields"`
	SearchFields 		[]SearchField				`json:"search_fields"`
	Kvs					map[string]Kv				`json:"kvs"`
	Attr				Attr						`json:"javascript"`
}