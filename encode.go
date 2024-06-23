package main

import (
	"strconv"
	"fmt"
	"reflect"
	"bytes"
	"sort"
)

type encoder struct {
	buf *bytes.Buffer
}

func (e *encoder) error(v reflect.Value, format string, a ...any) error {
	return fmt.Errorf("%w: %w", InvalidMarshalError{Val: v}, fmt.Errorf(format, a...))
}

func (e *encoder) value(v reflect.Value) error {
	v, isNil := indirect(v)
	if isNil { // TODO?: Maybe just use a zero value instead
		return e.error(v, "%w: got a nil value", ValueError{})
	}
	switch k := v.Kind(); k {
	default:
		return e.error(v, "%w: unsupported Value Kind %v", ValueError{}, k)
	case reflect.String:
		str := v.String()
		len := strconv.Itoa(len(str))
		e.buf.WriteString(len)
		e.buf.WriteByte(':')
		e.buf.WriteString(str)
		return nil
	case reflect.Struct:
		type fieldPair struct {
			tag string
			i int
		}
	    ty := v.Type()
		fields := make([]fieldPair, 0, ty.NumField())
		for i := 0; i < ty.NumField(); i ++ {
			tag, ok := ty.Field(i).Tag.Lookup("bencode")
			if !ok {
				continue
			}
			if len(tag) > 0 && tag[0] == '?' {
				tag = tag[1:]
			}
			fields = append(fields, fieldPair{ tag, i })
		}
		sort.Slice(fields, func(i, j int) bool {
			return fields[i].tag < fields[j].tag
		})
		e.buf.WriteByte('d')
		for _, field := range fields {
			keyLen := strconv.Itoa(len(field.tag))
			e.buf.WriteString(keyLen)
			e.buf.WriteByte(':')
			e.buf.WriteString(field.tag)
			err := e.value(v.Field(field.i))
			if err != nil {
				return err
			}
		}
		e.buf.WriteByte('e')
		return nil
	case reflect.Array, reflect.Slice:
		e.buf.WriteByte('l')
		for i := 0; i < v.Len(); i ++ {
			err := e.value(v.Index(i))
			if err != nil {
				return err
			}
		}
		e.buf.WriteByte('e')
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var nstr string
		if isAnyUint(k) {
			n := v.Uint()
			nstr = strconv.FormatUint(n, 10)
		} else {
			n := v.Int()
			nstr = strconv.FormatInt(n, 10)
		}
		e.buf.WriteByte('i')
		e.buf.WriteString(nstr)
		e.buf.WriteByte('e')
		return nil
	}
}

func Marshal(a any) ([]byte, error) {
	e := encoder{buf: bytes.NewBuffer(nil)}
	err := e.value(reflect.ValueOf(a))
	if err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}