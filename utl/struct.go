package utl

import (
	"errors"
	"reflect"
	"strings"
)

// 将struct的字段和指定的tag对应起来
// @param structObj struct的实例指针
// @param tagType tag类型，如 `json:name`的tagType为json
// @param ignoreEmpty 是否忽略空tag,如果为false，当tag为空的时候会将字段的首字母小写来作为tag
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

// 将struct的指定的tag和字段对应起来
// @param structPtr struct的实例指针
// @param tagType tag类型，如 `json:name`的tagType为json
// @param ignoreEmpty 是否忽略空tag,如果为false，当tag为空的时候会将字段的首字母小写来作为tag
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

// 将struct的指定字段转到map中
// @param structPtr struct的实例指针
// @param feilds []string类型代表要取的字段,map[string]string类型代表字段映射
// {"structFeild":"mapKey"}, 如果mapKey为空则使用structFeild做为key
// @param mapObj 要存放数据的map,如果传入的话会将数据放到这个map里
func StructToMap(structPtr interface{}, feilds interface{}, mapObj ...map[string]interface{}) map[string]interface{} {
	var mp map[string]interface{}
	if len(mapObj) > 0 {
		mp = mapObj[0]
	} else {
		mp = make(map[string]interface{})
	}
	tp := reflect.ValueOf(structPtr).Elem()
	if reflect.TypeOf(feilds).Kind() == reflect.Map {
		var mapping map[string]string = feilds.(map[string]string)
		for k, v := range mapping {
			if v != "" {
				mp[v] = tp.FieldByName(k).Interface()
			} else {
				mp[k] = tp.FieldByName(k).Interface()
			}
		}
	} else {
		fds := feilds.([]string)
		for _, v := range fds {
			mp[v] = tp.FieldByName(v).Interface()
		}
	}
	return mp
}

// 将struct转为map
// @param structPtr struct的实例指针
// @param v 有四种情况：
// 1. 不传,将把struct的字段名首字母小写做为map的key
// 2. 包含string类型,将把struct字段中对应的tag做为map的key
// 3. 包含(map[string]interface{})类型, 将把数据存到这个map中
// 4. 包含bool类型,为true的话struct中字段值为空的字段不会加入map
func StrutToMapByTag(structPtr interface{}, v ...interface{}) map[string]interface{} {
	var tag string
	var mp map[string]interface{}
	omitEmpty := false

	for _, o := range v {
		switch o.(type) {
		case string:
			tag = o.(string)
		case bool:
			omitEmpty = o.(bool)
		case map[string]interface{}:
			mp = o.(map[string]interface{})
		}
	}

	vl := reflect.ValueOf(structPtr).Elem()
	tp := reflect.TypeOf(structPtr).Elem()
	for i := 0; i < tp.NumField(); i++ {
		f := tp.Field(i)
		var key string
		if tag == "" {
			key = strings.ToLower(f.Name[0:1]) + f.Name[1:]
		} else {
			key := f.Tag.Get(tag)
			if key == "-" {
				continue
			}
			if key == "" {
				key = strings.ToLower(f.Name[0:1]) + f.Name[1:]
			}
		}
		value := vl.FieldByName(key)
		if omitEmpty && hIsEmptyValue(value, true, true) {
			continue
		}
		mp[key] = value.Interface()
	}
	return mp
}

// 将map根据指定的tag转为struct
// @param structPtr struct的实例指针
// @param mapObj 为要存放数据的map
// @param tagType 为struct中要用的tag
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

// 将map中的key转为struct的另一个tag表示
// @param obj map实例
// @param stp truct类型
// @param srcTag 原tag名,如json
// @param dstTag 目标tag名,如bson
func StructTagConvert(obj map[string]interface{}, stp reflect.Type, srcTag, dstTag string) map[string]interface{} {
	ret := make(map[string]interface{})

	for i := 0; i < stp.NumField(); i++ {
		f := stp.Field(i)
		tagA := f.Tag.Get(srcTag)
		if tagA == "-" {
			continue
		}
		if tagA == "" {
			tagA = strings.ToLower(f.Name[0:1]) + f.Name[1:]
		}

		if v, ok := obj[tagA]; ok {
			tagB := f.Tag.Get(dstTag)
			if tagB == "-" {
				continue
			}
			if tagB == "" {
				tagB = tagA
			}
			ret[tagB] = v
		}
	}
	return ret
}

// 检查给定的值是否为空值
func hIsEmptyValue(v reflect.Value, deref, checkStruct bool) bool {
	switch v.Kind() {
	case reflect.Invalid:
		return true
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		if deref {
			if v.IsNil() {
				return true
			}
			return hIsEmptyValue(v.Elem(), checkStruct, deref)
		} else {
			return v.IsNil()
		}
	case reflect.Struct:
		if !checkStruct {
			return false
		}
		// return true if all fields are empty. else return false.
		// we cannot use equality check, because some fields may be maps/slices/etc
		// and consequently the structs are not comparable.
		// return v.Interface() == reflect.Zero(v.Type()).Interface()
		for i, n := 0, v.NumField(); i < n; i++ {
			if !hIsEmptyValue(v.Field(i), checkStruct, deref) {
				return false
			}
		}
		return true
	}
	return false
}
