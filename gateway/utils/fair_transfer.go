package utils

import (
	"context"
	"io"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// FairTransferManager 公平传输管理器
// 用于在多个并发传输之间公平分配带宽
type FairTransferManager struct {
	// 全局带宽限制 (bytes/s)
	globalRateLimit int

	// 每个用户的带宽限制 (bytes/s)
	perUserRateLimit int

	// 每个传输连接的最小带宽保证 (bytes/s)
	minTransferRate int

	// 限速器
	globalLimiter *rate.Limiter

	// 用户级别的限速器
	userLimiters map[uint]*rate.Limiter

	// 传输连接级别的限速器
	transferLimiters map[string]*rate.Limiter

	// 锁
	mutex sync.RWMutex

	// 活跃传输连接数
	activeTransfers map[string]uint

	// 活跃用户数
	activeUsers map[uint]int

	activeCounter int
}

// Manager 全局公平传输管理器实例
// 提供用户级公平带宽分配功能
var Manager = NewFairTransferManager(10*1024*1024, 5*1024*1024, 128*1024)

// NewFairTransferManager 创建一个新的公平传输管理器
func NewFairTransferManager(globalRateLimit, perUserRateLimit, minTransferRate int) *FairTransferManager {
	manager := &FairTransferManager{
		globalRateLimit:  globalRateLimit,
		perUserRateLimit: perUserRateLimit,
		minTransferRate:  minTransferRate,
		globalLimiter:    rate.NewLimiter(rate.Limit(globalRateLimit), globalRateLimit),
		userLimiters:     make(map[uint]*rate.Limiter),
		transferLimiters: make(map[string]*rate.Limiter),
		activeTransfers:  make(map[string]uint),
		activeUsers:      make(map[uint]int),
		activeCounter:    0,
	}

	return manager
}

// GetTransferLimiter 获取传输连接的限速器
func (m *FairTransferManager) GetTransferLimiter(transferID string) *rate.Limiter {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	limiter, exists := m.transferLimiters[transferID]
	if !exists {
		return nil
	}

	return limiter
}

// GetUserLimiter 获取用户的限速器
func (m *FairTransferManager) GetUserLimiter(userID uint) *rate.Limiter {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	limiter, exists := m.userLimiters[userID]
	if !exists {
		limiter = rate.NewLimiter(rate.Limit(m.perUserRateLimit), 5*1024*1024)
		m.userLimiters[userID] = limiter
	}
	m.activeUsers[userID]++
	return limiter
}

// GetGlobalLimiter 获取全局限速器
func (m *FairTransferManager) GetGlobalLimiter() *rate.Limiter {
	return m.globalLimiter
}

// FairRateLimitedWriter 公平速率限制写入器
type FairRateLimitedWriter struct {
	writer   io.Writer
	limiters []*rate.Limiter // 多级限速器：全局、用户、传输连接
	ctx      context.Context
}

// NewFairRateLimitedWriter 创建一个新的公平速率限制写入器
func NewFairRateLimitedWriter(writer io.Writer, limiters ...*rate.Limiter) *FairRateLimitedWriter {
	return &FairRateLimitedWriter{
		writer:   writer,
		limiters: limiters,
		ctx:      context.Background(),
	}
}

// Write 实现io.Writer接口
func (w *FairRateLimitedWriter) Write(p []byte) (int, error) {
	// 等待所有限速器都有足够的令牌
	for _, limiter := range w.limiters {
		if limiter != nil {
			if err := limiter.WaitN(w.ctx, len(p)); err != nil {
				return 0, err
			}
		}
	}

	// 写入数据
	return w.writer.Write(p)
}

type FairRateLimitedReader struct {
	reader   io.Reader
	limiters []*rate.Limiter
	ctx      context.Context
}

func NewFairRateLimitedReader(reader io.Reader, limiters ...*rate.Limiter) *FairRateLimitedReader {
	return &FairRateLimitedReader{
		reader:   reader,
		limiters: limiters,
		ctx:      context.Background(),
	}
}

// Read 实现 io.Reader 接口
func (r *FairRateLimitedReader) Read(p []byte) (int, error) {
	// 等待所有限速器都有足够的令牌
	for _, limiter := range r.limiters {
		if limiter != nil {
			if err := limiter.WaitN(r.ctx, len(p)); err != nil {
				return 0, err
			}
		}
	}

	// 读取数据
	return r.reader.Read(p)
}

// 自动更新速率
func (m *FairTransferManager) UpdateRate() {
	preRateLimit := m.globalRateLimit / m.activeCounter
	m.perUserRateLimit = preRateLimit
	for _, limiter := range m.transferLimiters {
		if limiter != nil {
			limiter.SetLimit(rate.Limit(preRateLimit))
		}
	}
}
func (m *FairTransferManager) UnregisterTransfer(userID uint) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.activeUsers[userID]--
	if m.activeUsers[userID] <= 0 {
		delete(m.activeUsers, userID)
		m.activeCounter--
		m.UpdateRate()
	}
}

// DynamicRateLimiter 动态速率限制器
// 可以根据实时情况动态调整速率限制
type DynamicRateLimiter struct {
	limiter            *rate.Limiter
	targetRate         int
	currentRate        int
	lastAdjust         time.Time
	adjustmentInterval time.Duration
	mutex              sync.RWMutex
}

// NewDynamicRateLimiter 创建一个新的动态速率限制器
func NewDynamicRateLimiter(initialRate int) *DynamicRateLimiter {
	return &DynamicRateLimiter{
		limiter:            rate.NewLimiter(rate.Limit(initialRate), initialRate),
		targetRate:         initialRate,
		currentRate:        initialRate,
		adjustmentInterval: time.Second * 5, // 每5秒调整一次
	}
}

// WaitN 等待N个令牌
func (d *DynamicRateLimiter) WaitN(ctx context.Context, n int) error {
	d.mutex.RLock()
	limiter := d.limiter
	d.mutex.RUnlock()

	return limiter.WaitN(ctx, n)
}

// SetTargetRate 设置目标速率
func (d *DynamicRateLimiter) SetTargetRate(rate int) {
	d.mutex.Lock()
	d.targetRate = rate
	d.mutex.Unlock()
}

// AdjustRate 动态调整速率
func (d *DynamicRateLimiter) AdjustRate() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	now := time.Now()
	if now.Sub(d.lastAdjust) < d.adjustmentInterval {
		return
	}

	d.lastAdjust = now

	// 如果当前速率与目标速率不同，则进行调整
	if d.currentRate != d.targetRate {
		d.currentRate = d.targetRate
		d.limiter.SetLimit(rate.Limit(d.currentRate))
		d.limiter.SetBurst(d.currentRate)
	}
}

// GetCurrentRate 获取当前速率
func (d *DynamicRateLimiter) GetCurrentRate() int {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.currentRate
}
