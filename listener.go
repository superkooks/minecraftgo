package minecraftgo

import (
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
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

			var fullBuf *bytes.Buffer
			if c.Encrypted {
				c.Cipher.XORKeyStream(b[:n], b[:n])

				// Used to get initial length of packet
				var q UncompressedPacket
				Unmarshal(b[:n], &q)

				fullBuf = new(bytes.Buffer)
				fullBuf.Write(b[:n])

				// Read the rest of the bytes
				for fullBuf.Len() < int(q.Length)-5 {
					b := make([]byte, 2048)
					n, err := c.TCP.Read(b)
					if err != nil {
						panic(err)
					}

					c.Cipher.XORKeyStream(b[:n], b[:n])
					fullBuf.Write(b[:n])
				}
			} else {
				// Used to get initial length of packet
				var q UncompressedPacket
				Unmarshal(b[:n], &q)

				fullBuf = new(bytes.Buffer)
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
			}

			var q UncompressedPacket
			Unmarshal(fullBuf.Bytes(), &q)
			fmt.Println("New packet:", q.PacketID)

			switch q.PacketID {
			case 0x01:
				// Encryption Request

				// Special unmarshalling due to multiple byte arrays
				var serverID String
				q.Data = serverID.Deserialize(q.Data)
				fmt.Println("server id:", serverID)

				var pubKeyLength VarInt
				q.Data = pubKeyLength.Deserialize(q.Data)

				var pubKey []byte
				pubKey = q.Data[:pubKeyLength]
				q.Data = q.Data[pubKeyLength:]

				var verifyTokenLength VarInt
				q.Data = verifyTokenLength.Deserialize(q.Data)

				var verifyToken []byte
				verifyToken = q.Data[:verifyTokenLength]
				q.Data = q.Data[verifyTokenLength:]

				sharedSecret := make([]byte, 16)
				rand.Read(sharedSecret)

				digest := authDigest(string(serverID) + string(sharedSecret) + string(pubKey))

				resp, err := http.Post("https://sessionserver.mojang.com/session/minecraft/join", "application/json", bytes.NewBufferString(`
					{
						"accessToken": "`+c.AuthResp.AccessToken+`",
						"selectedProfile": "`+c.AuthResp.SelectedProfile.ID+`",
						"serverId": "`+digest+`"
					}	
				`))
				if err != nil {
					panic(err)
				}

				b, err := io.ReadAll(resp.Body)
				if err != nil {
					panic(err)
				}

				fmt.Println("session join resp:", string(b))

				fmt.Println("verify token:", verifyToken)
				fmt.Println("digest:", digest)

				pub, err := x509.ParsePKIXPublicKey(pubKey)
				if err != nil {
					panic(err)
				}

				encShared, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), sharedSecret)
				if err != nil {
					panic(err)
				}

				encVerify, err := rsa.EncryptPKCS1v15(rand.Reader, pub.(*rsa.PublicKey), verifyToken)
				if err != nil {
					panic(err)
				}

				// Send encryption response
				c.TCP.Write(Marshal(UncompressedPacket{
					PacketID: 0x01,
					Data: Marshal(EncryptionResponse{
						SharedSecretLen: VarInt(len(encShared)),
						SharedSecret:    encShared,
						VerifyTokenLen:  VarInt(len(encVerify)),
						VerifyToken:     encVerify,
					}),
				}))

				c.Encrypted = true

				fmt.Println("Enabling encryption")
				block, err := aes.NewCipher(sharedSecret)
				if err != nil {
					panic(err)
				}

				c.Cipher = newCFB8(block, sharedSecret, true)

			case 0x03:
				fmt.Println("Enabling compression")
				c.Compressed = true
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
