package command

import (
	"fmt"
	"path/filepath"
	"strings"
)

func Create(dir, linkName, packageType string) (command []string) {
	command = []string{filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))}
	if strings.HasSuffix(packageType, "jar") {
		command = []string{"java", "-jar", command[0]}
	}
	return
}
