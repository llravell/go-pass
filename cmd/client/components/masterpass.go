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
)

var ErrEmptyMasterPassword = errors.New("got empty master password")

type MasterPassPrompter struct {
	authUC *usecase.AuthUseCase
}

func NewMasterPassPrompter(authUC *usecase.AuthUseCase) *MasterPassPrompter {
	return &MasterPassPrompter{
		authUC: authUC,
	}
}

func (p *MasterPassPrompter) PromptAndValidate(ctx context.Context) (string, error) {
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
