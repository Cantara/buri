package pack

import (
	"fmt"
	"strings"

	log "github.com/cantara/bragi/sbragi"
)

type Type string

const (
	Jar = Type("jar")
	Go  = Type("go")
	Tar = Type("tar")
	Zip = Type("zip")
)

func (s *Type) String() string {
	return fmt.Sprint(*s)
}

func TypeFromString(s string) (pt Type) {
	switch strings.ToLower(s) {
	case "java":
		pt = Jar
	case "jar":
		pt = Jar
	case "go":
		pt = Go
	case "tgz":
		pt = Tar
	case "tar":
		pt = Tar
	case "zip":
		pt = Zip
	default:
		//err = errors.New("unsuported service type")
		log.Info("service type not found. treating as website / frontend") //Could be smart to return to error and use tag website and artifact for name of website
		pt = Type(fmt.Sprintf("website_%s", s))
	}
	return
}

func (t Type) TrimExtention(s string) string {
	switch t {
	case Jar:
		return strings.TrimSuffix(s, ".jar")
	case Tar:
		return strings.TrimSuffix(s, ".tgz")
	case Zip:
		return strings.TrimSuffix(s, ".zip")
	}
	return s
}
