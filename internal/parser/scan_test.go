package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseStructureScan_SotiyoWithManufacturingRigs(t *testing.T) {
	scan := `High Power Slots
Standup Anticapital Missile Launcher II
Standup Multirole Missile Launcher II
Medium Power Slots
Standup Warp Scrambler II
Low Power Slots
Standup Armor Reinforcer II
Rig Slots
Standup XL-Set Ship Manufacturing Efficiency I
Standup XL-Set Structure and Component Manufacturing Efficiency I
Service Slots
Standup Capital Shipyard I
Standup Manufacturing Plant I
Standup Invention Lab I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "sotiyo", result.Structure)
	assert.Len(t, result.Rigs, 2)

	assert.Equal(t, "Standup XL-Set Ship Manufacturing Efficiency I", result.Rigs[0].Name)
	assert.Equal(t, "ship", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)

	assert.Equal(t, "Standup XL-Set Structure and Component Manufacturing Efficiency I", result.Rigs[1].Name)
	assert.Equal(t, "component", result.Rigs[1].Category)
	assert.Equal(t, "t1", result.Rigs[1].Tier)

	assert.Len(t, result.Services, 2)
	assert.Equal(t, "Standup Capital Shipyard I", result.Services[0].Name)
	assert.Equal(t, "manufacturing", result.Services[0].Activity)
	assert.Equal(t, "Standup Manufacturing Plant I", result.Services[1].Name)
	assert.Equal(t, "manufacturing", result.Services[1].Activity)
}

func Test_ParseStructureScan_RaitaruWithT2ShipRig(t *testing.T) {
	scan := `Rig Slots
Standup M-Set Ship Manufacturing Efficiency II
Service Slots
Standup Manufacturing Plant I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "raitaru", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "ship", result.Rigs[0].Category)
	assert.Equal(t, "t2", result.Rigs[0].Tier)

	assert.Len(t, result.Services, 1)
	assert.Equal(t, "manufacturing", result.Services[0].Activity)
}

func Test_ParseStructureScan_TataraWithReactionRigs(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Biochemical Reactor Efficiency II
Standup L-Set Composite Reactor Efficiency I
Service Slots
Standup Biochemical Reactor I
Standup Composite Reactor I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "tatara", result.Structure)
	assert.Len(t, result.Rigs, 2)

	assert.Equal(t, "reaction", result.Rigs[0].Category)
	assert.Equal(t, "t2", result.Rigs[0].Tier)

	assert.Equal(t, "reaction", result.Rigs[1].Category)
	assert.Equal(t, "t1", result.Rigs[1].Tier)

	assert.Len(t, result.Services, 2)
	assert.Equal(t, "reaction", result.Services[0].Activity)
	assert.Equal(t, "reaction", result.Services[1].Activity)
}

func Test_ParseStructureScan_AthanorWithPolymerReactorRig(t *testing.T) {
	scan := `Rig Slots
Standup M-Set Polymer Reactor Efficiency I
Service Slots
Standup Polymer Reactor I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "athanor", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "reaction", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)

	assert.Len(t, result.Services, 1)
	assert.Equal(t, "reaction", result.Services[0].Activity)
}

func Test_ParseStructureScan_AzbelWithEquipmentAndAmmoRigs(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Equipment Manufacturing Efficiency II
Standup L-Set Ammunition Manufacturing Efficiency I
Service Slots
Standup Manufacturing Plant I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "azbel", result.Structure)
	assert.Len(t, result.Rigs, 2)

	assert.Equal(t, "equipment", result.Rigs[0].Category)
	assert.Equal(t, "t2", result.Rigs[0].Tier)

	assert.Equal(t, "ammo", result.Rigs[1].Category)
	assert.Equal(t, "t1", result.Rigs[1].Tier)
}

func Test_ParseStructureScan_DroneRig(t *testing.T) {
	scan := `Rig Slots
Standup M-Set Drone and Fighter Manufacturing Efficiency I
Service Slots
Standup Manufacturing Plant I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "raitaru", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "drone", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)
}

func Test_ParseStructureScan_EmptyScan(t *testing.T) {
	result := ParseStructureScan("")

	assert.Equal(t, "", result.Structure)
	assert.Len(t, result.Rigs, 0)
	assert.Len(t, result.Services, 0)
}

func Test_ParseStructureScan_LabOnlyRigs(t *testing.T) {
	scan := `Rig Slots
Standup M-Set Laboratory Optimization I
Service Slots
Standup Research Lab I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "", result.Structure)
	assert.Len(t, result.Rigs, 0)
	assert.Len(t, result.Services, 0)
}

func Test_ParseStructureScan_ThukkerComponentRig(t *testing.T) {
	scan := `Rig Slots
Standup M-Set Thukker Component Manufacturing Efficiency I
Service Slots
Standup Manufacturing Plant I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "raitaru", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "thukker", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)
}

func Test_ParseStructureScan_MixedManufacturingAndReaction(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Ship Manufacturing Efficiency I
Standup L-Set Composite Reactor Efficiency II
Service Slots
Standup Manufacturing Plant I
Standup Composite Reactor I`

	result := ParseStructureScan(scan)

	// First rig is a manufacturing rig, so structure should be azbel
	assert.Equal(t, "azbel", result.Structure)
	assert.Len(t, result.Rigs, 2)

	assert.Equal(t, "ship", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)

	assert.Equal(t, "reaction", result.Rigs[1].Category)
	assert.Equal(t, "t2", result.Rigs[1].Tier)

	assert.Len(t, result.Services, 2)
	assert.Equal(t, "manufacturing", result.Services[0].Activity)
	assert.Equal(t, "reaction", result.Services[1].Activity)
}

func Test_ParseStructureScan_HybridReactorRig(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Hybrid Reactor Efficiency I
Service Slots
Standup Hybrid Reactor I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "tatara", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "reaction", result.Rigs[0].Category)
	assert.Equal(t, "t1", result.Rigs[0].Tier)
}

func Test_ParseStructureScan_GenericReactorRig(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Reactor Efficiency II
Service Slots
Standup Biochemical Reactor I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "tatara", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "reaction", result.Rigs[0].Category)
	assert.Equal(t, "t2", result.Rigs[0].Tier)

	assert.Len(t, result.Services, 1)
	assert.Equal(t, "reaction", result.Services[0].Activity)
}

func Test_ParseStructureScan_ReprocessingRig(t *testing.T) {
	scan := `Rig Slots
Standup L-Set Reprocessing Monitor II
Service Slots
Standup Manufacturing Plant I`

	result := ParseStructureScan(scan)

	assert.Equal(t, "tatara", result.Structure)
	assert.Len(t, result.Rigs, 1)
	assert.Equal(t, "Standup L-Set Reprocessing Monitor II", result.Rigs[0].Name)
	assert.Equal(t, "reprocessing", result.Rigs[0].Category)
	assert.Equal(t, "t2", result.Rigs[0].Tier)
}
