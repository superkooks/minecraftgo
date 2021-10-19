package main

type VarInt int

func (v VarInt) Serialize() []byte {
	data := []byte{}
	for {
		if (v & 0xffffff80) == 0 {
			data = append(data, byte(v))
			return data
		}

		data = append(data, byte(v&0x7f|0x80))
		v >>= 7
	}
}

// Returns the rest of the data that has not been read
func (v *VarInt) Deserialize(data []byte) []byte {
	value := 0
	index := 0
	for {
		b := data[index]
		value |= int(b&0b01111111) << (index * 7)

		if (data[index] & 0b10000000) == 0 {
			break
		}

		index++

		if index > 5 {
			panic("varint is too big")
		}
	}

	*v = VarInt(value)
	return data[index+1:]
}

type String string

func (s String) Serialize() []byte {
	return append(VarInt(len(s)).Serialize(), []byte(s)...)
}

// Returns the rest of the data that has not been read
func (s *String) Deserialize(data []byte) []byte {
	length := new(VarInt)
	data = length.Deserialize(data)

	offset := len(length.Serialize())
	*s = String(data[:int(*length)+offset])

	return data[int(*length):]
}
