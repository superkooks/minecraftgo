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

func (c *Conn) Read(buf []byte) (n int, err error) {
	n, err = c.TCP.Read(buf)
	if err != nil {
		return n, err
	}

	if c.Encrypted {
		c.Cipher.XORKeyStream(buf[:n], buf[:n])
	}

	return
}

func (c *Conn) Listener() {
	for {

		if !c.Compressed {
			var p UncompressedPacket
			Unmarshal(c, &p)

			fmt.Println("New packet:", p.PacketID)

			switch p.PacketID {
			case 0x01:
				// Encryption Request

				// Special unmarshalling due to multiple byte arrays
				var serverID String
				serverID.Deserialize(c)
				fmt.Println("server id:", serverID)

				var pubKeyLength VarInt
				pubKeyLength.Deserialize(c)

				pubKey := make([]byte, int(pubKeyLength))
				c.Read(pubKey)

				var verifyTokenLength VarInt
				verifyTokenLength.Deserialize(c)

				verifyToken := make([]byte, int(verifyTokenLength))
				c.Read(verifyToken)

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
				p := Marshal(EncryptionResponse{
					SharedSecretLen: VarInt(len(encShared)),
					SharedSecret:    encShared,
					VerifyTokenLen:  VarInt(len(encVerify)),
					VerifyToken:     encVerify,
				})
				c.TCP.Write(Marshal(UncompressedPacket{
					PacketID: 0x01,
					Length:   VarInt(len(p)),
				}))
				c.TCP.Write(p)

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
				Unmarshal(c, &comp)

				if comp.Threshold > 0 {
					fmt.Println("Enabling compression")
					c.Compressed = true
					c.CompressionThreshold = int(comp.Threshold)
				} else {
					fmt.Println("Received set compression packet with threshold <= 0, not enabling compression")
				}
			}
		} else {
			// Used to get initial length of packet
			var p CompressedPacket
			Unmarshal(c, &p)

			var reader io.Reader
			if int(p.DataLength) >= c.CompressionThreshold {
				var err error
				reader, err = zlib.NewReader(c)
				if err != nil {
					panic(err)
				}
			} else {
				reader = c
			}

			var d CompressedData
			Unmarshal(reader, &d)

			fmt.Println("New compressed packet:", d.PacketID)
			switch d.PacketID {
			case 0x02:
				// Login Success
				fmt.Println("*** Received login success")

			case 0x18:
				var plugin PluginMessage
				Unmarshal(reader, &plugin)
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
