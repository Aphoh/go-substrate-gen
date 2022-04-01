package tdk

const TDKComposite = "Composite"

type TDField struct {
	Name string `json:"name"`
	// This is the id of the type that this contains
	Type     string   `json:"type"`
	TypeName string   `json:"typeName"`
	Docs     []string `json:"docs"`
}

type TDComposite struct {
	Fields []TDField `json:"fields"`
}
