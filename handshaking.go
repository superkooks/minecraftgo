package minecraftgo

// CLIENT
type Handshake struct {
	Version VarInt
	Address String
	Port    uint16
	Next    VarInt
}

// SERVER
type Response struct {
	Description Text        `json:"description"`
	Players     Players     `json:"players"`
	Version     VersionInfo `json:"version"`
	Favicon     string      `json:"favicon"`
	ModInfo     ModInfo     `json:"modinfo,omitempty"`
}

type Text struct {
	Text string `json:"text"`
}

type Players struct {
	Max    int      `json:"max"`
	Online int      `json:"online"`
	Sample []Player `json:"sample"`
}

type Player struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type VersionInfo struct {
	Name     string `json:"name"`
	Protocol int    `json:"protocol"`
}

type ModInfo struct {
	Type    string `json:"type"`
	ModList []Mod  `json:"modList"`
}

type Mod struct {
	ModID   string `json:"modid"`
	Version string `json:"version"`
}
