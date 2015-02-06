package utl

import (
	"errors"
	"reflect"
	"strings"
)

//将struct的字段和指定的tag对应起来
//structObj为struct的实例指针
//tagType为tag类型，如 `json:name`的tagType为json
//ignoreEmpty 是否忽略空tag,如果为false，当tag为空的时候会将字段的首字母小写来作为tag
func StructFieldMapByTag(structObj interface{}, tagType string, ignoreEmpty bool) map[string]string {
	fieldsMap := make(map[string]string)
	tp := reflect.TypeOf(structObj).Elem()
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		name := f.Name
		tag := f.Tag.Get(tagType)
		if tag == "-" {
			continue
		}

		if tag == "" {
			if !ignoreEmpty {
				fieldsMap[strings.ToLower(name[0:1])+name[1:]] = name
			}
			continue
		}
		fieldsMap[tag] = name

	}
	return fieldsMap
}

//将struct的指定的tag和字段对应起来
//structPtr为struct的实例指针
//tagType为tag类型，如 `json:name`的tagType为json
//ignoreEmpty 是否忽略空tag,如果为false，当tag为空的时候会将字段的首字母小写来作为tag
func StructTagMapByField(structPtr interface{}, tagType string, ignoreEmpty bool) map[string]string {
	fieldsMap := make(map[string]string)
	tp := reflect.TypeOf(structPtr).Elem()
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		name := f.Name
		tag := f.Tag.Get(tagType)
		if tag == "-" {
			continue
		}

		if tag == "" {
			if !ignoreEmpty {
				fieldsMap[name] = strings.ToLower(name[0:1]) + name[1:]
			}
			continue
		}
		fieldsMap[name] = tag
	}
	return fieldsMap
}

//将struct的指定字段转到map中
//structPtr为struct的实例指针
//mapObj为要存放数据的map
//feilds为[]string类型代表要取的字段
//feilds为map[string]string类型代表字段映射 {"structFeild":"mapKey"}, 如果mapKey为空则使用structFeild做为key
func StructToMap(structPtr interface{}, mapObj map[string]interface{}, feilds interface{}) {
	tp := reflect.ValueOf(structPtr).Elem()
	if reflect.TypeOf(feilds).Kind() == reflect.Map {
		var mapping map[string]string = feilds.(map[string]string)
		for k, v := range mapping {
			if v != "" {
				mapObj[v] = tp.FieldByName(k).Interface()
			} else {
				mapObj[k] = tp.FieldByName(k).Interface()
			}
		}
	} else {
		fds := feilds.([]string)
		for _, v := range fds {
			mapObj[v] = tp.FieldByName(v).Interface()
		}
	}
}

//将map根据指定的tag转为struct
//structPtr为struct的实例指针
//mapObj为要存放数据的map
//tagType 为struct中要用的tag
func MapToStruct(mapObj interface{}, structPtr interface{}, tagType string) error {
	if reflect.TypeOf(mapObj).Kind() != reflect.Map {
		return errors.New("mapObj is not a map")
	}
	m := mapObj.(map[string]interface{})
	vl := reflect.ValueOf(structPtr).Elem()
	tp := reflect.TypeOf(structPtr).Elem()

	if vl.Kind() != reflect.Ptr || vl.IsNil() {
		return errors.New("structPtr is not a struct instance pointer")
	}

	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		name := f.Name
		tag := f.Tag.Get(tagType)

		if tag == "-" {
			continue
		}
		if tag == "" {
			tag = strings.ToLower(name[0:1]) + name[1:]
		}
		vf := vl.Field(i)
		if vf.CanSet() {
			if v, ok := m[tag]; ok {
				vf.Set(reflect.ValueOf(v))
			}

		}
	}
	return nil
}
