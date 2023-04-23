package version

import "errors"

type Type string

const Base = Type("release")

var ErrTypeDoesNotExist = errors.New("version type does not exists")
