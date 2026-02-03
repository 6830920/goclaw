package identity

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"goclaw/internal/config"
)

// Identity 表示助手的身份信息
type Identity struct {
	Name     string            `json:"name"`     // 助手名称
	Creature string            `json:"creature"` // 助手类型/生物
	Vibe     string            `json:"vibe"`     // 助手风格
	Emoji    string            `json:"emoji"`    // 助手表情
	Notes    []string          `json:"notes"`    // 备注
	Config   map[string]string `json:"config"`   // 额外配置
}

// IdentityManager 管理身份信息
type IdentityManager struct {
	workspace string
	identity  *Identity
}

// NewIdentityManager 创建身份管理器
func NewIdentityManager(workspace string) *IdentityManager {
	return &IdentityManager{
		workspace: workspace,
	}
}

// LoadIdentityFromFiles 从文件加载身份信息
func (im *IdentityManager) LoadIdentityFromFiles() error {
	// 尝试加载IDENTITY.md
	identityPath := filepath.Join(im.workspace, "IDENTITY.md")
	identity, err := im.loadIdentityFromFile(identityPath)
	if err == nil {
		im.identity = identity
		return nil
	}

	// 尝试加载SOUL.md
	soulPath := filepath.Join(im.workspace, "SOUL.md")
	soul, err := im.loadSoulFromFile(soulPath)
	if err == nil {
		im.identity = soul
		return nil
	}

	// 如果都没有，创建默认身份
	im.identity = &Identity{
		Name:     "Goclaw Assistant",
		Creature: "AI Assistant",
		Vibe:     "Helpful and efficient",
		Emoji:    "🤖",
		Notes:    []string{"Default identity for Goclaw"},
		Config:   make(map[string]string),
	}

	return nil
}

// loadIdentityFromFile 从IDENTITY.md文件加载身份
func (im *IdentityManager) loadIdentityFromFile(path string) (*Identity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	identity := &Identity{
		Config: make(map[string]string),
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- **Name:**") {
			name := strings.TrimPrefix(line, "- **Name:**")
			name = strings.TrimSpace(name)
			// 处理带括号的名称，如 "捣蛋 (Dǎo Dàn)"
			if strings.Contains(name, "(") && strings.Contains(name, ")") {
				parts := strings.Split(name, "(")
				if len(parts) > 0 {
					name = strings.TrimSpace(parts[0])
					// 移除末尾的空格和右括号内容
					name = strings.ReplaceAll(name, ")", "")
				}
			}
			identity.Name = strings.TrimSpace(name)
		} else if strings.HasPrefix(line, "- **Creature:**") {
			creature := strings.TrimPrefix(line, "- **Creature:**")
			identity.Creature = strings.TrimSpace(creature)
		} else if strings.HasPrefix(line, "- **Vibe:**") {
			vibe := strings.TrimPrefix(line, "- **Vibe:**")
			identity.Vibe = strings.TrimSpace(vibe)
		} else if strings.HasPrefix(line, "- **Emoji:**") {
			emoji := strings.TrimPrefix(line, "- **Emoji:**")
			identity.Emoji = strings.TrimSpace(emoji)
		} else if strings.HasPrefix(line, "- ") && !strings.Contains(line, "**") {
			note := strings.TrimPrefix(line, "- ")
			if note != "" {
				identity.Notes = append(identity.Notes, strings.TrimSpace(note))
			}
		}
	}

	return identity, nil
}

// loadSoulFromFile 从SOUL.md文件加载身份信息
func (im *IdentityManager) loadSoulFromFile(path string) (*Identity, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// SOUL.md通常包含更详细的个性描述，这里简化处理
	identity := &Identity{
		Name:     "Goclaw Assistant",
		Creature: "Digital Being",
		Vibe:     "Authentic and capable",
		Emoji:    "💡",
		Notes:    []string{"Powered by GoClaw framework"},
		Config:   make(map[string]string),
	}

	// 提取关键段落
	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, "**Be genuinely helpful") {
			// 提取核心理念
			if i+1 < len(lines) {
				nextLine := lines[i+1]
				if strings.Contains(nextLine, "Skip the") {
					identity.Vibe = "Genuinely helpful, direct approach"
				}
			}
		}
	}

	return identity, nil
}

// GetIdentity 获取身份信息
func (im *IdentityManager) GetIdentity() *Identity {
	if im.identity == nil {
		_ = im.LoadIdentityFromFiles()
	}
	return im.identity
}

// GetIdentityDescription 获取身份描述
func (im *IdentityManager) GetIdentityDescription() string {
	identity := im.GetIdentity()
	if identity == nil {
		return "No identity configured"
	}

	desc := fmt.Sprintf("%s %s - %s", identity.Emoji, identity.Name, identity.Vibe)
	if identity.Creature != "" {
		desc += fmt.Sprintf(" (%s)", identity.Creature)
	}

	return desc
}

// ApplyToConfig 将身份信息应用到配置
func (im *IdentityManager) ApplyToConfig(cfg *config.Config) {
	identity := im.GetIdentity()
	if identity != nil {
		if cfg.Identity == nil {
			cfg.Identity = make(map[string]string)
		}
		cfg.Identity["name"] = identity.Name
		cfg.Identity["vibe"] = identity.Vibe
		cfg.Identity["creature"] = identity.Creature
		cfg.Identity["emoji"] = identity.Emoji
	}
}