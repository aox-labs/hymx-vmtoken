package utils

import (
	"encoding/json"
	"fmt"
	hymxSchema "github.com/hymatrix/hymx/schema"
	"os"
)

// generate aox vm token module
func generateModule(moduleFormat string) {
	item, _ := s.GenModule([]byte{}, hymxSchema.Module{
		Base:         hymxSchema.DefaultBaseModule,
		ModuleFormat: moduleFormat,
	})

	itemBy, _ := json.Marshal(item)

	filename := fmt.Sprintf("mod-%s.json", item.Id)
	fmt.Println("generated module file: ", filename)
	os.WriteFile(filename, itemBy, 0644)
}
