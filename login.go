package minecraftgo

// CLIENT
type LoginStart struct {
	Username String
}

// SERVER

// Not used - for reference only
type EncryptionRequest struct {
	ServerID          String
	PubKeyLength      VarInt
	PubKey            []byte
	VerifyTokenLength VarInt
	VerifyToken       []byte
}