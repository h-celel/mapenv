package mapenv

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	mapEnvTagName string = "mpe"
)

// Decode decode current environmental variables into an output structure.
// Output must be a pointer to a struct.
func Decode(v interface{}) error {
	val := reflect.ValueOf(v)
	t := reflect.TypeOf(v)

	if val.Kind() != reflect.Ptr {
		return newDecodeError("must decode to pointer", "", nil)
	}

	for t.Kind() == reflect.Ptr {
		t = t.Elem()

		if val.IsNil() {
			val.Set(reflect.New(t))
		}

		val = val.Elem()
	}

	if t.Kind() != reflect.Struct {
		return newDecodeError(fmt.Sprintf("cannot decode into value of type: %s", t.String()), "", nil)
	}

	newVal := reflect.New(t)

	for i := 0; i < newVal.Elem().NumField(); i++ {
		fTyp := t.Field(i)
		isUnexported := fTyp.PkgPath != ""
		if isUnexported {
			continue
		}

		var s string
		var tag string
		var ok bool

		fieldTags := getFieldTags(fTyp)
		for _, tag = range fieldTags {
			if s, ok = os.LookupEnv(tag); ok {
				break
			}
		}
		if len(s) == 0 {
			continue
		}

		fVal := newVal.Elem().Field(i)
		err := decodeValue(s, fVal.Addr())
		if err != nil {
			return newDecodeError(fmt.Sprintf("unable to decode value in field '%s'", tag), tag, err)
		}
	}

	val.Set(newVal.Elem())

	return nil
}

// decodeValue decodes a string variable as a value. Base types are parsed using `strconv`. Maps, structs, arrays and
// slices are decoded as json objects using standard json unmarshaling. Channels and functions are skipped, as they're
// not supported.
func decodeValue(s string, v reflect.Value) error {
	switch v.Elem().Kind() {
	case reflect.String:
		v.Elem().SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.Elem().SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.Elem().SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.Elem().SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.Elem().SetFloat(f)
	case reflect.Complex64, reflect.Complex128:
		f, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return err
		}
		v.Elem().SetComplex(f)
	case reflect.Map, reflect.Struct, reflect.Array, reflect.Slice:
		i := v.Interface()
		switch i.(type) {
		case *time.Time:
			t, err := parseTime(s)
			if err != nil {
				return err
			}
			v.Elem().Set(reflect.ValueOf(t))
		default:
			err := json.Unmarshal([]byte(s), i)
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return decodeValue(s, v.Elem())
	case reflect.Chan, reflect.Func:
	default:
		return fmt.Errorf("unsupported field kind: %s", v.Elem().Kind().String())
	}
	return nil
}

// parseTime parses a string as time.Time. It supports the RFC3339 format, unix seconds, and json marshalled time.Time
// structs.
func parseTime(s string) (time.Time, error) {
	// attempt to parse time as RFC3339 string
	t, err := time.Parse(time.RFC3339Nano, s)
	if err == nil {
		return t, nil
	}

	// attempt to parse time as float number of unix seconds
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		sec, dec := math.Modf(f)
		return time.Unix(int64(sec), int64(dec*(1e9))), nil
	}

	// attempt to parse time as json marshaled value
	if err := json.Unmarshal([]byte(s), &t); err == nil {
		return t, nil
	}

	return time.Time{}, err
}

// getFieldTags returns the tags or names that a struct field is identified by. It prioritizes the mpe tag over the
// json tag. It defaults to the field name if neither tag is available.
func getFieldTags(t reflect.StructField) (res []string) {
	if tags := t.Tag.Get(mapEnvTagName); len(tags) > 0 {
		for _, s := range strings.Split(tags, ",") {
			if len(s) > 0 {
				res = append(res, s)
			}
		}
	}

	// ignore json tags and field name if mpe tag is present
	if len(res) > 0 {
		return
	}

	if tags := t.Tag.Get("json"); len(tags) > 0 {
		jsonTags := strings.Split(tags, ",")
		if len(jsonTags) > 0 && len(jsonTags[0]) > 0 {
			res = append(res, jsonTags[0])
		}
	}

	// ignore field name if json tag is present
	if len(res) > 0 {
		return
	}

	res = append(res, t.Name)

	return
}
