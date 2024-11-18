package base

import (
	"encoding/json"
	"os"
)

func (db *DbTopLevel) SaveDb(fpath string) bool {
	// Save as JSON
	j, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		Error.Println(err)
		return false
	}
	if err := os.WriteFile(fpath, j, 0666); err != nil {
		Error.Println(err)
		return false
	}
	return true
}
