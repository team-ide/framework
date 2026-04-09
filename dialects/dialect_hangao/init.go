package dialect_hangao

import (
	"github.com/team-ide/framework/db"
)

var (
	cfgContent = `H4sIAAAAAAAA/1SQzUrFMBSE932K8wJy91mJ3IsuXAjiA0ybYxPITznn1B9K312aFq1ZZT5mMkO8xA8W1xERDSgvsHDTARM7Mpm5ca/F0VTVRmF1l8syK8vqlgmqn1X8er+EqraRKrZeFg9DD+W1xQP0eoDTowH6OgTOcPSOpDuM+ibpZCrI/Ffd0ARB1n3wdlRTrp4d+ajoE3f8ZVz8KRUzRnZU5pS6DBvCHr6jJ5RH1EOE8ffi+3+G60Nn39uHBJQRtYnbUdKWbuA5qh0dPwAAAP//AQAA//9FN8XRUwEAAA==`
)

func init() {
	_, err := db.AddByDialectConfigBase64("hangao", cfgContent)
	if err != nil {
		panic(err)
		return
	}
}
