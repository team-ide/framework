package dialect_opengauss

import (
	"github.com/team-ide/framework/db"
)

var (
	cfgContent = `H4sIAAAAAAAA/4yUXW/TPBTH7/spzjM9FyDRlW7jxlKlAQvjYrBpBe2yOnNOG4vYDj7O1inkuyM7aerQgZab5Pz9O++Rc6ceyIkJAIBEc4O+yFhiRQK8qynqORsBtiKzwZpZzGZNzeRa0VTI/Ghd3p43hWUfFOt8O2ty9HiPTG30L5AveiGJWiAvZUEaBayx5E5U/N2VCWRQU5I7ahU61NyVHB7mUtucBOSK8b6kCW09mVxAZdlvHPFEadxQ36PVGsNhbuUPcuBqA1Plc5hOQ659qum7s9MTmFYQ3qIzCC6Xq5v3y+Xd9e3FIjPans9PTs+AjLaeZDHbD+n0eH78NmnhEImHHjcCOlajl0VX5BQuB6L/vvjQW9cVmYPTCdPPmozcNekIPQn4NcyoU2CHQROqunNYtQMSR+nReXhUvoAmfo+PlZGONBkP90/QDNYYMha0Mg9Y9itMddwe6hJlQTCPGlNJ0qeFL7Or7OO3xGHXwkqix9JuYPevrUJLb54jOf5m0L3+jsX9j6wRFsaxiuUnkYOYUloZpWv9J6eVGVG4fZbCbUrtpz0QgzRgn26vv4Aya+s0emVN3+zxzoUH8u5zdpvBfDFPU6xf7cb3NXT/3wKOjl43o/2gyQ+HvjiQEp92nKGr6GXx+1Ut4P9m79b+I3bv+LLoccMxduI2ju6fqtGFE+ysv0/irRSEK8VegKnLcvIbAAD//wEAAP//ze73WkAFAAA=`
)

func init() {
	_, err := db.AddByDialectConfigBase64("opengauss", cfgContent)
	if err != nil {
		panic(err)
		return
	}
}
