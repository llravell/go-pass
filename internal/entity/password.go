package entity

type Password struct {
	Name    string
	Value   string
	Meta    string
	Version int
	Deleted bool
}

func (pass *Password) BumpVersion() {
	pass.Version++
}
