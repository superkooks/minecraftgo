package minecraftgo

// CLIENT
type LoginStart struct {
	Username String
}

type EncryptionResponse struct {
	SharedSecretLen VarInt
	SharedSecret    []byte
	VerifyTokenLen  VarInt
	VerifyToken     []byte
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

type SetCompression struct {
	Threshold VarInt
}

type PluginMessage struct {
	Namespace String
	Data      []byte
}
