package main

type UncompressedPacket struct {
	Length   VarInt
	PacketID VarInt
	Data     []byte
}
