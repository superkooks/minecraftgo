package main

import (
	"encoding/binary"
	"math"
	"reflect"
)

func Marshal(m interface{}) []byte {
	// Treat uncompressed packet differently
	if p, ok := m.(UncompressedPacket); ok {
		p.Length = VarInt(len(Marshal(p.PacketID)) + len(p.Data))
		m = p
	}

	if reflect.TypeOf(m).Kind() == reflect.Struct {
		in := reflect.ValueOf(m)

		// Recursively marshal the fields of struct m
		var data []byte
		for _, v := range reflect.VisibleFields(reflect.TypeOf(m)) {
			data = append(data, Marshal(in.FieldByIndex(v.Index).Interface())...)
		}

		return data
	}

	// If it isn't a struct, serialize it accordingly
	switch v := m.(type) {

	// Custom types
	case VarInt:
		return v.Serialize()

	case String:
		return v.Serialize()

	// Builtin types
	case bool:
		if v {
			return []byte{0x01}
		} else {
			return []byte{0x00}
		}

	case byte:
		return []byte{v}

	case []byte:
		return v

	case int16:
		return intToByteArray(v)

	case uint16:
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, v)
		return b

	case int32:
		return intToByteArray(v)

	case uint32:
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, v)
		return b

	case int64:
		return intToByteArray(v)

	case uint64:
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, v)
		return b

	case float32:
		return Marshal(math.Float32bits(v))

	case float64:
		return Marshal(math.Float64bits(v))

	default:
		panic("could not convert type " + reflect.TypeOf(m).Kind().String())
	}
}
