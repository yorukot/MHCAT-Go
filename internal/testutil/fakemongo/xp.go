package fakemongo

import (
	"context"
	"strings"

	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/domain"
	"github.com/yorukot/MHCAT/MHCAT-REFACTOR/internal/core/ports"
)

type TextXPConfigRepository struct {
	Configs map[string]domain.TextXPConfig
	Err     error
}

type VoiceXPConfigRepository struct {
	Configs map[string]domain.VoiceXPConfig
	Err     error
}

type TextXPRewardRoleRepository struct {
	Configs []domain.XPRewardRoleConfig
	Err     error
}

type VoiceXPRewardRoleRepository struct {
	Configs []domain.XPRewardRoleConfig
	Err     error
}

type XPAdminRepository struct {
	TextProfiles  map[string]domain.XPProfile
	VoiceProfiles map[string]domain.XPProfile
	Err           error
}

func NewTextXPConfigRepository() *TextXPConfigRepository {
	return &TextXPConfigRepository{Configs: map[string]domain.TextXPConfig{}}
}

func NewVoiceXPConfigRepository() *VoiceXPConfigRepository {
	return &VoiceXPConfigRepository{Configs: map[string]domain.VoiceXPConfig{}}
}

func NewTextXPRewardRoleRepository() *TextXPRewardRoleRepository {
	return &TextXPRewardRoleRepository{}
}

func NewVoiceXPRewardRoleRepository() *VoiceXPRewardRoleRepository {
	return &VoiceXPRewardRoleRepository{}
}

func NewXPAdminRepository() *XPAdminRepository {
	return &XPAdminRepository{
		TextProfiles:  map[string]domain.XPProfile{},
		VoiceProfiles: map[string]domain.XPProfile{},
	}
}

func (r *TextXPConfigRepository) SaveTextXPConfig(ctx context.Context, config domain.TextXPConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.Configs[strings.TrimSpace(config.GuildID)] = config
	return nil
}

func (r *TextXPConfigRepository) GetTextXPConfig(ctx context.Context, guildID string) (domain.TextXPConfig, error) {
	if err := r.ready(ctx); err != nil {
		return domain.TextXPConfig{}, err
	}
	guildID = strings.TrimSpace(guildID)
	config, ok := r.Configs[guildID]
	if !ok {
		return domain.TextXPConfig{}, ports.ErrTextXPConfigMissing
	}
	return config, nil
}

func (r *TextXPConfigRepository) DeleteTextXPConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrTextXPConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *VoiceXPConfigRepository) SaveVoiceXPConfig(ctx context.Context, config domain.VoiceXPConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	r.Configs[strings.TrimSpace(config.GuildID)] = config
	return nil
}

func (r *VoiceXPConfigRepository) DeleteVoiceXPConfig(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	if _, ok := r.Configs[guildID]; !ok {
		return ports.ErrVoiceXPConfigMissing
	}
	delete(r.Configs, guildID)
	return nil
}

func (r *TextXPRewardRoleRepository) ListTextXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return filterRewardRoles(r.Configs, guildID), nil
}

func (r *TextXPRewardRoleRepository) SaveTextXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	config = config.Normalize()
	r.Configs = deleteRewardRole(r.Configs, config.GuildID, config.Level, config.RoleID)
	r.Configs = append(r.Configs, config)
	return nil
}

func (r *TextXPRewardRoleRepository) DeleteTextXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	before := len(r.Configs)
	r.Configs = deleteRewardRole(r.Configs, guildID, level, roleID)
	if len(r.Configs) == before {
		return ports.ErrTextXPRewardRoleMissing
	}
	return nil
}

func (r *VoiceXPRewardRoleRepository) ListVoiceXPRewardRoles(ctx context.Context, guildID string) ([]domain.XPRewardRoleConfig, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return filterRewardRoles(r.Configs, guildID), nil
}

func (r *VoiceXPRewardRoleRepository) SaveVoiceXPRewardRole(ctx context.Context, config domain.XPRewardRoleConfig) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	config = config.Normalize()
	r.Configs = deleteRewardRole(r.Configs, config.GuildID, config.Level, config.RoleID)
	r.Configs = append(r.Configs, config)
	return nil
}

func (r *VoiceXPRewardRoleRepository) DeleteVoiceXPRewardRole(ctx context.Context, guildID string, level int64, roleID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	before := len(r.Configs)
	r.Configs = deleteRewardRole(r.Configs, guildID, level, roleID)
	if len(r.Configs) == before {
		return ports.ErrVoiceXPRewardRoleMissing
	}
	return nil
}

func (r *XPAdminRepository) GetTextXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return domain.XPProfile{}, err
	}
	profile, ok := r.TextProfiles[xpProfileKey(guildID, userID)]
	if !ok {
		return domain.XPProfile{}, ports.ErrTextXPProfileMissing
	}
	return profile, nil
}

func (r *XPAdminRepository) SaveTextXPProfile(ctx context.Context, profile domain.XPProfile) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	profile = profile.Normalize()
	r.TextProfiles[xpProfileKey(profile.GuildID, profile.UserID)] = profile
	return nil
}

func (r *XPAdminRepository) GetVoiceXPProfile(ctx context.Context, guildID string, userID string) (domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return domain.XPProfile{}, err
	}
	profile, ok := r.VoiceProfiles[xpProfileKey(guildID, userID)]
	if !ok {
		return domain.XPProfile{}, ports.ErrVoiceXPProfileMissing
	}
	return profile, nil
}

func (r *XPAdminRepository) SaveVoiceXPProfile(ctx context.Context, profile domain.XPProfile) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	profile = profile.Normalize()
	key := xpProfileKey(profile.GuildID, profile.UserID)
	if profile.LeaveJoin == "" {
		if existing, ok := r.VoiceProfiles[key]; ok && strings.TrimSpace(existing.LeaveJoin) != "" {
			profile.LeaveJoin = strings.TrimSpace(existing.LeaveJoin)
		} else {
			profile.LeaveJoin = domain.VoiceXPSessionLeft
		}
	}
	r.VoiceProfiles[key] = profile
	return nil
}

func (r *XPAdminRepository) MarkVoiceXPJoined(ctx context.Context, guildID string, userID string) error {
	return r.markVoiceXPSession(ctx, guildID, userID, domain.VoiceXPSessionJoined)
}

func (r *XPAdminRepository) MarkVoiceXPLeft(ctx context.Context, guildID string, userID string) error {
	return r.markVoiceXPSession(ctx, guildID, userID, domain.VoiceXPSessionLeft)
}

func (r *XPAdminRepository) ListTextXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return filterXPProfiles(r.TextProfiles, guildID), nil
}

func (r *XPAdminRepository) ListVoiceXPProfiles(ctx context.Context, guildID string) ([]domain.XPProfile, error) {
	if err := r.ready(ctx); err != nil {
		return nil, err
	}
	return filterXPProfiles(r.VoiceProfiles, guildID), nil
}

func (r *XPAdminRepository) DeleteTextXPProfile(ctx context.Context, guildID string, userID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	key := xpProfileKey(guildID, userID)
	if _, ok := r.TextProfiles[key]; !ok {
		return ports.ErrTextXPProfileMissing
	}
	delete(r.TextProfiles, key)
	return nil
}

func (r *XPAdminRepository) DeleteVoiceXPProfile(ctx context.Context, guildID string, userID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	key := xpProfileKey(guildID, userID)
	if _, ok := r.VoiceProfiles[key]; !ok {
		return ports.ErrVoiceXPProfileMissing
	}
	delete(r.VoiceProfiles, key)
	return nil
}

func (r *XPAdminRepository) DeleteTextXPGuild(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if deleteXPProfileGuild(r.TextProfiles, guildID) == 0 {
		return ports.ErrTextXPProfileMissing
	}
	return nil
}

func (r *XPAdminRepository) DeleteVoiceXPGuild(ctx context.Context, guildID string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	if deleteXPProfileGuild(r.VoiceProfiles, guildID) == 0 {
		return ports.ErrVoiceXPProfileMissing
	}
	return nil
}

func (r *XPAdminRepository) markVoiceXPSession(ctx context.Context, guildID string, userID string, state string) error {
	if err := r.ready(ctx); err != nil {
		return err
	}
	guildID = strings.TrimSpace(guildID)
	userID = strings.TrimSpace(userID)
	if guildID == "" || userID == "" {
		return domain.ErrInvalidXPAdjustment
	}
	key := xpProfileKey(guildID, userID)
	profile, ok := r.VoiceProfiles[key]
	if !ok {
		profile = domain.XPProfile{GuildID: guildID, UserID: userID}
	}
	profile = profile.Normalize()
	profile.GuildID = guildID
	profile.UserID = userID
	profile.LeaveJoin = state
	r.VoiceProfiles[key] = profile
	return nil
}

func (r *TextXPConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func (r *VoiceXPConfigRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func (r *TextXPRewardRoleRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func (r *VoiceXPRewardRoleRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func (r *XPAdminRepository) ready(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return r.Err
}

func filterRewardRoles(configs []domain.XPRewardRoleConfig, guildID string) []domain.XPRewardRoleConfig {
	guildID = strings.TrimSpace(guildID)
	out := make([]domain.XPRewardRoleConfig, 0, len(configs))
	for _, config := range configs {
		config = config.Normalize()
		if config.GuildID == guildID {
			out = append(out, config)
		}
	}
	return out
}

func deleteRewardRole(configs []domain.XPRewardRoleConfig, guildID string, level int64, roleID string) []domain.XPRewardRoleConfig {
	guildID = strings.TrimSpace(guildID)
	roleID = strings.TrimSpace(roleID)
	out := configs[:0]
	for _, config := range configs {
		config = config.Normalize()
		if config.GuildID == guildID && config.Level == level && config.RoleID == roleID {
			continue
		}
		out = append(out, config)
	}
	return out
}

func xpProfileKey(guildID string, userID string) string {
	return strings.TrimSpace(guildID) + "/" + strings.TrimSpace(userID)
}

func deleteXPProfileGuild(profiles map[string]domain.XPProfile, guildID string) int {
	guildID = strings.TrimSpace(guildID)
	deleted := 0
	for key, profile := range profiles {
		if strings.TrimSpace(profile.GuildID) != guildID && !strings.HasPrefix(key, guildID+"/") {
			continue
		}
		delete(profiles, key)
		deleted++
	}
	return deleted
}

func filterXPProfiles(profiles map[string]domain.XPProfile, guildID string) []domain.XPProfile {
	guildID = strings.TrimSpace(guildID)
	out := []domain.XPProfile{}
	for key, profile := range profiles {
		if strings.TrimSpace(profile.GuildID) != guildID && !strings.HasPrefix(key, guildID+"/") {
			continue
		}
		out = append(out, profile.Normalize())
	}
	return out
}

var _ ports.TextXPConfigRepository = (*TextXPConfigRepository)(nil)
var _ ports.TextXPConfigReader = (*TextXPConfigRepository)(nil)
var _ ports.VoiceXPConfigRepository = (*VoiceXPConfigRepository)(nil)
var _ ports.TextXPRewardRoleRepository = (*TextXPRewardRoleRepository)(nil)
var _ ports.VoiceXPRewardRoleRepository = (*VoiceXPRewardRoleRepository)(nil)
var _ ports.XPAdminRepository = (*XPAdminRepository)(nil)
var _ ports.TextXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.VoiceXPSessionRepository = (*XPAdminRepository)(nil)
var _ ports.VoiceXPAccrualRepository = (*XPAdminRepository)(nil)
var _ ports.XPResetRepository = (*XPAdminRepository)(nil)
var _ ports.XPRankRepository = (*XPAdminRepository)(nil)
