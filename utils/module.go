package utils

import (
	"encoding/json"
	"fmt"
	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	hymxSchema "github.com/hymatrix/hymx/schema"
	"os"
)

// generate aox vm token module
func generateModule() {
	item, _ := s.GenerateModule([]byte{}, hymxSchema.Module{
		Base:         hymxSchema.DefaultBaseModule,
		ModuleFormat: schema.VmTokenModuleFormat,
	})

	itemBy, _ := json.Marshal(item)

	filename := fmt.Sprintf("mod-%s.json", item.Id)
	fmt.Println("generated module file: ", filename)
	os.WriteFile(filename, itemBy, 0644)
}
