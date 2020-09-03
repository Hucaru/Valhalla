package item

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// DropTableEntry contains all the information needed about a drop item or mesos
type DropTableEntry struct {
	IsMesos bool  `json:"isMesos"`
	ItemID  int32 `json:"itemId"`
	Min     int32 `json:"min"`
	Max     int32 `json:"max"`
	QuestID int32 `json:"questId"` // TODO: Validate this
	Chance  int32 `json:"chance"`
}

// DropTable is the global lookup table for drops
var DropTable map[int32][]DropTableEntry

// PopulateDropTable from a json file
func PopulateDropTable(dropJSON string) error {
	jsonFile, err := os.Open(dropJSON)

	if err != nil {
		return fmt.Errorf(err.Error())
	}

	defer jsonFile.Close()

	jsonBytes, _ := ioutil.ReadAll(jsonFile)

	json.Unmarshal(jsonBytes, &DropTable)

	return nil
}
