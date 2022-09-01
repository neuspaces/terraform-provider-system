package systemd

const PropertyLoadState = "LoadState"

type LoadState string

const (
	LoadStateNotFound LoadState = "not-found"

	LoadStateLoaded LoadState = "loaded"

	LoadStateError LoadState = "error"

	LoadStateMasked LoadState = "masked"
)

const PropertyActiveState = "ActiveState"

type ActiveState string

const (
	// ActiveStateActive indicates that unit is active.
	ActiveStateActive ActiveState = "active"

	// ActiveStateReloading indicates that the unit is active and currently reloading its configuration.
	ActiveStateReloading ActiveState = "reloading"

	// ActiveStateInactive indicates that it is inactive and the previous run was successful or no previous run has taken place yet.
	ActiveStateInactive ActiveState = "inactive"

	// ActiveStateFailed indicates that it is inactive and the previous run was not successful.
	ActiveStateFailed ActiveState = "failed"

	// ActiveStateActivating indicates that the unit has previously been inactive but is currently in the process of entering an active state.
	ActiveStateActivating ActiveState = "activating"

	// ActiveStateDeactivating indicates that the unit is currently in the process of deactivation.
	ActiveStateDeactivating ActiveState = "deactivating"
)
