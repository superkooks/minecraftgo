package main

import (
	"encoding/binary"
	"math"
	"reflect"
)

// Unmarshal is a wrapper for unmarshal
func Unmarshal(data []byte, out interface{}) {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		panic("cannot unmarshal to a non-pointer")
	}

	o, _ := unmarshal(data, reflect.TypeOf(out).Elem())
	p := reflect.ValueOf(out).Elem()
	p.Set(o)
}

// Returns the value and the rest of the data that has not been read
func unmarshal(data []byte, typ reflect.Type) (reflect.Value, []byte) {
	if typ.Kind() == reflect.Struct {
		in := reflect.New(typ)

		// Recursively marshal the fields of struct m
		d := data
		for _, v := range reflect.VisibleFields(typ) {
			var u reflect.Value
			u, d = unmarshal(d, reflect.Indirect(in).FieldByIndex(v.Index).Type())
			reflect.Indirect(in).FieldByIndex(v.Index).Set(u)
		}

		return reflect.Indirect(in), data
	}

	// If it isn't a struct, serialize it accordingly
	switch typ {

	// Custom types
	case reflect.TypeOf(VarInt(0)):
		v := new(VarInt)
		d := v.Deserialize(data)
		return reflect.ValueOf(*v), d

	case reflect.TypeOf(String("")):
		v := new(String)
		d := v.Deserialize(data)
		return reflect.ValueOf(*v), d

	default:
		// Builtin types
		switch typ.Kind() {
		case reflect.Bool:
			if data[0] != 0 {
				return reflect.ValueOf(true), data[1:]
			} else {
				return reflect.ValueOf(false), data[1:]
			}

		case reflect.Uint8:
			return reflect.ValueOf(data[0]), data[1:]

		case reflect.Slice:
			return reflect.ValueOf(data), []byte{}

		case reflect.Int16:
			return reflect.ValueOf(byteArrayToInt(data[:2])), data[2:]

		case reflect.Uint16:
			return reflect.ValueOf(binary.BigEndian.Uint16(data[:2])), data[2:]

		case reflect.Int32:
			return reflect.ValueOf(byteArrayToInt(data[:4])), data[4:]

		case reflect.Uint32:
			return reflect.ValueOf(binary.BigEndian.Uint32(data[:4])), data[4:]

		case reflect.Int64:
			return reflect.ValueOf(byteArrayToInt(data[:8])), data[8:]

		case reflect.Uint64:
			return reflect.ValueOf(binary.BigEndian.Uint64(data[:8])), data[8:]

		case reflect.Float32:
			u, d := unmarshal(data, reflect.TypeOf(uint32(0)))
			return reflect.ValueOf(math.Float32frombits(u.Interface().(uint32))), d

		case reflect.Float64:
			u, d := unmarshal(data, reflect.TypeOf(uint64(0)))
			return reflect.ValueOf(math.Float64frombits(u.Interface().(uint64))), d

		default:
			panic("cannot convert type " + typ.Kind().String())
		}
	}
}
