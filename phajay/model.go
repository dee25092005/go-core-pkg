package phajay

type GenerateQRRequest struct {
	Amount      int    `json:"amount"`
	Description string `json:"description"`
	Tag1        string `json:"tag1,omitempty"`
	Tag2        string `json:"tag2,omitempty"`
	Tag3        string `json:"tag3,omitempty"`
}
