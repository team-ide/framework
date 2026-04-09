package dialect_xx

import (
	"github.com/team-ide/framework/db"
)

var (
	cfgContent = ``
)

func init() {
	_, err := db.AddByDialectConfigBase64("xx", cfgContent)
	if err != nil {
		panic(err)
		return
	}
}
