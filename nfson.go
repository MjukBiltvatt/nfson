package nfson

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/valyala/fastjson"
)

const TagName = "nfson"
const divider = "."

var timeFormats = []string{"", ""}

// Map maps the json from parsedJSONbytes or data to object obj, see readme for further details
func Map(parsedJSONbytes *fastjson.Value, obj interface{}, timeLoc *time.Location, subTagName string, recurseSubTag bool, baseTags ...string) {
	v := reflect.ValueOf(obj).Elem()

	tempTagName := TagName + subTagName

	//Loop through all struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)

		//Skip the field if it cannot be set
		if !field.IsValid() || !field.CanSet() {
			continue
		}

		//Get the tag of the field
		tag := v.Type().Field(i).Tag.Get(tempTagName)

		//Split tag
		tags := SplitTag(tag)
		//Append baseTags, mainly used to simplify mapping of nested structs
		tags = append(baseTags, tags...)

		if !parsedJSONbytes.Exists(tags...) {
			continue
		}

		//Set the struct field value depending on the underlying type
		if field.Kind() != reflect.Pointer {
			// If underlying type is neither pointer nor struct
			switch field.Interface().(type) {
			case string:
				field.SetString(string(parsedJSONbytes.Get(tags...).GetStringBytes()))
				continue
			case int, int8, int16, int32, int64:
				field.SetInt(parsedJSONbytes.GetInt64(tags...))
				continue
			case uint, uint8, uint16, uint32, uint64:
				field.SetUint(parsedJSONbytes.GetUint64(tags...))
				continue
			case float32, float64:
				field.SetFloat(parsedJSONbytes.GetFloat64(tags...))
				continue
			case bool:
				field.SetBool(parsedJSONbytes.GetBool(tags...))
				continue
			case time.Time:
				field.Set(reflect.ValueOf(jtime(parsedJSONbytes, timeLoc, tags...)))
				continue
			}

		} else {
			// If underlying type is pointer
			if parsedJSONbytes.Get(tags...).Type() == fastjson.TypeNull {
				continue
			}
			switch field.Interface().(type) {
			case *time.Time:
				t := jtime(parsedJSONbytes, timeLoc, tags...)
				//Only set field if time is not zero
				if field.IsNil() && !t.IsZero() {
					//Nil pointer
					field.Set(reflect.ValueOf(&t))
				} else if !t.IsZero() {
					//Value pointer
					field.Elem().Set(reflect.ValueOf(t))
				}
				continue
			case *string:
				str := string(parsedJSONbytes.GetStringBytes(tags...))
				field.Set(reflect.ValueOf(&str))
				continue
			case *int:
				int := int(parsedJSONbytes.GetInt64(tags...))
				field.Set(reflect.ValueOf(&int))
				continue
			case *int8:
				int8 := int8(parsedJSONbytes.GetInt64(tags...))
				field.Set(reflect.ValueOf(&int8))
				continue
			case *int16:
				int16 := int16(parsedJSONbytes.GetInt64(tags...))
				field.Set(reflect.ValueOf(&int16))
				continue
			case *int32:
				int32 := int32(parsedJSONbytes.GetInt64(tags...))
				field.Set(reflect.ValueOf(&int32))
				continue
			case *int64:
				int64 := parsedJSONbytes.GetInt64(tags...)
				field.Set(reflect.ValueOf(&int64))
				continue
			case *uint:
				uint := uint(parsedJSONbytes.GetUint64(tags...))
				field.Set(reflect.ValueOf(&uint))
				continue
			case *uint8:
				uint8 := uint8(parsedJSONbytes.GetUint64(tags...))
				field.Set(reflect.ValueOf(&uint8))
				continue
			case *uint16:
				uint16 := uint16(parsedJSONbytes.GetUint64(tags...))
				field.Set(reflect.ValueOf(&uint16))
				continue
			case *uint32:
				uint32 := uint32(parsedJSONbytes.GetUint64(tags...))
				field.Set(reflect.ValueOf(&uint32))
				continue
			case *uint64:
				uint64 := parsedJSONbytes.GetUint64(tags...)
				field.Set(reflect.ValueOf(&uint64))
				continue
			case *float32:
				float32 := float32(parsedJSONbytes.GetFloat64(tags...))
				field.Set(reflect.ValueOf(&float32))
				continue
			case *float64:
				float64 := parsedJSONbytes.GetFloat64(tags...)
				field.Set(reflect.ValueOf(&float64))
				continue
			case *bool:
				bool := parsedJSONbytes.GetBool(tags...)
				field.Set(reflect.ValueOf(&bool))
				continue
			}
		}

		if field.Kind() == reflect.Struct {
			//Map nested struct
			if recurseSubTag {
				Map(parsedJSONbytes, field.Addr().Interface(), timeLoc, subTagName, true, tags...)
			} else {
				Map(parsedJSONbytes, field.Addr().Interface(), timeLoc, "", false, tags...)
			}
			continue
		} else if field.Kind() == reflect.Pointer && field.Elem().Kind() == reflect.Struct {
			//Map nested pointer to struct
			if recurseSubTag {
				Map(parsedJSONbytes, field.Interface(), timeLoc, subTagName, true, tags...)
			} else {
				Map(parsedJSONbytes, field.Interface(), timeLoc, "", false, tags...)
			}
			continue
		}
	}
}

func SplitTag(tag string) []string {
	return strings.Split(tag, divider)
}

func jtimeE(parsedJSONbytes *fastjson.Value, loc *time.Location, tags ...string) (time.Time, error) {
	data := parsedJSONbytes.Get(tags...).String()

	//Attempt to parse as timestamp in format MM/dd/yyyy HH:mm:ss
	if match, err := regexp.MatchString(`^\d{2}\/\d{2}\/\d{4} \d{2}:\d{2}:\d{2}$`, data); err != nil {
		return time.Time{}, err
	} else if match {
		return time.ParseInLocation("01/02/2006 15:04:05", data, loc)
	}

	//Attempt to parse as date in format MM/dd/yyyy
	if match, err := regexp.MatchString(`^\d{2}\/\d{2}\/\d{4}$`, data); err != nil {
		return time.Time{}, err
	} else if match {
		return time.ParseInLocation("01/02/2006", data, loc)
	}

	//Attempt to parse as timestamp in format yyyy-MM-dd HH:mm:ss
	if match, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}$`, data); err != nil {
		return time.Time{}, err
	} else if match {
		return time.ParseInLocation("2006-01-02 15:04:05", data, loc)
	}

	//Attempt to parse as date in format yyyy-MM-dd
	if match, err := regexp.MatchString(`^\d{4}-\d{2}-\d{2}$`, data); err != nil {
		return time.Time{}, err
	} else if match {
		return time.ParseInLocation("2006-01-02", data, loc)
	}

	//Attempt to parse as date in format yyyy-MM
	if match, err := regexp.MatchString(`^\d{4}-\d{2}$`, data); err != nil {
		return time.Time{}, err
	} else if match {
		return time.ParseInLocation("2006-01", data, loc)
	}

	return time.Time{}, fmt.Errorf("failed to parse \"%v\" as type \"%s\"", data, "time.Time")
}

func jtime(parsedJSONbytes *fastjson.Value, loc *time.Location, tags ...string) time.Time {
	time, err := jtimeE(parsedJSONbytes, loc, tags...)
	// TODO: handle error
	if err != nil {
		fmt.Println("error:", err)
	}
	return time
}
