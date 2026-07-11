package commands

import "testing"

func TestPlanHasDangerous(t *testing.T) {
	if (Plan{}).HasDangerous() {
		t.Fatal("empty plan must not be dangerous")
	}
	if !(Plan{Operations: []PlannedOperation{{Risk: RiskHigh}}}).HasDangerous() {
		t.Fatal("high-risk plan must be dangerous")
	}
	if !(Plan{Operations: []PlannedOperation{{Operation: OperationDangerous}}}).HasDangerous() {
		t.Fatal("dangerous operation must be dangerous")
	}
}

func TestDefaultRegistryUsesBuiltinDefinitions(t *testing.T) {
	registry := DefaultRegistry(Scope{Kind: ScopeGlobal})
	if registry.Scope.Kind != ScopeGlobal || len(registry.Commands) == 0 {
		t.Fatalf("default registry = %#v", registry)
	}
}
