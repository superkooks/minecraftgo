package minecraftgo

type UncompressedPacket struct {
	Length   VarInt
	PacketID VarInt
	Data     []byte
}
