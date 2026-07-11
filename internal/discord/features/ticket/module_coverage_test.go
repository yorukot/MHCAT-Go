package ticket

import (
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestTicketModuleMetadata(t *testing.T) {
	module := NewModule(fakemongo.NewTicketConfigRepository())
	if module.Name() != "ticket" || len(module.Commands()) == 0 {
		t.Fatalf("ticket metadata name=%q commands=%d", module.Name(), len(module.Commands()))
	}
}

func TestTicketPanelMessage(t *testing.T) {
	message := ticketPanelMessage("Support", "Open a ticket", 0x123456)
	if len(message.Embeds) != 1 || len(message.Components) != 1 || message.Embeds[0].Title != "Support" {
		t.Fatalf("ticket panel message = %#v", message)
	}
}
