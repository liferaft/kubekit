package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// HardConfirmation ask to the user a question, the answer has to be typed, it's
// not defaulted. If the received answer is same as the expected one, the
// confirmation succeed
func HardConfirmation(question, answer string) bool {
	answer = strings.ToLower(answer)
	question = fmt.Sprintf("%s [type '%s']?: ", question, answer)

	reader := bufio.NewReader(os.Stdin)
	fmt.Printf(question)
	resp, _ := reader.ReadString('\n')

	return strings.ToLower(strings.TrimSuffix(resp, "\n")) == answer
}
