package dialect_kingbase-v8r6

import (
	"github.com/team-ide/framework/db"
)

var (
	cfgContent = `H4sIAAAAAAAA/2SRz2rjQAzG734K3XIyJn8JA4FlSU67C0vctPRUFI+wh3jGZjROW4zPPZc+Tl+nl75F8dhJ3NQn6adPsvSNtOpIVgQAAJKNgIMy6R6ZRBTVFZNtRF0i82NhZfOrzgp2LSmsa6JaosNW2/j2DHndAwHOVnSicZKRxgFTvLP5IDeo6fLn8Li0C89LtKi5W679kspaMu40b1Szj5rRWcGc60KSAKkY9zkF9OTIyO/Dp4HSmFI3Nym0xlYhi+RAFmxlIFROQhi2a10a57PpZAxhCT4QfUoQ38c3m38Pu3izXXXxgP6/W6/Gk+lsvgBHqJWk6Gzw+czu/Ou6LzlMBXihRpdk3cohfL68fry/3S63ix78USb9jUwD1Gmu6oF7Ln9Y3bJN75N/kxb8VewEmCrPgy8AAAD//wEAAP//u+R11SQCAAA=`
)

func init() {
	_, err := db.AddByDialectConfigBase64("kingbase-v8r6", cfgContent)
	if err != nil {
		panic(err)
		return
	}
}
