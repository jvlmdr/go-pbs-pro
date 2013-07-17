package grideng

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// Reads task number from environment variable key.
// Variable might be GE_TASK_ID, SGE_TASK_ID, etc.
func TaskNumFromEnv(key string) (int, error) {
	str := os.Getenv(key)
	if len(str) == 0 {
		what := fmt.Sprintf("Environment variable %s is empty", key)
		return 0, errors.New(what)
	}

	num, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, err
	}
	return int(num), nil
}
