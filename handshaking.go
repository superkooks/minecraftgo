package minecraftgo

// CLIENT
type Handshake struct {
	Version VarInt
	Address String
	Port    uint16
	Next    VarInt
}

// SERVER
type PingResponse struct {
	Description PingText        `json:"description"`
	Players     PingPlayers     `json:"players"`
	Version     PingVersionInfo `json:"version"`
	Favicon     string          `json:"favicon"`
	ModInfo     PingModInfo     `json:"modinfo,omitempty"`
}

type PingText struct {
	Text string `json:"text"`
}

type PingPlayers struct {
	Max    int          `json:"max"`
	Online int          `json:"online"`
	Sample []PingPlayer `json:"sample"`
}

type PingPlayer struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PingVersionInfo struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type PingModInfo struct {
	Type    string    `json:"type"`
	ModList []PingMod `json:"modList"`
}

type PingMod struct {
	ModID   string `json:"modid"`
	Version string `json:"version"`
}
