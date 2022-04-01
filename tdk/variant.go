package tdk


type TDVariant struct {
	Variants []TDVariantElem `json:"variants"`
}

type TDVariantElem struct {
	Name   string    `json:"name"`
	Fields []TDField `json:"fields"`
}
