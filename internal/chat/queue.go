package chat

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// MessageQueue 消息队列结构
type MessageQueue struct {
	queue    chan QueuedMessage
	workers  int
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
	handlers map[string]MessageHandler
	mutex    sync.RWMutex
}

// QueuedMessage 队列中的消息结构
type QueuedMessage struct {
	ID        string
	SessionID string
	UserID    string
	Message   string
	Timestamp time.Time
	ReplyChan chan MessageResponse
	Context   map[string]interface{}
}

// MessageResponse 消息响应结构
type MessageResponse struct {
	ID      string
	Content string
	Error   error
	Data    interface{}
}

// MessageHandler 消息处理器接口
type MessageHandler func(context.Context, QueuedMessage) MessageResponse

// NewMessageQueue 创建新的消息队列
func NewMessageQueue(queueSize, workers int) *MessageQueue {
	ctx, cancel := context.WithCancel(context.Background())
	
	mq := &MessageQueue{
		queue:    make(chan QueuedMessage, queueSize),
		workers:  workers,
		ctx:      ctx,
		cancel:   cancel,
		handlers: make(map[string]MessageHandler),
	}
	
	// 启动工作者协程
	mq.startWorkers()
	
	return mq
}

// AddHandler 添加消息处理器
func (mq *MessageQueue) AddHandler(name string, handler MessageHandler) {
	mq.mutex.Lock()
	defer mq.mutex.Unlock()
	mq.handlers[name] = handler
}

// Enqueue 添加消息到队列
func (mq *MessageQueue) Enqueue(msg QueuedMessage) error {
	select {
	case <-mq.ctx.Done():
		return context.Canceled
	case mq.queue <- msg:
		return nil
	}
}

// ProcessWithHandler 使用指定处理器处理消息
func (mq *MessageQueue) ProcessWithHandler(handlerName string, msg QueuedMessage) MessageResponse {
	mq.mutex.RLock()
	handler, exists := mq.handlers[handlerName]
	mq.mutex.RUnlock()
	
	if !exists {
		return MessageResponse{
			ID:    msg.ID,
			Error: fmt.Errorf("handler %s not found", handlerName),
		}
	}
	
	return handler(mq.ctx, msg)
}

// startWorkers 启动工作者协程
func (mq *MessageQueue) startWorkers() {
	for i := 0; i < mq.workers; i++ {
		mq.wg.Add(1)
		go mq.worker(i)
	}
}

// worker 工作者协程
func (mq *MessageQueue) worker(workerID int) {
	defer mq.wg.Done()
	
	for {
		select {
		case <-mq.ctx.Done():
			return
		case msg := <-mq.queue:
			response := mq.ProcessWithHandler("default", msg)
			
			// 发送响应
			if msg.ReplyChan != nil {
				select {
				case msg.ReplyChan <- response:
				case <-time.After(5 * time.Second):
					// 超时处理
				}
			}
		}
	}
}

// Shutdown 关闭队列
func (mq *MessageQueue) Shutdown() {
	mq.cancel()
	close(mq.queue)
	mq.wg.Wait()
}

// GetQueueStats 获取队列统计信息
func (mq *MessageQueue) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"queue_length": len(mq.queue),
		"workers":      mq.workers,
		"capacity":     cap(mq.queue),
	}
}