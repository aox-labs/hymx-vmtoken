package utils

import (
	"github.com/aox-labs/hymx-vmtoken/vmtoken/schema"
	"testing"
)

func Test_Generate_Module(t *testing.T) {
	// generate basic token module

	generateModule(schema.VmTokenBasicModuleFormat)           // mod-9bQh650l10NZ7GHUvj1L_kIIiivp9Zj7kJNY3CLEcRM.json
	generateModule(schema.VmTokenCrossChainModuleFormat)      // mod-QW_l2HiEgurKA-_gxe5JXYYuqpkFwyps_V1RjGO86-c.json
	generateModule(schema.VmTokenCrossChainMultiModuleFormat) // mod-PxAJFkNuJqcRKB6475tqqT2R0G1OT-0KDcqHMbejV84.json
}
