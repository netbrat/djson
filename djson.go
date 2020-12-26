// Copyright @ 2020 netbrat	<netbrat@qq.com

// Package djson in Go.
package djson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/cast"
	"io/ioutil"
	"reflect"
	"strings"
)

// 脚本回调函数定义
type ScriptFunc func(content string, args interface{}) (interface{}, error)

// 脚本对象
type Script struct {
	Tag			string
	ScriptFunc 	ScriptFunc
	Args		interface{}
}


// json文件内容解析到struct
func FileUnmarshal(file string, obj interface{}, scripts []Script) error {
	if jsData, err := ioutil.ReadFile(file); err != nil{
		return err
	}else{
		jsData = bytes.TrimPrefix(jsData, []byte("\xef\xbb\xbf"))
		var data map[string]interface{}
		if err := json.Unmarshal(jsData, &data); err != nil{
			return err
		}
		return reflectSetStruct(data, obj, scripts)
	}
}


// json字符串或[]byte解析到struct
func Unmarshal(js interface{}, obj interface{}, scripts []Script) error {
	var jsData []byte
	var data map[string]interface{}
	switch reflect.TypeOf(js).Kind() {
	case reflect.String:
		jsData = []byte(js.(string))
	case reflect.Array, reflect.Slice:
	default:
		return fmt.Errorf("wrong json data source type")
	}
	if err := json.Unmarshal(jsData, &data); err != nil {
		return err
	}
	return reflectSetStruct(data, obj, scripts)
}

// 通过反射设置struct字段值
func reflectSetStruct(data map[string]interface{}, obj interface{}, scripts []Script) (err error){
	t := reflect.TypeOf(obj).Elem()
	v := reflect.ValueOf(obj).Elem()
	sfName := ""
	jsonTag := ""
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		sv := v.FieldByName(sf.Name)
		sfName = sf.Name
		//tag
		tag := strings.Split(sf.Tag.Get("json"),",") 	//json tag
		defValue := sf.Tag.Get("default") 					//default tag
		isRequire := cast.ToBool(sf.Tag.Get("require")) 	//require tag
		if tag[0] == "" { tag[0] = sf.Name }
		if tag[0] == "-" { continue }
		jsonTag = tag[0]
		//value
		var value interface{} = nil
		if data != nil {
			value, _ = data[tag[0]]
			// check and run script
			if value, err = runScript(value, scripts); err != nil{
				err = fmt.Errorf("%s(%s) run script error：%s",sf.Name,jsonTag, err.Error())
				return
			}
		}

		if isRequire{
			if value == nil || cast.ToString(value) == ""{
				err = fmt.Errorf("%s(%s) is require", sf.Name, jsonTag)
				return
			}
		}

		if sf.Anonymous { // 匿名字段
			value = data
		}
		if err = reflectSetValue(sf.Type, sv, value, defValue, scripts); err != nil {
			err = fmt.Errorf("%s(%s) ：%s", sfName, jsonTag, err.Error())
			return
		}
	}
	defer func(){
		if r := recover(); r != nil{
			err =  fmt.Errorf(fmt.Sprintf("%s(%s) : %s",sfName,jsonTag, r))
		}
	}()

	return
}

// 通过反射设置值
func reflectSetValue(rt reflect.Type, rv reflect.Value, value interface{}, defValue string, scripts []Script) (err error){
	if value == nil {
		if defValue == ""{
			return
		}else{
			value = defValue
		}
	}
	isPtr := false
	if rt.Kind() == reflect.Ptr {
		isPtr = true
		rt = rt.Elem()
	}
	switch rt.Kind() {
	case reflect.Interface:
		rv.Set(reflect.ValueOf(value))

	case reflect.String:
		tempV := cast.ToString(value)
		if tempV == "" && defValue != "" {
			tempV = defValue
		}
		if isPtr{
			rv.Set(reflect.ValueOf(&tempV))
		}else {
			rv.SetString(tempV)
		}

	case reflect.Bool:
		if cast.ToString(value) == "" && defValue != ""{
			value = defValue
		}
		tempV := cast.ToBool(value)
		if isPtr {
			rv.Set(reflect.ValueOf(&tempV))
		}else {
			rv.SetBool(tempV)
		}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: //整型
		tempV := cast.ToInt64(value)
		if tempV == 0 && defValue != "" {
			tempV = cast.ToInt64(defValue)
		}
		if isPtr{
			rv.Set(reflect.ValueOf(&tempV))
		}else {
			rv.SetInt(tempV)
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		tempV := cast.ToUint64(value)
		if tempV == 0 && defValue != "" {
			tempV = cast.ToUint64(defValue)
		}
		if isPtr{
			rv.Set(reflect.ValueOf(&tempV))
		}else {
			rv.SetUint(tempV)
		}

	case reflect.Float32, reflect.Float64:
		tempV := cast.ToFloat64(value)
		if tempV == 0 && defValue != "" {
			tempV = cast.ToFloat64(defValue)
		}
		if isPtr{
			rv.Set(reflect.ValueOf(&tempV))
		}else {
			rv.SetFloat(tempV)
		}

	case reflect.Struct:
		obj := reflect.New(rt)
		if err = reflectSetStruct(value.(map[string]interface{}), obj.Interface(), scripts); err != nil{
			return
		}
		if isPtr{
			rv.Set(obj)
		}else{
			rv.Set(obj.Elem())
		}

	case reflect.Slice, reflect.Array:
		kind := rt.Elem().Kind()
		arrayValue := value.([]interface{})
		es := make([]reflect.Value,0)
		for _, v := range arrayValue {
			obj := reflect.New(rt.Elem())
			if kind == reflect.Struct || kind == reflect.Map {
				if err = reflectSetStruct(v.(map[string]interface{}), obj.Interface(), scripts); err != nil {
					return
				}
			}else{
				if err = reflectSetValue(rt.Elem(), obj.Elem(), v,"", scripts); err != nil{
					return
				}
			}
			es = append(es, obj.Elem())
		}
		rv.Set(reflect.Append(rv, es...))

	case reflect.Map: //map
		res := reflect.MakeMap(rt)
		mapV := value.(map[string]interface{})
		for key, v := range mapV{
			k := reflect.ValueOf(key)
			obj := reflect.New(rt.Elem())
			if err = reflectSetValue(rt.Elem(), obj.Elem(), v, "", scripts); err != nil{
				return
			}
			res.SetMapIndex(k, obj.Elem())
		}
		rv.Set(res)
	}
	return
}

// 检查并运行回调脚本
func runScript(value interface{}, scripts []Script) (v interface{}, err error){
	v = value
	if value == nil || reflect.TypeOf(value).Kind() != reflect.String {
		return
	}
	if scripts == nil{
		return
	}
	vString := value.(string)
	for _, script := range scripts {
		if  script.Tag == "" || script.ScriptFunc == nil { continue }
		tagLen := len(script.Tag)
		if len(vString)> tagLen && strings.ToUpper(vString[:tagLen]) == strings.ToUpper(script.Tag) {
			v, err = script.ScriptFunc(vString[tagLen:], script.Args)
			return
		}
	}
	return
}