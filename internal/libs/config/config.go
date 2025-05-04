package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	_tagKeyEnv     = "env"
	_tagKeyDefault = "default"
	_tagKeyBinding = "binding"

	_bindingRequired = "required"
)

var errKeyNotFound = fmt.Errorf("key not found")

func Load(c any) error {
	val := reflect.ValueOf(c)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	typ := val.Type()

	getters := []func(t reflect.StructTag, expectedType reflect.Type) (any, error){
		getFromEnv,
		getFromDefault,
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		found := false
		j := 0
		for found == false && j < len(getters) {
			getVal := getters[j]
			v, err := getVal(field.Tag, field.Type)
			if err == nil {
				val.Field(i).Set(reflect.ValueOf(v))
				found = true
			} else if err != errKeyNotFound {
				return err
			}
			j += 1
		}

		if !found && isRequired(field.Tag) {
			return fmt.Errorf("required config field not found: %s", field.Name)
		}

	}

	return nil
}

func getFromEnv(t reflect.StructTag, expectedType reflect.Type) (any, error) {
	keyName := t.Get(_tagKeyEnv)
	rawVal, ok := os.LookupEnv(keyName)
	if !ok {
		return nil, errKeyNotFound
	}

	castVal, err := getValueFromString(rawVal, expectedType)
	if err != nil {
		return nil, err
	}

	return castVal, nil
}

func getFromDefault(t reflect.StructTag, expectedType reflect.Type) (any, error) {
	rawVal := t.Get(_tagKeyDefault)
	if rawVal == "" {
		return nil, errKeyNotFound
	}

	castVal, err := getValueFromString(rawVal, expectedType)
	if err != nil {
		return nil, errKeyNotFound
	}

	return castVal, nil
}

func getValueFromString(val string, expectedType reflect.Type) (any, error) {
	switch expectedType.Kind() {
	case reflect.String:
		return val, nil
	case reflect.Int:
		return strconv.Atoi(val)
	case reflect.Bool:
		return strconv.ParseBool(val)
	default:
		return nil, fmt.Errorf("unsupported type: %v", expectedType)
	}
}

func isRequired(t reflect.StructTag) bool {
	bindings := getBindings(t)
	for _, b := range bindings {
		if b == _bindingRequired {
			return true
		}
	}
	return false
}

func getBindings(t reflect.StructTag) []string {
	val := t.Get(_tagKeyBinding)
	if val == "" {
		return []string{}
	}

	return strings.Split(val, ",")
}
