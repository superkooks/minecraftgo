package minecraftgo

import (
	"reflect"
	"unsafe"
)

func intToByteArray(num interface{}) []byte {
	// Holy shit this is sketchy
	if !reflect.TypeOf(num).ConvertibleTo(reflect.TypeOf(int64(0))) {
		// If num can't be converted to an int64, then it isn't a number?
		panic("cannot convert non-integer to byte array")
	}

	size := int(unsafe.Sizeof(num))
	arr := make([]byte, size)
	for i := 0; i < size; i++ {
		byt := *(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&num)) + uintptr(i)))
		arr[i] = byt
	}
	return arr
}

func byteArrayToInt(arr []byte) int64 {
	val := int64(0)
	size := len(arr)
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}

	return val
}
