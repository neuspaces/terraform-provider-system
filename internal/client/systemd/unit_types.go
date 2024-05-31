package systemd

type UnitType string

const (
	UnitTypeService UnitType = "service"

	UnitTypeSocket UnitType = "socket"

	UnitTypeDevice UnitType = "device"

	UnitTypeMount UnitType = "mount"

	UnitTypeAutoMount UnitType = "automount"

	UnitTypeSwap UnitType = "swap"

	UnitTypeTarget UnitType = "target"

	UnitTypePath UnitType = "path"

	UnitTypeTimer UnitType = "timer"

	UnitTypeSlide UnitType = "slice"

	UnitTypeScope UnitType = "scope"
)
