package components

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	usecase "github.com/llravell/go-pass/internal/usecase/client"
	"github.com/llravell/go-pass/pkg/encryption"
)

var ErrEmptyMasterPassword = errors.New("got empty master password")

type EncryptionKeyProvider struct {
	authUC *usecase.AuthUseCase
	key    *encryption.Key
}

func NewEncryptionKeyProvider(authUC *usecase.AuthUseCase) *EncryptionKeyProvider {
	return &EncryptionKeyProvider{
		authUC: authUC,
	}
}

func (p *EncryptionKeyProvider) promptMasterPassword(ctx context.Context) (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter master password: ")
	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}

	masterPassword := strings.TrimSpace(input[:len(input)-1])
	if len(masterPassword) == 0 {
		return "", ErrEmptyMasterPassword
	}

	err = p.authUC.ValidateMasterPassword(ctx, masterPassword)
	if err != nil {
		return "", err
	}

	return masterPassword, nil
}

func (p *EncryptionKeyProvider) Get(ctx context.Context) (*encryption.Key, error) {
	if p.key != nil {
		return p.key, nil
	}

	masterPassword, err := p.promptMasterPassword(ctx)
	if err != nil {
		return nil, err
	}

	err = p.authUC.ValidateMasterPassword(ctx, masterPassword)
	if err != nil {
		return nil, err
	}

	p.key = encryption.GenerateKeyFromMasterPass(masterPassword)

	return p.key, nil
}
