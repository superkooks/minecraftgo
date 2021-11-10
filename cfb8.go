package minecraftgo

import "crypto/cipher"

// This code was copied from
// https://stackoverflow.com/questions/23897809/different-results-in-go-and-pycrypto-when-using-aes-cfb/37234233#37234233
// Thanks kostya!

// CFB stream with 8 bit segment size
// See http://csrc.nist.gov/publications/nistpubs/800-38a/sp800-38a.pdf
type cfb8 struct {
	b         cipher.Block
	blockSize int
	in        []byte
	out       []byte

	decrypt bool
}

func (x *cfb8) XORKeyStream(dst, src []byte) {
	for i := range src {
		x.b.Encrypt(x.out, x.in)
		copy(x.in[:x.blockSize-1], x.in[1:])
		if x.decrypt {
			x.in[x.blockSize-1] = src[i]
		}
		dst[i] = src[i] ^ x.out[0]
		if !x.decrypt {
			x.in[x.blockSize-1] = dst[i]
		}
	}
}

func newCFB8(block cipher.Block, iv []byte, decrypt bool) cipher.Stream {
	blockSize := block.BlockSize()
	if len(iv) != blockSize {
		// stack trace will indicate whether it was de or encryption
		panic("cipher.newCFB: IV length must equal block size")
	}
	x := &cfb8{
		b:         block,
		blockSize: blockSize,
		out:       make([]byte, blockSize),
		in:        make([]byte, blockSize),
		decrypt:   decrypt,
	}
	copy(x.in, iv)

	return x
}
