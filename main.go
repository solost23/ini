package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"strconv"
	"strings"
)

// ini配置文件解析

// MysqlConfig MySQL配置结构体
type MysqlConfig struct {
	Address  string `ini:"address"`
	Port     int    `ini:"port"`
	Username string `ini:"username"`
	Password string `ini:"password"`
}

// RedisConfig Redis配置结构体
type RedisConfig struct {
	Host     string `ini:"host"`
	Port     int    `ini:"post"`
	Password string `ini:"password"`
	Database int    `ini:"database"`
	Test     bool   `ini:"test"`
}

// Config
type Config struct {
	MysqlConfig `ini:"mysql"`
	RedisConfig `ini:"redis"`
}

func loadIni(fileName string, v interface{}) (err error) {
	// 0.参数校验，
	// 0.1传进来的v参数必须是指针类型，因为要在函数中对其赋值
	t := reflect.TypeOf(v)
	//fmt.Println(t, t.Kind())
	if t.Kind() != reflect.Ptr {
		err = errors.New("v param should be a pointer") // 创建一个错误
		return
	}
	// 0.2传进来的v参数必须是结构体类型，因为配置文件中各种键值对需要赋值给结构体的字段
	if t.Elem().Kind() != reflect.Struct {
		err = errors.New("v param should be a struct") // 创建一个错误
		return
	}
	// 1.读文件得到字节类型的数据
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		err = errors.New("file open failed")
		return
	}
	//string(b) // 将字节类型的文件内容转换成字符串
	lineSlice := strings.Split(string(b), "\r\n")
	//fmt.Printf("%#v \n", lineSlice)
	// 2.一行一行地读数据
	var structName string
	for index, line := range lineSlice {
		//fmt.Printf("第%d行 ", index + 1)
		// 去掉每个字符串首尾的空格
		line = strings.TrimSpace(line)
		// 如果是空行忽略
		if len(line) == 0 {
			continue
		}
		// 2.1如果是注释就跳过
		//if line[0] == '#' || line[0] == ';' {
		//	continue
		//}
		if strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		// 2.2如果是方括号开头的就表示是节（section）
		if strings.HasPrefix(line, "[") {
			if line[0] != '[' || line[len(line)-1] != ']' {
				err = fmt.Errorf("line:%d, syntax error", index+1)
				return
			}
			// 把这一行首尾[]去掉，取到中间的内容把空格去掉拿到内容
			sectionName := strings.TrimSpace(line[1 : len(line)-1])
			if len(sectionName) == 0 {
				err = fmt.Errorf("line:%d, syntax error", index+1)
				return
			}
			// 根据字符串selectionName去结构体v里面根据反射找到对应的结构体
			for i := 0; i < t.Elem().NumField(); i++ {
				field := t.Elem().Field(i)
				if field.Tag.Get("ini") == sectionName {
					// 说明找到了对应的嵌套结构体，把字段名记下来
					structName = field.Name
					//fmt.Printf("找到%s对应的嵌套结构体%s \n", sectionName, structName)
					break
				}
			}
		} else {
			// 2.3如果不是，开头就是=分隔的键值对

			// 1.以等号分割这一行，等号左边是k,右边是v。
			if strings.Index(line, "=") == -1 || strings.HasPrefix(line, "=") {
				err = fmt.Errorf("line:%d syntax err", index+1)
				return
			}

			idx := strings.Index(line, "=")
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			// 2.根据structName去v里面把对应的嵌套结构体取出来。
			value := reflect.ValueOf(v)

			sValue := value.Elem().FieldByName(structName) // 拿到嵌套结构体的值信息
			sType := sValue.Type()                         // 拿到嵌套结构体的类型信息

			if sType.Kind() != reflect.Struct {
				err = fmt.Errorf("v 中的%s字段不是一个结构体 \n", structName)
				return err
			}

			var fieldName string
			// 3.遍历嵌套结构体的每一个字段，判断tag是不是等于k,
			for i := 0; i < sValue.NumField(); i++ {
				field := sType.Field(i) // tag信息存储在类型信息中的
				if field.Tag.Get("ini") == key {
					// 找到对应的字段
					fieldName = field.Name
					break
				}
			}
			// 4.如果k == tag,给这个字段赋值。
			// 4.1根据fieldName取出这个字段。
			if len(fieldName) == 0 {
				// 在结构体中找不到对应的字符
				continue
			}
			fileObj := sValue.FieldByName(fieldName)
			// 4.2对其赋值
			//fmt.Println(fieldName, fileObj.Type().Kind())
			switch fileObj.Type().Kind() {
			case reflect.String:
				fileObj.SetString(val)
			case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
				var valInt int64
				valInt, err = strconv.ParseInt(val, 10, 64)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", index+1)
					return
				}
				fileObj.SetInt(valInt)
			case reflect.Bool:
				var valBool bool
				valBool, err = strconv.ParseBool(val)
				if err != nil {
					err = fmt.Errorf("line:%d value type error", index+1)
					return
				}
				fileObj.SetBool(valBool)
			}
		}

	}

	return
}

func main() {
	var cfg Config

	err := loadIni("./conf.ini", &cfg)
	if err != nil {
		fmt.Printf("load ini failed, err:%v \n", err)
	}

	fmt.Println("mysql", cfg.Address, cfg.MysqlConfig.Port, cfg.Username, cfg.MysqlConfig.Password)
	fmt.Println("redis", cfg.Host, cfg.RedisConfig.Port, cfg.RedisConfig.Password, cfg.Database, cfg.Test)
}
