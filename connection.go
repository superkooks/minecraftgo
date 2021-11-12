package minecraftgo

import (
	"bytes"
	"crypto/cipher"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
)

type Conn struct {
	TCP                  *net.TCPConn
	Compressed           bool
	Encrypted            bool
	AuthResp             APIAuthResponse
	Cipher               cipher.Stream
	CompressionThreshold int
}

func Connect(ip *net.TCPAddr, username string, email string, password string) (*Conn, error) {
	// Authenticate the user with Mojang
	resp, err := http.Post("https://authserver.mojang.com/authenticate", "application/json", bytes.NewBufferString(`
		{
			"agent": {
				"name": "Minecraft",
				"version": 1
			},
			"username": "`+email+`",
			"password": "`+password+`",
			"requestUser": true
		}
	`))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var authResp APIAuthResponse
	j := json.NewDecoder(resp.Body)
	j.Decode(&authResp)

	c := new(Conn)
	c.AuthResp = authResp

	c.TCP, err = net.DialTCP("tcp", nil, ip)
	if err != nil {
		return nil, err
	}

	go c.Listener()

	// Send handshake (state=2)
	p := Marshal(Handshake{
		Version: 340,
		Address: String(ip.IP.String() + "\x00FML\x00"), // Not used
		Port:    uint16(ip.Port),                        // Not used
		Next:    2,
	})
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
		Length:   VarInt(len(p)),
	}))
	c.TCP.Write(p)

	// Send Login Start
	p = Marshal(LoginStart{
		Username: String(username),
	})
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
		Length:   VarInt(len(p)),
	}))
	c.TCP.Write(p)
	fmt.Println("sent login request")

	return c, nil
}

func Ping(ip *net.TCPAddr) ([]byte, error) {
	c := new(Conn)

	var err error
	c.TCP, err = net.DialTCP("tcp", nil, ip)
	if err != nil {
		return []byte{}, err
	}

	// Send handshake (state=1)
	p := Handshake{
		Version: 340,
		Address: "",    // Not used
		Port:    25565, // Not used
		Next:    1,
	}
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
	}))
	c.TCP.Write(Marshal(p))

	// Send request
	c.TCP.Write(Marshal(UncompressedPacket{
		PacketID: 0,
	}))

	// Used to get initial length of packet
	var q UncompressedPacket
	Unmarshal(c.TCP, &q)

	var out String
	Unmarshal(c.TCP, &out)

	return []byte(out), nil
}
