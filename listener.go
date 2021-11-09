package minecraftgo

import (
	"bytes"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *Conn) Listener() {
	for {
		if !c.Compressed {
			b := make([]byte, 2048)
			n, err := c.TCP.Read(b)
			if err != nil {
				panic(err)
			}

			// Used to get initial length of packet
			var q UncompressedPacket
			Unmarshal(b[:n], &q)

			fullBuf := new(bytes.Buffer)
			fullBuf.Write(b[:n])

			// Read the rest of the bytes
			for fullBuf.Len() < int(q.Length)-5 {
				b := make([]byte, 2048)
				n, err := c.TCP.Read(b)
				if err != nil {
					panic(err)
				}

				fullBuf.Write(b[:n])
			}

			switch q.PacketID {
			case 0x01:
				// Encryption Request

				// Special unmarshalling due to multiple byte arrays
				var serverID String
				q.Data = serverID.Deserialize(q.Data)

				var pubKeyLength VarInt
				q.Data = pubKeyLength.Deserialize(q.Data)

				var pubKey []byte
				pubKey = q.Data[:pubKeyLength]
				q.Data = q.Data[pubKeyLength:]

				var verifyTokenLength VarInt
				q.Data = verifyTokenLength.Deserialize(q.Data)

				var verifyToken []byte
				pubKey = q.Data[:verifyTokenLength]
				q.Data = q.Data[verifyTokenLength:]

				sharedSecret := make([]byte, 16)
				rand.Read(sharedSecret)

				digest := authDigest(string(serverID) + string(sharedSecret) + string(pubKey))

				http.Post("https://sessionserver.mojang.com/session/minecraft/join", "application/json", bytes.NewBufferString(`
					{
						"accessToken": "`+c.AccessToken+`",
						"selectedProfile": "<player's uuid without dashes>",
						"serverId": "<serverHash>"
					}	
				`))

				fmt.Println(verifyToken)
				fmt.Println(digest)
			}
		} else {

		}
	}
}

func authDigest(s string) string {
	h := sha1.New()
	io.WriteString(h, s)
	hash := h.Sum(nil)

	// Check for negative hashes
	negative := (hash[0] & 0x80) == 0x80
	if negative {
		hash = twosComplement(hash)
	}

	// Trim away zeroes
	res := strings.TrimLeft(hex.EncodeToString(hash), "0")
	if negative {
		res = "-" + res
	}

	return res
}

// little endian
func twosComplement(p []byte) []byte {
	carry := true
	for i := len(p) - 1; i >= 0; i-- {
		p[i] = byte(^p[i])
		if carry {
			carry = p[i] == 0xff
			p[i]++
		}
	}
	return p
}
