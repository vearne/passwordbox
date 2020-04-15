package model

type DetailItem struct {
	ID         int    `json:"-"`
	Title      string `json:"title"`
	Account    string `json:"account"`
	Password   string `json:"password"`
	Comment    string `json:"comment"`
	ModifiedAt string `json:"modifiedAt"`
}

type SimpleItem struct {
	ID           int `json:"id"`
	Title        string
	IVCiphertext string // base64 encoded
}

type Database struct {
	Name     string `survey:"name"`
	Password string `survey:"password"`
	Hint     string `survey:"comment"`
}
