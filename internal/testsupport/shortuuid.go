package testsupport

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ShortUUID returns a short uuid
func ShortUUID() string {
	return strings.Split(uuid.New().String(), "-")[0]
}

// UnittestStage returns the name of the stage to use in unit test
func UnittestStage() string {
	return fmt.Sprintf("unittest-%s", ShortUUID())
}
