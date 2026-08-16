package visitors

import "testing"

func TestVisitorStatusValues(t *testing.T) {
	if StatusActive == StatusBlacklisted {
		t.Fatal("visitor active and blacklisted states must remain distinct")
	}
}
