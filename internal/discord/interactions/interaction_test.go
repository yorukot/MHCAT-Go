package interactions

import "testing"

func TestActorHasPermissionPreservesAdministratorOverride(t *testing.T) {
	const (
		administrator  = int64(1 << 3)
		manageMessages = int64(1 << 13)
	)
	if !(Actor{PermissionBits: administrator}).HasPermission(manageMessages) {
		t.Fatal("administrator should satisfy manage messages like Discord.js")
	}
	if !(Actor{PermissionBits: manageMessages}).HasPermission(manageMessages) {
		t.Fatal("explicit manage messages permission should be accepted")
	}
	if (Actor{PermissionBits: administrator}).HasPermission(0) {
		t.Fatal("zero must not become a valid permission request")
	}
	if (Actor{}).HasPermission(manageMessages) {
		t.Fatal("missing permission should be rejected")
	}
}
