package command

import (
	"fmt"
	"path/filepath"

	"github.com/cantara/buri/pack"
)

func Create(dir, linkName string, packageType pack.Type) (command []string) {
	command = []string{filepath.Clean(fmt.Sprintf("%s/%s", dir, linkName))}
	if packageType == pack.Jar {
		command = []string{"java", "-jar", command[0]}
	}
	return
}
