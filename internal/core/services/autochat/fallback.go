package autochat

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

//go:embed legacy_chat.json
var legacyChatJSON []byte

type legacyResponse struct {
	key   string
	value string
}

type FallbackService struct {
	configs   ports.AutoChatConfigReader
	balances  ports.BalanceRepository
	responses []legacyResponse
}

func NewFallbackService(configs ports.AutoChatConfigReader, balances ports.BalanceRepository) (FallbackService, error) {
	if configs == nil || balances == nil {
		return FallbackService{}, errors.New("autochat fallback repositories are required")
	}
	responses, err := decodeLegacyResponses(legacyChatJSON)
	if err != nil {
		return FallbackService{}, fmt.Errorf("decode legacy autochat responses: %w", err)
	}
	return FallbackService{configs: configs, balances: balances, responses: responses}, nil
}

func (s FallbackService) Reply(ctx context.Context, guildID string, channelID string, content string) (domain.AutoChatFallbackReply, error) {
	if err := ctx.Err(); err != nil {
		return domain.AutoChatFallbackReply{}, err
	}
	guildID = strings.TrimSpace(guildID)
	channelID = strings.TrimSpace(channelID)
	if guildID == "" || channelID == "" {
		return domain.AutoChatFallbackReply{}, nil
	}
	config, err := s.configs.GetAutoChatConfig(ctx, guildID)
	if errors.Is(err, ports.ErrAutoChatConfigMissing) {
		return domain.AutoChatFallbackReply{}, nil
	}
	if err != nil {
		return domain.AutoChatFallbackReply{}, err
	}
	if strings.TrimSpace(config.ChannelID) != channelID {
		return domain.AutoChatFallbackReply{}, nil
	}

	balance, err := s.balances.GetBalance(ctx, guildID)
	if err != nil && !errors.Is(err, ports.ErrBalanceMissing) {
		return domain.AutoChatFallbackReply{}, err
	}
	if err == nil {
		amount, parseErr := strconv.ParseFloat(strings.TrimSpace(balance.Amount), 64)
		if parseErr == nil && amount >= 0 {
			return domain.AutoChatFallbackReply{}, nil
		}
	}
	return s.localReply(content), ctx.Err()
}

func (s FallbackService) localReply(content string) domain.AutoChatFallbackReply {
	if strings.Contains(content, "說出") {
		stripped := strings.NewReplacer("說", "", "出", "").Replace(content)
		switch {
		case stripped == "":
			return domain.AutoChatFallbackReply{Content: "說出甚麼?"}
		case strings.Contains(content, "幹"):
			return domain.AutoChatFallbackReply{Content: "很抱歉，讀取到你說出了一些不好的字元，因此拒絕說出w\n字元:"}
		case strings.Contains(stripped, "我"):
			return domain.AutoChatFallbackReply{Content: `"` + strings.Replace(stripped, "我", "你", 1) + `"`}
		default:
			return domain.AutoChatFallbackReply{Content: stripped}
		}
	}

	best := ""
	highest := 0.0
	for _, response := range s.responses {
		probability := legacySimilarity(response.key, content)
		if probability > highest {
			highest = probability
			best = response.value
		}
	}
	if best == "" {
		return domain.AutoChatFallbackReply{Content: "我看不懂你的意思，在講一次好不好w"}
	}
	return domain.AutoChatFallbackReply{Content: best, UseTypingDelay: true}
}

func decodeLegacyResponses(data []byte) ([]legacyResponse, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil {
		return nil, err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '{' {
		return nil, errors.New("legacy response corpus must be an object")
	}
	var indexed []struct {
		index    uint64
		response legacyResponse
	}
	responses := make([]legacyResponse, 0, 384)
	for decoder.More() {
		keyToken, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyToken.(string)
		if !ok {
			return nil, errors.New("legacy response key must be a string")
		}
		var value string
		if err := decoder.Decode(&value); err != nil {
			return nil, err
		}
		response := legacyResponse{key: key, value: value}
		if index, ok := javascriptArrayIndex(key); ok {
			indexed = append(indexed, struct {
				index    uint64
				response legacyResponse
			}{index: index, response: response})
			continue
		}
		responses = append(responses, response)
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	sort.Slice(indexed, func(i int, j int) bool { return indexed[i].index < indexed[j].index })
	ordered := make([]legacyResponse, 0, len(indexed)+len(responses))
	for _, entry := range indexed {
		ordered = append(ordered, entry.response)
	}
	return append(ordered, responses...), nil
}

func javascriptArrayIndex(value string) (uint64, bool) {
	if value == "" || len(value) > 1 && value[0] == '0' {
		return 0, false
	}
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil || parsed == math.MaxUint32 || strconv.FormatUint(parsed, 10) != value {
		return 0, false
	}
	return parsed, true
}

func legacySimilarity(left string, right string) float64 {
	leftUnits := utf16.Encode([]rune(strings.ToLower(left)))
	rightUnits := utf16.Encode([]rune(strings.ToLower(right)))
	longer := leftUnits
	shorter := rightUnits
	if len(leftUnits) < len(rightUnits) {
		longer, shorter = rightUnits, leftUnits
	}
	if len(longer) == 0 {
		return 1
	}
	return float64(len(longer)-legacyEditDistance(longer, shorter)) / float64(len(longer))
}

func legacyEditDistance(left []uint16, right []uint16) int {
	costs := make([]int, len(right)+1)
	for j := range costs {
		costs[j] = j
	}
	for i := 1; i <= len(left); i++ {
		previous := costs[0]
		costs[0] = i
		for j := 1; j <= len(right); j++ {
			old := costs[j]
			cost := 0
			if left[i-1] != right[j-1] {
				cost = 1
			}
			costs[j] = min(costs[j]+1, costs[j-1]+1, previous+cost)
			previous = old
		}
	}
	return costs[len(right)]
}
