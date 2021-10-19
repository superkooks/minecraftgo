package minecraftgo

import (
	"net"
)

type Conn struct {
	TCP        *net.TCPConn
	Compressed bool
}

func Connect(ip *net.TCPAddr) (*Conn, error) {
	c := new(Conn)

	var err error
	c.TCP, err = net.DialTCP("tcp", nil, ip)
	if err != nil {
		return nil, err
	}

	// Send handshake (state=2)

	return c, nil
}

func Ping(ip *net.TCPAddr) ([]byte, error) {
	c := new(Conn)

	var err error
	c.TCP, err = net.DialTCP("tcp", nil, ip)
	if err != nil {
		return []byte{}, err
	}

	c.TCP.SetReadBuffer(100000000000000000)

	// Send handshake (state=1)
	p := Handshake{
		Version: 340,
		Address: "",    // Not used
		Port:    25565, // Not used
		Next:    1,
	}
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
		Data:     Marshal(p),
	}))

	// Send request
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
	}))

	// Wait for response
	b := make([]byte, 22816040000)
	n, err := c.TCP.Read(b)
	if err != nil {
		return []byte{}, err
	}

	var q UncompressedPacket
	Unmarshal(b[:n], &q)

	var out String
	Unmarshal(q.Data, &out)

	return []byte(out), nil
}
