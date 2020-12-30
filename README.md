# djson
## 动态json解析

### 支持：

**1、默认值设定(tag: default)，以下情况会自动在解析后的struct中初始化默认**：
        
> 当json数据某项值不存在
> 
> 当数字型为0时
> 
> 当字符型为""时


**2、必填字段验证(tag: require)**

**3、回调函数**

> 当json数据项为字符串且内容前缀为设定的回调函数Tag相同时，则自动调用回调函数对内容进行处理（支持外部动态脚本）
    

### 示例

```


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

func main(){
    
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
    //定义脚本回调
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

    //如果不进行脚本回调解晰，使用
    // Unmarshal(testJson, &config, nil)
	if err := Unmarshal(testJson, &config, scripts); err != nil{
		panic(err)
	}
	fmt.Println(config)
}

//返回：
//{中国 北京 CN 56 [{黑龙江 {[哈尔滨 大庆]}} {广东 {[广州 深圳 珠海]}} {台湾 {[台北 高雄]}} {新疆自治区(脚本测试) {[乌鲁木齐]}}]}

```

### 直接读取json文件进行解析


```
if err := FileUnmarshal("./config.json", &config, scripts); err!= nil{
    panic(err)
}
fmt.Println(config)
```


### 使用其他脚本语言回调方法

javascript示例

```

// 导入第三方javascript引擎包
import "github.com/dop251/goja"


var testJson = `
{
    "test" : "JS: (attr.sex == 'M' ? '男' : '女') + '(年龄：' + attr.age + ')'"
}
`

type Config struct {
    Test    string      `json:"test"`
}

func main(){    

    //定义脚本回调
    scripts := []Script{
		{
			Tag:"JS:",
			ScriptFunc:func(content string, args interface{})(interface{}, error){
			    vm := goja.New()
			    vm.Set("attr", args)
			    if value, err := vm.RunString(content); err != nil{
			        panic(err)
			    } else {
					return value, nil
				}
			},
			Args: map[string]interface{}{ "sex": "F", "age": 20 }
		},
	}    

	config := Config{}

	if err := Unmarshal(testJson, &config, scripts); err != nil{
		panic(err)
	}
	fmt.Println(config)
}

// 返回：{女(年龄：20)}

```
