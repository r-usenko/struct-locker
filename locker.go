package structLocker

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"unicode"
	"unsafe"
)

const LockedFieldName = "_LOCKED"

// TODO Export refactoring error
var (
	errCalledByValStruct         = errors.New("structure called by val")
	errStructCalledByInvalidType = errors.New("structure is not struct")
	errMethodCalledByInvalidType = errors.New("method is not method")
	errMethodCalledByFunc        = errors.New("method is not method")
	errMethodErrorParsing        = errors.New("method is not method")
	errLockedFieldInvalidType    = fmt.Errorf("locked field %q, must be type of bool", LockedFieldName)
	errLockedFieldMissing        = fmt.Errorf("locked field %q, missing in set structure", LockedFieldName)
	errCantSetLockedField        = fmt.Errorf("can't set locked field %q", LockedFieldName)
	errValueInvalidType          = errors.New("value must have same type of field")
	errFieldMissing              = errors.New("destination field, missing in set structure")
	errStructureLocked           = errors.New("structure already locked for set")
)

var reg = regexp.MustCompile(`\.(\w+)-fm$`)

// LockStruct recursive set field LockedFieldName in structure to true
// - field LockedFieldName must be a bool type
// - structure must be struct and called by ref
func LockStruct(structure interface{}) error {
	m := reflect.ValueOf(structure)
	if m.Type().Kind() != reflect.Ptr {
		return errCalledByValStruct
	}

	mv := m.Elem()
	if mv.Type().Kind() != reflect.Struct {
		return errStructCalledByInvalidType
	}

	for i := 0; i < mv.NumField(); i++ {
		fv := mv.Field(i)

		if fv.Kind() == reflect.Ptr {
			fv = fv.Elem()
		}

		if fv.Kind() == reflect.Struct {
			fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
			if err := LockStruct(fv.Addr().Interface()); err != nil {
				return err
			}

			continue
		}

		if mv.Type().Field(i).Name != LockedFieldName {
			continue
		}

		if fv.Kind() != reflect.Bool {
			return errLockedFieldInvalidType
		}

		fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
		fv.SetBool(true)
	}

	return nil
}

// SetByName overwrite value of private field in struct.
//   - structure must have bool field LockedFieldName and is value == false
//   - value must be same type of private field TODO or ptr?
func SetByName(structure interface{}, name string, value interface{}) error {
	return UnsafeSetByName(structure, name, value, true)
}

// SetByMethod see SetByName for more details
// - private field must be named same as method except first character lowercase
func SetByMethod(structure interface{}, method interface{}, value interface{}) error {
	return UnsafeSetByMethod(structure, method, value, true)
}

// UnsafeSetByName - use SetByName instead
// DEPRECATED
//
//goland:noinspection GoDeprecation
func UnsafeSetByName(structure interface{}, name string, value interface{}, needCheckLocked bool) error {
	m := reflect.ValueOf(structure)
	if m.Type().Kind() != reflect.Ptr {
		return errCalledByValStruct
	}

	mv := m.Elem()
	if mv.Type().Kind() != reflect.Struct {
		return errStructCalledByInvalidType
	}

	if needCheckLocked {
		if err := checkLockedField(name, mv); err != nil {
			return err
		}
	}

	fv := mv.FieldByName(name)
	if reflect.DeepEqual(fv, reflect.Value{}) {
		return errFieldMissing
	}

	vv := reflect.ValueOf(value)
	if vv.Kind() != fv.Kind() {
		return errValueInvalidType
	}

	fv = reflect.NewAt(fv.Type(), unsafe.Pointer(fv.UnsafeAddr())).Elem()
	fv.Set(vv)

	return nil
}

// UnsafeSetByMethod  - use SetByMethod instead
// DEPRECATED
//
//goland:noinspection GoDeprecation
func UnsafeSetByMethod(structure interface{}, method interface{}, value interface{}, needCheckLocked bool) error {
	name, err := getPropertyName(method)
	if err != nil {
		return err
	}

	return UnsafeSetByName(structure, name, value, needCheckLocked)
}

// check structure have locked field
func checkLockedField(changedFieldName string, structure reflect.Value) error {
	if changedFieldName == LockedFieldName {
		return errCantSetLockedField
	}

	fLocked := structure.FieldByName(LockedFieldName)
	if reflect.DeepEqual(fLocked, reflect.Value{}) {
		return errLockedFieldMissing
	}

	if fLocked.Kind() != reflect.Bool {
		return errLockedFieldInvalidType
	}

	fLocked = reflect.NewAt(fLocked.Type(), unsafe.Pointer(fLocked.UnsafeAddr())).Elem()
	if fLocked.Interface().(bool) != false {
		return errStructureLocked
	}

	return nil
}

// convert named func to field name, to be less likely to make mistakes
func getPropertyName(i interface{}) (string, error) {
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Func {
		return "", errMethodCalledByInvalidType
	}

	rawName := runtime.FuncForPC(rv.Pointer()).Name()
	if rawName == "" {
		return "", errMethodCalledByFunc
	}

	res := reg.FindStringSubmatch(rawName)
	if len(res) != 2 {
		return "", errMethodErrorParsing
	}

	a := []rune(res[1])
	a[0] = unicode.ToLower(a[0])

	return string(a), nil
}
