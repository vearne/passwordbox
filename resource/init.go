package resource

import (
	"github.com/vearne/passwordbox/model"
)

var (
	DataPath           string
	MaxBackupFileCount int
)

var (
	GlobalOSS model.ObjectStorage
)

var (
	LoopExit = false
)
