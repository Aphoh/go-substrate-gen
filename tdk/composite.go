package tdk


type TDField struct {
	Name     string   `json:"name"`
	TypeId   string   `json:"type"` // This is the id of the type that this contains
	TypeName string   `json:"typeName"`
	Index    string   `json:"index"`
	Docs     []string `json:"docs"`
}

type TDComposite struct {
	Fields []TDField `json:"fields"`
}
