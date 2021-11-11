package minecraftgo

import (
	"bytes"
	"compress/zlib"
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

func (c *Conn) ReadStream() {
	for {
		b := make([]byte, 2048)
		n, err := c.TCP.Read(b)
		if err != nil {
			panic(err)
		}

		_, err = c.Buffer.Write(b[:n])
		if err != nil {
			panic(err)
		}
	}
}

func (c *Conn) Listener() {
	for {
		// TODO Deduplicate the code

		if !c.Compressed {
			c.Buffer.

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
				for fullBuf.Len() < int(q.Length)-3 {
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
				for fullBuf.Len() < int(q.Length)-3 {
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
				// Set Compression
				var comp SetCompression
				Unmarshal(q.Data, &comp)

				if comp.Threshold > 0 {
					fmt.Println("Enabling compression")
					c.Compressed = true
				} else {
					fmt.Println("Received set compression packet with threshold <= 0, not enabling compression")
				}
			}
		} else {
			fmt.Println("received packet with compression=true")
			b := make([]byte, 2048)
			// c.TCP.SetReadDeadline(time.Now().Add(time.Millisecond * 250))
			n, err := c.TCP.Read(b)
			if err != nil {
				panic(err)
			}

			fmt.Println("decrypting header")
			c.Cipher.XORKeyStream(b[:n], b[:n])

			fmt.Println("unmarshalling uncompressed header")
			// Used to get initial length of packet
			var q CompressedPacket
			Unmarshal(b[:n], &q)

			fullBuf := new(bytes.Buffer)
			fullBuf.Write(b[:n])

			// Read the rest of the bytes
			for fullBuf.Len() < int(q.PacketLength)-3 {
				b := make([]byte, 2048)
				n, err := c.TCP.Read(b)
				if err != nil {
					panic(err)
				}

				c.Cipher.XORKeyStream(b[:n], b[:n])
				fullBuf.Write(b[:n])
			}

			fmt.Println("read entire message")

			var p CompressedPacket
			Unmarshal(fullBuf.Bytes(), &p)

			fmt.Println("packet length:", q.PacketLength)
			fmt.Println("compressed data length:", q.DataLength)

			var d CompressedData
			if p.DataLength > 0 {
				reader, err := zlib.NewReader(bytes.NewBuffer(q.Data))
				if err != nil {
					panic(err)
				}

				decompressed, err := io.ReadAll(reader)
				if err != nil {
					panic(err)
				}

				hex.Dump(decompressed)

				Unmarshal(decompressed, &d)
			} else {
				Unmarshal(p.Data, &d)
			}

			fmt.Println("New compressed packet:", d.PacketID)
			switch d.PacketID {
			case 0x02:
				// Login Success
				fmt.Println("Received login success")

			case 0x18:
				var plugin PluginMessage
				Unmarshal(d.Data, &plugin)
				fmt.Println("plugin message in namespace", plugin.Namespace)
			}
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
