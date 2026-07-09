package utility_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/services/utility"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/discord/commands"
)

type goldenFixture struct {
	Name    string            `json:"name"`
	Command string            `json:"command"`
	Input   map[string]string `json:"input"`
	Want    struct {
		Contains    []string `json:"contains"`
		NotContains []string `json:"not_contains"`
	} `json:"want"`
}

func TestUtilityGoldenHelp(t *testing.T) {
	var fixtures []goldenFixture
	readGolden(t, "../../../../testdata/features/utility_help_golden.json", &fixtures)
	service := utility.NewHelpService(commands.BuiltinRegistry(commands.Scope{Kind: commands.ScopeGlobal}))
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			content := service.Overview()
			if query := fixture.Input["指令名稱"]; query != "" {
				var err error
				content, err = service.Detail(query)
				if err != nil {
					t.Fatalf("detail: %v", err)
				}
			}
			assertGoldenContent(t, content, fixture)
		})
	}
}

func TestUtilityGoldenPing(t *testing.T) {
	var fixtures []goldenFixture
	readGolden(t, "../../../../testdata/features/utility_ping_golden.json", &fixtures)
	service := utility.PingService{Clock: fakeClock{now: time.Unix(100, 150_000_000)}}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			assertGoldenContent(t, service.Response(time.Unix(100, 0)), fixture)
		})
	}
}

func TestUtilityGoldenStatus(t *testing.T) {
	var fixtures []goldenFixture
	readGolden(t, "../../../../testdata/features/utility_status_golden.json", &fixtures)
	service := utility.StatusService{Provider: fakeBotInfoProvider{info: ports.BotInfo{
		Name:             "MHCAT",
		ShardCount:       1,
		GuildCount:       2,
		UserCount:        30,
		GatewayConnected: true,
	}}}
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			content, err := service.BotStatus(context.Background())
			if err != nil {
				t.Fatalf("status: %v", err)
			}
			assertGoldenContent(t, content, fixture)
		})
	}
}

func readGolden(t *testing.T, path string, out any) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", path, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		t.Fatalf("decode golden %s: %v", path, err)
	}
}

func assertGoldenContent(t *testing.T, content string, fixture goldenFixture) {
	t.Helper()
	for _, want := range fixture.Want.Contains {
		if !strings.Contains(content, want) {
			t.Fatalf("%s missing %q:\n%s", fixture.Name, want, content)
		}
	}
	for _, forbidden := range fixture.Want.NotContains {
		if strings.Contains(content, forbidden) {
			t.Fatalf("%s contained forbidden %q:\n%s", fixture.Name, forbidden, content)
		}
	}
}
