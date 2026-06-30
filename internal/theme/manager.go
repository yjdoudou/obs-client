package theme

import (
	"encoding/base64"
	"encoding/json"
	"obs-client/internal/db"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type ThemeConfig struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Preview      string                 `json:"preview"`
	GlassEnabled bool                   `json:"glassEnabled"`
	Variables    map[string]interface{} `json:"variables"`
}

type UserTheme struct {
	ID             string  `json:"id" gorm:"primaryKey"`
	ThemeID        string  `json:"themeId"`
	IsDark         bool    `json:"isDark"`
	BackgroundPath string  `json:"backgroundPath"`
	OverlayOpacity float64 `json:"overlayOpacity"`
}

type BackgroundInfo struct {
	ImageBase64 string  `json:"imageBase64"`
	Opacity     float64 `json:"opacity"`
	HasImage    bool    `json:"hasImage"`
}

var themes = []ThemeConfig{
	{
		ID:           "glass-light",
		Name:         "玻璃亮色",
		Description:  "透明玻璃效果，现代感十足",
		Preview:      "#f5f7fa",
		GlassEnabled: true,
		Variables: map[string]interface{}{
			"primary-color": "#0066ff",
			"bg-app":        "rgba(245, 247, 250, 0.6)",
			"bg-card":       "rgba(255, 255, 255, 0.8)",
			"text-main":     "#333333",
			"text-light":    "#666666",
			"border-color":  "rgba(0, 0, 0, 0.08)",
			"glass-bg":      "rgba(255, 255, 255, 0.6)",
		},
	},
	{
		ID:           "glass-dark",
		Name:         "玻璃暗色",
		Description:  "深色透明玻璃效果，护眼模式",
		Preview:      "#121212",
		GlassEnabled: true,
		Variables: map[string]interface{}{
			"primary-color": "#0066ff",
			"bg-app":        "rgba(18, 18, 18, 0.6)",
			"bg-card":       "rgba(30, 30, 30, 0.8)",
			"text-main":     "#e0e0e0",
			"text-light":    "#888888",
			"border-color":  "rgba(255, 255, 255, 0.08)",
			"glass-bg":      "rgba(30, 30, 30, 0.6)",
		},
	},
	{
		ID:           "solid-light",
		Name:         "经典亮色",
		Description:  "传统不透明风格，简洁清爽",
		Preview:      "#ffffff",
		GlassEnabled: false,
		Variables: map[string]interface{}{
			"primary-color": "#0066ff",
			"bg-app":        "#ffffff",
			"bg-card":       "#ffffff",
			"text-main":     "#333333",
			"text-light":    "#666666",
			"border-color":  "#e8e8e8",
			"glass-bg":      "#ffffff",
		},
	},
	{
		ID:           "solid-dark",
		Name:         "经典暗色",
		Description:  "深色不透明风格，沉稳专业",
		Preview:      "#1a1a1a",
		GlassEnabled: false,
		Variables: map[string]interface{}{
			"primary-color": "#0066ff",
			"bg-app":        "#1a1a1a",
			"bg-card":       "#242424",
			"text-main":     "#e0e0e0",
			"text-light":    "#888888",
			"border-color":  "#333333",
			"glass-bg":      "#242424",
		},
	},
	{
		ID:           "minimal-light",
		Name:         "极简亮色",
		Description:  "极简风格，专注内容",
		Preview:      "#fafafa",
		GlassEnabled: false,
		Variables: map[string]interface{}{
			"primary-color": "#0070f3",
			"bg-app":        "#fafafa",
			"bg-card":       "#ffffff",
			"text-main":     "#213547",
			"text-light":    "#6b7280",
			"border-color":  "#e5e7eb",
			"glass-bg":      "#ffffff",
		},
	},
	{
		ID:           "ocean-blue",
		Name:         "海洋蓝",
		Description:  "清新蓝色主题，舒适视觉",
		Preview:      "#f0f9ff",
		GlassEnabled: true,
		Variables: map[string]interface{}{
			"primary-color": "#0ea5e9",
			"bg-app":        "rgba(240, 249, 255, 0.7)",
			"bg-card":       "rgba(255, 255, 255, 0.9)",
			"text-main":     "#0c4a6e",
			"text-light":    "#075985",
			"border-color":  "rgba(14, 165, 233, 0.1)",
			"glass-bg":      "rgba(255, 255, 255, 0.7)",
		},
	},
}

type Manager struct {
	db *gorm.DB
}

func NewManager(db *gorm.DB) *Manager {
	return &Manager{db: db}
}

func (m *Manager) GetAllThemes() []ThemeConfig {
	return themes
}

func (m *Manager) GetThemeByID(id string) (ThemeConfig, bool) {
	for _, t := range themes {
		if t.ID == id {
			return t, true
		}
	}
	return ThemeConfig{}, false
}

func (m *Manager) GetUserTheme(userID string) (string, bool) {
	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		return "glass-light", true
	}
	return userTheme.ThemeID, userTheme.IsDark
}

func (m *Manager) SaveUserTheme(userID string, themeID string, isDark bool) error {
	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		userTheme = UserTheme{
			ID:      userID,
			ThemeID: themeID,
			IsDark:  isDark,
		}
		return m.db.Create(&userTheme).Error
	}
	userTheme.ThemeID = themeID
	userTheme.IsDark = isDark
	return m.db.Save(&userTheme).Error
}

func (m *Manager) GetCurrentThemeJSON(userID string) (string, error) {
	themeID, _ := m.GetUserTheme(userID)
	theme, ok := m.GetThemeByID(themeID)
	if !ok {
		theme = themes[0]
	}
	data, err := json.Marshal(theme)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getBackgroundDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".obs-client", "backgrounds")
	return dir, os.MkdirAll(dir, 0755)
}

func (m *Manager) SaveBackgroundImage(userID string, base64Image string) error {
	dir, err := getBackgroundDir()
	if err != nil {
		return err
	}

	imageData, err := base64.StdEncoding.DecodeString(base64Image)
	if err != nil {
		return err
	}

	path := filepath.Join(dir, userID+".png")
	err = os.WriteFile(path, imageData, 0644)
	if err != nil {
		return err
	}

	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		userTheme = UserTheme{ID: userID, ThemeID: "glass-light", OverlayOpacity: 0.5}
	}
	userTheme.BackgroundPath = path

	return m.db.Save(&userTheme).Error
}

func (m *Manager) GetBackgroundInfo(userID string) (BackgroundInfo, error) {
	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		return BackgroundInfo{HasImage: false, Opacity: 0.5}, nil
	}

	if userTheme.BackgroundPath == "" {
		return BackgroundInfo{HasImage: false, Opacity: userTheme.OverlayOpacity}, nil
	}

	data, err := os.ReadFile(userTheme.BackgroundPath)
	if err != nil {
		return BackgroundInfo{HasImage: false, Opacity: userTheme.OverlayOpacity}, nil
	}

	base64Str := base64.StdEncoding.EncodeToString(data)
	return BackgroundInfo{
		ImageBase64: "data:image/png;base64," + base64Str,
		Opacity:     userTheme.OverlayOpacity,
		HasImage:    true,
	}, nil
}

func (m *Manager) SaveOverlayOpacity(userID string, opacity float64) error {
	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		userTheme = UserTheme{ID: userID, ThemeID: "glass-light", OverlayOpacity: opacity}
		return m.db.Create(&userTheme).Error
	}
	userTheme.OverlayOpacity = opacity
	return m.db.Save(&userTheme).Error
}

func (m *Manager) ClearBackgroundImage(userID string) error {
	var userTheme UserTheme
	result := m.db.Where("id = ?", userID).First(&userTheme)
	if result.Error != nil {
		return nil
	}

	if userTheme.BackgroundPath != "" {
		os.Remove(userTheme.BackgroundPath)
	}

	userTheme.BackgroundPath = ""
	return m.db.Save(&userTheme).Error
}

func init() {
	db.RegisterModel(&UserTheme{})
}
