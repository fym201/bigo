package utl

import (
	"reflect"
)

//将struct的字段和指定的tag对应起来
//structObj为struct的实例指针
//tagType为tag类型，如 `json:name`的tagType为json
func StructFieldMapByTag(structObj interface{}, tagType string) map[string]string {
	fieldsMap := make(map[string]string)
	tp := reflect.TypeOf(structObj).Elem()
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		name := f.Name
		tag := f.Tag.Get(tagType)
		if tag != "" && tag != "-" {
			fieldsMap[tag] = name
		}
	}
	return fieldsMap
}

//将struct的指定的tag和字段对应起来
//structObj为struct的实例指针
//tagType为tag类型，如 `json:name`的tagType为json
func StructTagMapByField(structObj interface{}, tagType string) map[string]string {
	fieldsMap := make(map[string]string)
	tp := reflect.TypeOf(structObj).Elem()
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		name := f.Name
		tag := f.Tag.Get(tagType)
		if tag != "" && tag != "-" {
			fieldsMap[name] = tag
		}
	}
	return fieldsMap
}
