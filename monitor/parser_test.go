package monitor

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestParseNodeReport(t *testing.T) {
	data := []byte(jsonReportMock)
	var report map[string]interface{}
	err := json.Unmarshal(data, &report)
	if err != nil {
		t.Error("Failed to Unmarshal report")
	}

	nodeReport := ParseNodeReport(report)
	if len(nodeReport.Nodes) != 3 {
		t.Error("Wrong node count in report")
	}

	if data, _ := json.Marshal(&nodeReport); err == nil {
		fmt.Printf("%s\n", string(data))
	}
}
