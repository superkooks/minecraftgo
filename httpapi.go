package minecraftgo

type APIAuthResponse struct {
	User              APIUser      `json:"user"`
	ClientToken       string       `json:"clientToken"`
	AccessToken       string       `json:"accessToken"`
	AvailableProfiles []APIProfile `json:"availableProfiles"`
	SelectedProfile   APIProfile   `json:"selectedProfile"`
}

type APIUser struct {
	Username   string        `json:"username"`
	Properties []APIProperty `json:"properties"`
	ID         string        `json:"id"` // Remote ID
}

type APIProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type APIProfile struct {
	Name string `json:"name"` // Username
	ID   string `json:"id"`   // Hex UUID
}
