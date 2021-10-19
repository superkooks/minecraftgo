package main

// CLIENT
type Handshake struct {
	Version VarInt
	Address String
	Port    uint16
	Next    VarInt
}
