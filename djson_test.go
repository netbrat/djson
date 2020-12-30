package djson

import (
	"fmt"
	"strings"
	"testing"
)


type Cities struct {
	City	[]string	`json:"city"`
}

type Province struct {
	Name	string		`json:"name" require:"true"`
	Cities	Cities		`json:"cities"`
}

type Config	struct {
	Name		string		`json:"name" default:"中国"`
	Capital		string		`json:"capital" default:"北京"`
	Code		string		`json:"code" default:"CN"`
	Nation		int			`json:"nation" default:"56"`
	Provinces 	[]Province	`json:"provinces"`
}



func TestUnmarshal(t *testing.T){
	var testJson = `
	{
		"name": "中国",
		"capital": "",
		"nation": 0,
		"provinces": [{
			"name": "黑龙江",
			"cities": {
				"city": ["哈尔滨", "大庆"]
			}
		}, {
			"name": "广东",
			"cities": {
				"city": ["广州", "深圳", "珠海"]
			}
		}, {
			"name": "台湾",
			"cities": {
				"city": ["台北", "高雄"]
			}
		}, {
			"name": "GO:新疆ZZQ",
			"cities": {
				"city": ["乌鲁木齐"]
			}
		}]
	}	
	`
	scripts := []Script{
		{
			Tag:"GO:",
			ScriptFunc:func(content string, args interface{})(interface{}, error){
				r := strings.ReplaceAll(content,"ZZQ","自治区") + args.(string)
				return r, nil
			},
			Args:"(脚本测试)",
		},
	}
	config := Config{}
	if err := Unmarshal(testJson, &config, scripts); err != nil{
		panic(err)
	}
	fmt.Println(config)
}




