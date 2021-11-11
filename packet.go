package minecraftgo

type UncompressedPacket struct {
	Length   VarInt
	PacketID VarInt
	Data     []byte
}

type CompressedPacket struct {
	PacketLength VarInt
	DataLength   VarInt
	Data         []byte
}

type CompressedData struct {
	PacketID VarInt
	Data     []byte
}
