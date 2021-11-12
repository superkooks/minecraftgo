package minecraftgo

import (
	"encoding/binary"
	"io"
	"math"
	"reflect"
)

// Unmarshal is a wrapper for unmarshal
func Unmarshal(data io.Reader, out interface{}) {
	if reflect.TypeOf(out).Kind() != reflect.Ptr {
		panic("cannot unmarshal to a non-pointer")
	}

	o := unmarshal(data, reflect.TypeOf(out).Elem())
	p := reflect.ValueOf(out).Elem()
	p.Set(o)
}

// Returns the value and the rest of the data that has not been read
func unmarshal(data io.Reader, typ reflect.Type) reflect.Value {
	if typ.Kind() == reflect.Struct {
		in := reflect.New(typ)

		// Recursively marshal the fields of struct m
		d := data
		for _, v := range reflect.VisibleFields(typ) {
			u := unmarshal(d, reflect.Indirect(in).FieldByIndex(v.Index).Type())
			reflect.Indirect(in).FieldByIndex(v.Index).Set(u)
		}

		return reflect.Indirect(in)
	}

	// If it isn't a struct, serialize it accordingly
	switch typ {

	// Custom types
	case reflect.TypeOf(VarInt(0)):
		v := new(VarInt)
		v.Deserialize(data)
		return reflect.ValueOf(*v)

	case reflect.TypeOf(String("")):
		v := new(String)
		v.Deserialize(data)
		return reflect.ValueOf(*v)

	default:
		// Builtin types
		switch typ.Kind() {
		case reflect.Bool:
			b := make([]byte, 1)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}

			if b[0] != 0 {
				return reflect.ValueOf(true)
			} else {
				return reflect.ValueOf(false)
			}

		case reflect.Uint8:
			b := make([]byte, 1)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(b[0])

		case reflect.Slice:
			b, err := io.ReadAll(data)
			if err != nil {
				panic(err)
			}

			return reflect.ValueOf(b)

		case reflect.Int16:
			b := make([]byte, 2)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(byteArrayToInt(b))

		case reflect.Uint16:
			b := make([]byte, 2)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(binary.BigEndian.Uint16(b))

		case reflect.Int32:
			b := make([]byte, 4)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(byteArrayToInt(b))

		case reflect.Uint32:
			b := make([]byte, 4)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(binary.BigEndian.Uint32(b))

		case reflect.Int64:
			b := make([]byte, 8)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(byteArrayToInt(b))

		case reflect.Uint64:
			b := make([]byte, 8)
			_, err := data.Read(b)
			if err != nil {
				panic(err)
			}
			return reflect.ValueOf(binary.BigEndian.Uint64(b))

		case reflect.Float32:
			u := unmarshal(data, reflect.TypeOf(uint32(0)))
			return reflect.ValueOf(math.Float32frombits(u.Interface().(uint32)))

		case reflect.Float64:
			u := unmarshal(data, reflect.TypeOf(uint64(0)))
			return reflect.ValueOf(math.Float64frombits(u.Interface().(uint64)))

		default:
			panic("cannot convert type " + typ.Kind().String())
		}
	}
}
