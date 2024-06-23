package main

import (
	"bytes"
	"fmt"
	"reflect"
	"strconv"
)

type decoder struct {
	input []byte
	pos   int
}

func (d *decoder) empty() bool {
	return len(d.input) <= d.pos
}

func (d *decoder) peek() byte {
	return d.input[d.pos]
}

func (d *decoder) error(format string, a ...any) error {
	return fmt.Errorf("%w: %w", InvalidUnmarshalError{Pos: d.pos}, fmt.Errorf(format, a...))
}

func (d *decoder) consumeString() (string, error) {
	if d.empty() || '0' > d.peek() || d.peek() > '9' {
		return "", d.error("%w: expected a string length prefix", SyntaxError{})
	}
	col := bytes.IndexByte(d.input[d.pos:], ':')
	if col == -1 {
		return "", d.error("%w: expected a string colon seperator", SyntaxError{})
	}
	col += d.pos
	sizestr := d.input[d.pos:col]
	size, err := strconv.Atoi(string(sizestr))
	assert(size >= 0, "somehow got negative size for a string length")
	if err != nil {
		return "", d.error("%w: %w", SyntaxError{}, err)
	}
	end := col + size + 1
	if len(d.input) < end {
		return "", d.error("%w: length goes out of bounds for input", SyntaxError{})
	}
	str := d.input[col+1 : end]
	d.pos = end
	return string(str), nil
}

func (d *decoder) check(v reflect.Value, kinds ...reflect.Kind) error {
	valid := false
	for _, kind := range kinds {
		if v.Kind() == kind {
			valid = true
			break
		}
	}
	if !valid {
		return d.error("%w: expected value to be %v, but got %v", ValueError{}, kinds, v.Kind())
	}
	if !v.CanSet() {
		return d.error("%w: corresponding value of kind %v is either unexported or not addressable", ValueError{}, v.Kind())
	}
	return nil
}

func leadingZero(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	if b[0] == '-' {
		b = b[1:]
	}
	if len(b) <= 1 {
		return false
	}
	return b[0] == '0'
}

func (d *decoder) integer(v reflect.Value) error {
	assert(d.peek() == 'i', "expected 'i' as an integer start")
	err := d.check(v, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64)
	if err != nil {
		return err
	}
	d.pos++
	end := bytes.IndexByte(d.input[d.pos:], 'e')
	if end == -1 {
		return d.error("%w: couldn't find integer ending 'e'", SyntaxError{})
	}
	end += d.pos
	nstr := d.input[d.pos:end]
	if leadingZero(nstr) {
		return d.error("%w: int values can't have leading zeroes", SyntaxError{})
	}
	if len(nstr) == 2 && nstr[0] == '-' && nstr[1] == '0' {
		return d.error("%w: -0 is an invalid int value", SyntaxError{})
	}
	n, err := strconv.ParseInt(string(nstr), 10, 64)
	if err != nil {
		return d.error("%w: %w", SyntaxError{}, err)
	}
	if isAnyUint(v.Kind()) {
		v.SetUint(uint64(n))
	} else {
		v.SetInt(n)
	}
	d.pos = end + 1
	return nil
}

func (d *decoder) list(v reflect.Value) error {
	assert(d.peek() == 'l', "expected 'l' as a list start")
	d.pos++
	err := d.check(v, reflect.Array, reflect.Slice)
	if err != nil {
		return err
	}
	for i := 0; ; i++ {
		if d.empty() {
			return d.error("%w: expected list ending 'e' or another list item", SyntaxError{})
		}
		if d.peek() == 'e' {
			break
		}
		if v.Kind() == reflect.Slice {
			if i >= v.Cap() {
				v.Grow(1)
			}
			if i >= v.Len() {
				v.SetLen(i + 1)
			}
		}

		if i < v.Len() {
			vi := v.Index(i)
			err := d.value(vi)
			if err != nil {
				return err
			}
		} else {
			return d.error("%w: too many values for array", ValueError{})
		}
	}
	d.pos++
	return nil
}

func (d *decoder) dict(v reflect.Value) error {
	assert(d.peek() == 'd', "expected 'd' as a dict start")
	d.pos++
	err := d.check(v, reflect.Struct)
	if err != nil {
		return err
	}
	tagToField := make(map[string]reflect.Value)
	required := make(map[string]struct{})
	ty := v.Type()
	for i := 0; i < ty.NumField(); i++ {
		tag, ok := ty.Field(i).Tag.Lookup("bencode")
		if !ok {
			continue
		}
		optional := len(tag) > 0 && tag[0] == '?'
		if optional {
			tag = tag[1:]
		} else {
			required[tag] = struct{}{}
		}
		tagToField[tag] = v.Field(i)
	}
	for {
		if d.empty() {
			return d.error("%w: expected dict ending 'e' or another key-value pair", SyntaxError{})
		}
		if d.peek() == 'e' {
			break
		}
		keyTag, err := d.consumeString()
		if err != nil {
			return err
		}
		field, ok := tagToField[keyTag]
		if !ok {
			return d.error("%w: unexpected key '%s' (type '%s' doesn't have a field with such bencode tag)", ValueError{}, keyTag, v.Type())
		}
		err = d.value(field)
		if err != nil {
			return err
		}
		delete(required, keyTag)
	}
	if len(required) > 0 {
		missing := make([]string, 0, len(required))
		for k, _ := range required {
			missing = append(missing, k)
		}
		return d.error("%w: type '%s' missing required keys %v", ValueError{}, v.Type(), missing)
	}
	d.pos++
	return nil
}

func (d *decoder) string(v reflect.Value) error {
	err := d.check(v, reflect.String)
	if err != nil {
		return err
	}
	str, err := d.consumeString()
	if err != nil {
		return err
	}
	v.SetString(str)
	return nil
}

func (d *decoder) value(v reflect.Value) error {
	if d.empty() {
		return d.error("%w: expected a value, but got empty input", SyntaxError{})
	}
	v, isNil := indirect(v)
	if isNil {
		return d.error("%w: got a nil value", ValueError{})
	}
	switch d.peek() {
	case 'i':
		return d.integer(v)
	case 'l':
		return d.list(v)
	case 'd':
		return d.dict(v)
	default:
		return d.string(v)
	}
}

func Unmarshal(input []byte, a any) error {
	d := decoder{input: input, pos: 0}
	return d.value(reflect.ValueOf(a))
}
