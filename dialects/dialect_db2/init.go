package dialect_db2

import (
	"github.com/team-ide/framework/db"
)

var (
	cfgContent = `H4sIAAAAAAAA/0yRT4vbMBDF7/4UQy46KXXSphQZU7xrQwPbJKy95FIwY0l1RC1bSEr/YPzdi/+0sU6j37yRnp7Q1s83tAzIZxIIq35KywIAAOFaBm/HNO7vTtohulzTuDfo3K/OiiH6cs6LU/I1i/tb5/wQXc6vRdybzvohmuaVe7MNg+/YODmBFrVkUHelqnQpqokZtKhzg1zOJqIHdbONcaVJkTwlecaA9AI9VujkQAL528tWMNhsAqWxlvMA77TGEYuO/5AW7L0FqrwASkcHIKo9PYRhGAI1MBVs2Up4OT5npzyLkXNp/EjSp/3xlBe78pLk+fX8msa7/fsPh49zbwrAS+dLUYGqNO/0O1Ht2W63PWw/bcPVux/dCXqsGfyXafT8Ntun8E9CYRUUXQU3Hni1aJZ/25DAeavaOnMczRLkt4WtdISQwP8xkk03jNWLcp5Be2+a4C8AAAD//wEAAP//FgqjCwcCAAA=`
)

func init() {
	_, err := db.AddByDialectConfigBase64("db2", cfgContent)
	if err != nil {
		panic(err)
		return
	}
}
