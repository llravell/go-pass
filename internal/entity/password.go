package entity

import (
	"github.com/llravell/go-pass/pkg/encryption"
	pb "github.com/llravell/go-pass/pkg/grpc"
)

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

func (pass *Password) Open(key *encryption.Key) error {
	decryptedValue, err := key.Decrypt(pass.Value)
	if err != nil {
		return err
	}

	pass.Value = decryptedValue

	return nil
}

func (pass *Password) Close(key *encryption.Key) error {
	encryptedValue, err := key.Encrypt(pass.Value)
	if err != nil {
		return err
	}

	pass.Value = encryptedValue

	return nil
}

func (pass *Password) Equal(target *Password) bool {
	return (pass.Name == target.Name &&
		pass.Meta == target.Meta &&
		pass.Version == target.Version)
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
