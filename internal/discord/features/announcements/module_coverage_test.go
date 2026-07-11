package announcements

import (
	"errors"
	"testing"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/testutil/fakemongo"
)

func TestAnnouncementModuleMetadataAndCombinedConstructors(t *testing.T) {
	repository := fakemongo.NewAnnouncementConfigRepository()
	modules := []Module{
		NewModuleWithSend(repository, nil),
		NewModuleWithRelay(repository, nil),
	}
	for _, module := range modules {
		if module.Name() != "announcement-config" || len(module.Commands()) == 0 {
			t.Fatalf("announcement metadata name=%q commands=%d", module.Name(), len(module.Commands()))
		}
	}
	message := announcementUnknownError(errors.New("hidden"))
	if len(message.Embeds) != 1 || message.Embeds[0].Title == "" {
		t.Fatalf("announcement error message = %#v", message)
	}
}

func TestAnnouncementDraftDeleteRemovesEntry(t *testing.T) {
	store := NewDraftStore()
	id, err := store.Put(AnnouncementDraft{Title: "draft"})
	if err != nil {
		t.Fatalf("put draft: %v", err)
	}
	store.Delete(id)
	if _, err := store.Take(id); !errors.Is(err, ErrAnnouncementDraftNotFound) {
		t.Fatalf("take deleted draft: %v", err)
	}
}
