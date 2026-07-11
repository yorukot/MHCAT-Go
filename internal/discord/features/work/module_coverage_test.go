package work

import "testing"

func TestWorkModuleMetadataAndRandomHelpers(t *testing.T) {
	module := NewModule(nil)
	if module.Name() != "work" || len(module.Commands()) == 0 {
		t.Fatalf("work module name=%q commands=%d", module.Name(), len(module.Commands()))
	}
	message := legacyWorkErrorMessage("failure")
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("work error message = %#v", message)
	}
	left, right := randomCaptcha()
	if left < 0 || left > 9 || right < 0 || right > 9 {
		t.Fatalf("captcha = %d + %d", left, right)
	}
	if digit := secureDigit(); digit < 0 || digit > 9 {
		t.Fatalf("secure digit = %d", digit)
	}
}
