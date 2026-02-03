// Package main provides development status API for Goclaw
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// DevStatusResponse contains development status information
type DevStatusResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// DevStatusData contains the development status information
type DevStatusData struct {
	CurrentModel        string      `json:"currentModel"`
	LastCommit         CommitInfo  `json:"lastCommit"`
	LastFileMod        FileModInfo `json:"lastFileMod"`
	TokenUsage         TokenUsage  `json:"tokenUsage"`
	ImplementedFeatures []string    `json:"implementedFeatures"`
	PlannedFeatures    []string    `json:"plannedFeatures"`
	ProjectStatus      string      `json:"projectStatus"`
	BuildTime          string      `json:"buildTime"`
}

// CommitInfo contains git commit information
type CommitInfo struct {
	Hash        string `json:"hash"`
	Message     string `json:"message"`
	Author      string `json:"author"`
	Date        string `json:"date"`
	TimeAgo     string `json:"timeAgo"`
	Branch      string `json:"branch"`
}

// FileModInfo contains file modification information
type FileModInfo struct {
	Filename    string `json:"filename"`
	ModifiedTime string `json:"modifiedTime"`
	TimeAgo     string `json:"timeAgo"`
	Path        string `json:"path"`
}

// TokenUsage contains token usage information
type TokenUsage struct {
	TotalTokens int     `json:"totalTokens"`
	EstimatedCost float64 `json:"estimatedCost"`
	LastUpdate  string  `json:"lastUpdate"`
}

// handleDevStatus provides development status information
func handleDevStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Gather development status information
		statusData := gatherDevStatus()

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(DevStatusResponse{
			Status: "ok",
			Data:   statusData,
		})
	}
}

// gatherDevStatus collects all development status information
func gatherDevStatus() DevStatusData {
	data := DevStatusData{}

	// Get current model information
	data.CurrentModel = getCurrentModel()

	// Get git commit information
	data.LastCommit = getGitCommitInfo()

	// Get last file modification
	data.LastFileMod = getLastFileModification()

	// Get token usage information
	data.TokenUsage = getTokenUsage()

	// Get implemented and planned features
	data.ImplementedFeatures, data.PlannedFeatures = getFeatures()

	// Get project status
	data.ProjectStatus = getProjectStatus()

	// Build time
	data.BuildTime = time.Now().Format("2006-01-02 15:04:05")

	return data
}

// getCurrentModel returns the current AI model being used
func getCurrentModel() string {
	// Read from goclaw_tasks.json to get model info
	tasksFile := filepath.Join(os.Getenv("HOME"), ".openclaw", "workspace", "goclaw_tasks.json")
	if _, err := os.Stat(tasksFile); err == nil {
		content, err := ioutil.ReadFile(tasksFile)
		if err == nil {
			var tasks map[string]interface{}
			if err := json.Unmarshal(content, &tasks); err == nil {
				// Return hardcoded GLM-4.7 model info
				return "zai/glm-4.7 (智谱AI GLM-4.7)"
			}
		}
	}
	return "zai/glm-4.7 (智谱AI GLM-4.7)"
}

// getGitCommitInfo returns the last git commit information
func getGitCommitInfo() CommitInfo {
	info := CommitInfo{
		Hash:    "N/A",
		Message: "N/A",
		Author:  "N/A",
		Date:    "N/A",
		Branch:  "main",
	}

	// Get current branch
	if branch, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output(); err == nil {
		info.Branch = string(branch)[:len(branch)-1] // Remove newline
	}

	// Get last commit hash
	if hash, err := exec.Command("git", "log", "-1", "--pretty=format:%H").Output(); err == nil {
		info.Hash = string(hash)[:7] // Show short hash
	}

	// Get last commit message
	if message, err := exec.Command("git", "log", "-1", "--pretty=format:%s").Output(); err == nil {
		info.Message = string(message)
	}

	// Get last commit author
	if author, err := exec.Command("git", "log", "-1", "--pretty=format:%an").Output(); err == nil {
		info.Author = string(author)
	}

	// Get last commit date
	if date, err := exec.Command("git", "log", "-1", "--pretty=format:%ci").Output(); err == nil {
		info.Date = string(date)[:19] // Remove timezone
		if commitTime, err := time.Parse("2006-01-02 15:04:05", info.Date); err == nil {
			info.TimeAgo = timeAgo(time.Since(commitTime))
		}
	}

	return info
}

// getLastFileModification returns the last modified file information
func getLastFileModification() FileModInfo {
	info := FileModInfo{
		Filename:     "N/A",
		ModifiedTime: "N/A",
		Path:         "N/A",
	}

	// Find the most recently modified Go file
	var lastMod time.Time
	var lastFile string

	err := filepath.Walk("/home/daniel/projects/goclaw/", func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Only consider .go files
		if filepath.Ext(path) == ".go" && fi.ModTime().After(lastMod) {
			lastMod = fi.ModTime()
			lastFile = path
		}
		return nil
	})

	if err == nil && lastFile != "" {
		info.Path = lastFile
		info.Filename = filepath.Base(lastFile)
		info.ModifiedTime = lastMod.Format("2006-01-02 15:04:05")
		info.TimeAgo = timeAgo(time.Since(lastMod))
	}

	return info
}

// getTokenUsage returns token usage information
func getTokenUsage() TokenUsage {
	usage := TokenUsage{
		TotalTokens:   0,
		EstimatedCost: 0.0,
		LastUpdate:   time.Now().Format("2006-01-02 15:04:05"),
	}

	// Read from memory files to estimate token usage
	memoryDir := filepath.Join(os.Getenv("HOME"), ".openclaw", "workspace", "memory")
	if _, err := os.Stat(memoryDir); err == nil {
		files, err := ioutil.ReadDir(memoryDir)
		if err == nil {
			totalSize := 0
			for _, file := range files {
				if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
					totalSize += int(file.Size())
				}
			}
			// Rough estimation: ~4 characters per token
			usage.TotalTokens = totalSize / 4
		}
	}

	return usage
}

// getFeatures returns implemented and planned features
func getFeatures() ([]string, []string) {
	implemented := []string{
		"✅ 核心模块开发（向量存储、定时任务、API服务、配置管理）",
		"✅ 测试系统（单元测试、集成测试、测试覆盖率）",
		"✅ 功能对比分析（OpenClaw vs Goclaw）",
		"✅ AI模型集成（Minimax、通义千问、智谱AI）",
		"✅ Web界面和PWA支持",
		"✅ 记忆系统（短期、长期、工作记忆）",
		"✅ 定时任务系统",
		"✅ REST API接口",
		"✅ 开发状态监控",
	}

	planned := []string{
		"⏳ 会话管理系统",
		"⏳ 安全模型完善",
		"⏳ 工具系统基础框架",
		"⏳ 技能系统",
		"⏳ 多渠道消息系统",
		"⏳ 媒体处理能力",
		"⏳ 节点系统",
		"⏳ 浏览器控制",
		"⏳ Docker容器化",
		"⏳ CI/CD流水线",
	}

	return implemented, planned
}

// getProjectStatus returns the overall project status
func getProjectStatus() string {
	// Read from goclaw_tasks.json
	tasksFile := filepath.Join(os.Getenv("HOME"), ".openclaw", "workspace", "goclaw_tasks.json")
	if _, err := os.Stat(tasksFile); err == nil {
		content, err := ioutil.ReadFile(tasksFile)
		if err == nil {
			var tasks map[string]interface{}
			if err := json.Unmarshal(content, &tasks); err == nil {
				// Calculate completion percentage
				completedCount := 0
				totalCount := 0
				
				if tasksArray, ok := tasks["tasks"].([]interface{}); ok {
					for _, task := range tasksArray {
						if taskMap, ok := task.(map[string]interface{}); ok {
							if completed, ok := taskMap["completed"].(bool); ok {
								totalCount++
								if completed {
									completedCount++
								}
							}
						}
					}
				}
				
				if totalCount > 0 {
					percentage := float64(completedCount) / float64(totalCount) * 100
					return fmt.Sprintf("🚀 开发中 - 完成度: %.1f%% (%d/%d 任务)", percentage, completedCount, totalCount)
				}
			}
		}
	}
	
	return "🚀 开发中"
}

// timeAgo returns a human-readable time difference
func timeAgo(duration time.Duration) string {
	seconds := int(duration.Seconds())
	
	if seconds < 60 {
		return fmt.Sprintf("%d 秒前", seconds)
	}
	
	minutes := seconds / 60
	if minutes < 60 {
		return fmt.Sprintf("%d 分钟前", minutes)
	}
	
	hours := minutes / 60
	if hours < 24 {
		return fmt.Sprintf("%d 小时前", hours)
	}
	
	days := hours / 24
	if days < 30 {
		return fmt.Sprintf("%d 天前", days)
	}
	
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%d 月前", months)
	}
	
	years := months / 12
	return fmt.Sprintf("%d 年前", years)
}
