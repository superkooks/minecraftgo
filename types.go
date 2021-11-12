package minecraftgo

import "io"

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
func (v *VarInt) Deserialize(data io.Reader) {
	value := 0
	index := 0
	for {
		b := make([]byte, 1)
		_, err := data.Read(b)
		if err != nil {
			panic(err)
		}
		value |= int(b[0]&0b01111111) << (index * 7)

		if (b[0] & 0b10000000) == 0 {
			break
		}

		index++

		if index > 5 {
			panic("varint is too big")
		}
	}

	*v = VarInt(value)
}

type String string

func (s String) Serialize() []byte {
	return append(VarInt(len(s)).Serialize(), []byte(s)...)
}

// Returns the rest of the data that has not been read
func (s *String) Deserialize(data io.Reader) {
	length := new(VarInt)
	length.Deserialize(data)

	b := make([]byte, int(*length))
	_, err := data.Read(b)
	if err != nil {
		panic(err)
	}

	*s = String(b)
}
