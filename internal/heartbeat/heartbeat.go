package heartbeat

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"goclaw/internal/config"
	"goclaw/pkg/ai"
)

const (
	DefaultHeartbeatPrompt = "Read HEARTBEAT.md if it exists (workspace context). Follow it strictly. Do not infer or repeat old tasks from prior chats. If nothing needs attention, reply HEARTBEAT_OK."
	DefaultHeartbeatEvery  = 30 * time.Minute
)

// HeartbeatManager 管理心跳功能
type HeartbeatManager struct {
	cfg         *config.Config
	aiClient    ai.Client
	workspace   string
	interval    time.Duration
	stopChan    chan struct{}
	stoppedChan chan struct{}
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(cfg *config.Config, aiClient ai.Client, workspace string) *HeartbeatManager {
	interval := DefaultHeartbeatEvery
	if cfg.Heartbeat.Interval != "" {
		if dur, err := time.ParseDuration(cfg.Heartbeat.Interval); err == nil {
			interval = dur
		}
	}

	return &HeartbeatManager{
		cfg:         cfg,
		aiClient:    aiClient,
		workspace:   workspace,
		interval:    interval,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
	}
}

// IsHeartbeatContentEffectivelyEmpty 检查HEARTBEAT.md内容是否"有效为空"
func IsHeartbeatContentEffectivelyEmpty(content string) bool {
	if content == "" {
		return false
	}

	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		
		// 跳过空行
		if line == "" {
			continue
		}
		
		// 跳过markdown标题行 (# 后跟空格或行尾)
		if matched, _ := regexp.MatchString(`^#+(\s|$)`, line); matched {
			continue
		}
		
		// 跳过空的markdown列表项
		if matched, _ := regexp.MatchString(`^[-*+]\s*(\[[\sXx]?\]\s*)?$`, line); matched {
			continue
		}
		
		// 找到非空、非注释行 - 有可执行内容
		return false
	}
	
	// 所有行都是空行或注释
	return true
}

// RunOnce 执行一次心跳
func (hm *HeartbeatManager) RunOnce(ctx context.Context) error {
	heartbeatFile := filepath.Join(hm.workspace, "HEARTBEAT.md")
	
	// 检查HEARTBEAT.md是否存在且有效
	content, err := os.ReadFile(heartbeatFile)
	if err != nil {
		// 文件不存在，使用默认行为
		return hm.sendHeartbeatOK()
	}
	
	contentStr := string(content)
	if IsHeartbeatContentEffectivelyEmpty(contentStr) {
		// 文件存在但无效内容，发送HEARTBEAT_OK
		return hm.sendHeartbeatOK()
	}
	
	// 有有效内容，交给AI处理
	prompt := hm.cfg.Heartbeat.Prompt
	if prompt == "" {
		prompt = DefaultHeartbeatPrompt
	}
	
	// 构建心跳消息
	heartbeatMsg := fmt.Sprintf("%s\n\nHEARTBEAT.md content:\n%s", prompt, contentStr)
	
	if hm.aiClient != nil {
		// TODO: 实际调用AI处理心跳
		// resp, err := hm.aiClient.SendMessage(ctx, "user", heartbeatMsg)
		// if err != nil {
		//     return err
		// }
		// 暂时模拟AI响应
		fmt.Printf("Heartbeat processed: %s\n", heartbeatMsg)
	} else {
		// 没有AI客户端，直接发送HEARTBEAT_OK
		return hm.sendHeartbeatOK()
	}
	
	return nil
}

// sendHeartbeatOK 发送心跳确认
func (hm *HeartbeatManager) sendHeartbeatOK() error {
	fmt.Println("HEARTBEAT_OK")
	return nil
}

// Start 启动心跳循环
func (hm *HeartbeatManager) Start(ctx context.Context) {
	ticker := time.NewTicker(hm.interval)
	defer ticker.Stop()
	defer close(hm.stoppedChan)

	for {
		select {
		case <-ticker.C:
			if err := hm.RunOnce(ctx); err != nil {
				fmt.Printf("Heartbeat error: %v\n", err)
			}
		case <-hm.stopChan:
			fmt.Println("Heartbeat manager stopped")
			return
		case <-ctx.Done():
			fmt.Println("Heartbeat manager context cancelled")
			return
		}
	}
}

// Stop 停止心跳循环
func (hm *HeartbeatManager) Stop() {
	close(hm.stopChan)
	<-hm.stoppedChan
}

// CheckAndRun 检查并运行心跳（手动触发）
func (hm *HeartbeatManager) CheckAndRun(ctx context.Context) error {
	heartbeatFile := filepath.Join(hm.workspace, "HEARTBEAT.md")
	
	// 检查文件是否存在
	if _, err := os.Stat(heartbeatFile); os.IsNotExist(err) {
		// 文件不存在，发送HEARTBEAT_OK
		return hm.sendHeartbeatOK()
	}
	
	return hm.RunOnce(ctx)
}