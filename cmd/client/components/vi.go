package components

import (
	"os"
	"os/exec"
)

func EditViaVI(text string) (string, error) {
	tmpFile, err := os.CreateTemp("", "text_*.txt")
	if err != nil {
		return "", err
	}

	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(text)
	if err != nil {
		return "", err
	}

	//nolint:gosec
	vi := exec.Command("vi", tmpFile.Name())
	vi.Stdin = os.Stdin
	vi.Stdout = os.Stdout

	err = vi.Run()
	if err != nil {
		return "", err
	}

	updatedText, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	return string(updatedText), nil
}
