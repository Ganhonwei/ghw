package table

import (
	"encoding/json"
	"goserver/gen/tb"
	"io/ioutil"
	"sync"
)

var (
	tables      *tb.Tables
	tablesMutex sync.RWMutex
)

func loader(file string) ([]map[string]interface{}, error) {
	if bytes, err := ioutil.ReadFile("config/table/json/" + file + ".json"); err != nil {
		return nil, err
	} else {
		jsonData := make([]map[string]interface{}, 0)
		if err = json.Unmarshal(bytes, &jsonData); err != nil {
			return nil, err
		}
		return jsonData, nil
	}
}

func LoadTables() error {
	tablesMutex.Lock()
	defer tablesMutex.Unlock()

	if newTables, err := tb.NewTables(loader); err != nil {
		return err
	} else {
		tables = newTables
		return nil
	}
}

func GetTables() *tb.Tables {
	tablesMutex.RLock()
	defer tablesMutex.RUnlock()
	return tables
}
