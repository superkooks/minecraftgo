package minecraftgo

type UncompressedPacket struct {
	Length   VarInt
	PacketID VarInt
}

type CompressedPacket struct {
	PacketLength VarInt
	DataLength   VarInt
}

type CompressedData struct {
	PacketID VarInt
}
