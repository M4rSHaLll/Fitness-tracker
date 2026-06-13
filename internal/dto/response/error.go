package response

type Error struct {
	Error string `json:"error"`
	Code  string `json:"code,omitempty"`
}
