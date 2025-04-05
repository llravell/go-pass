package entity

import pb "github.com/llravell/go-pass/pkg/grpc"

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

func (pass *Password) ToPB() *pb.Password {
	return &pb.Password{
		Name:    pass.Name,
		Value:   pass.Value,
		Meta:    pass.Meta,
		Version: int32(pass.Version), //nolint:gosec
	}
}

func NewPasswordFromPB(password *pb.Password) *Password {
	return &Password{
		Name:    password.GetName(),
		Value:   password.GetValue(),
		Meta:    password.GetMeta(),
		Version: int(password.GetVersion()),
	}
}
