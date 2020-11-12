# mdb


a simple database that can persist to disk or run in ram


---


to persist to disk:
```golang
	db, closeDB, err := mdb.Open(
		&mdb.Config{
			Path: "some/file/path.db",
		},
	)
```

and example in ram:
```golang
package main

import (
	"log"
	"github.com/cbluth/mdb"

)

func main() {
	// open db in ram
	db, closeDB, err := mdb.Open(nil)
	if err != nil {
		log.Fatalln(err)
	}
	defer closeDB()
	
	// set some key-value pairs
	db.SetKV("animal",
		[]mdb.KV{
			{Key: "hungry", Value: "yes"},
			{Key: "hippo", Value: "yes"},
			{Key: "location", Value: "river"},
		},
	)

	// set a map
	m := map[string]string{}
	m["hungry"] = "yes"
	m["hippo"] = "yes"
	m["location"] = "river"
	db.SetMap("animal", m)

	log.Println(
		db.Get("animal"),
	)
}
```

output:
```
2020/11/12 17:23:32 [{hippo yes} {hungry yes} {location river}] <nil>
```

---

TODO:
- add compression to the db file
- add auto-sync policy (auto-save to disk)
