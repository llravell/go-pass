package components

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strings"
)

var ErrInvalidUserResponse = errors.New("invalid response")

func isPositiveResponse(response string) bool {
	return response == "y" || response == "yes"
}

func isNegativeResponse(response string) bool {
	return response == "n" || response == "no" || response == "not"
}

func BoolPrompt(text string) (bool, error) {
	writer := bufio.NewWriter(os.Stdout)

	_, err := writer.WriteString(text + "\n")
	if err != nil {
		return false, err
	}

	_, err = writer.WriteString("y/n: ")
	if err != nil {
		return false, err
	}

	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	userResponse := strings.ToLower(strings.TrimSpace(input))

	if isPositiveResponse(userResponse) {
		return true, nil
	}

	if isNegativeResponse(userResponse) {
		return false, nil
	}

	return false, ErrInvalidUserResponse
}
