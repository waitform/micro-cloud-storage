package utils

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

// Profiler 性能分析器结构
type Profiler struct {
	cpuProfile *os.File
	memProfile *os.File
	startTime  time.Time
}

// NewProfiler 创建新的性能分析器
func NewProfiler() *Profiler {
	return &Profiler{
		startTime: time.Now(),
	}
}

// StartCPUProfile 开始CPU性能分析
func (p *Profiler) StartCPUProfile(filename string) error {
	var err error
	p.cpuProfile, err = os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建CPU性能分析文件: %v", err)
	}

	if err := pprof.StartCPUProfile(p.cpuProfile); err != nil {
		return fmt.Errorf("无法启动CPU性能分析: %v", err)
	}

	Info("开始CPU性能分析，结果将保存到: %s", filename)
	return nil
}

// StopCPUProfile 停止CPU性能分析
func (p *Profiler) StopCPUProfile() {
	if p.cpuProfile != nil {
		pprof.StopCPUProfile()
		p.cpuProfile.Close()
		Info("CPU性能分析已停止")
	}
}

// StartMemProfile 开始内存性能分析
func (p *Profiler) StartMemProfile(filename string) error {
	var err error
	p.memProfile, err = os.Create(filename)
	if err != nil {
		return fmt.Errorf("无法创建内存性能分析文件: %v", err)
	}

	Info("开始内存性能分析，结果将保存到: %s", filename)
	return nil
}

// StopMemProfile 停止内存性能分析并写入结果
func (p *Profiler) StopMemProfile() {
	if p.memProfile != nil {
		runtime.GC() // 执行垃圾回收以获取最新的内存状态
		if err := pprof.WriteHeapProfile(p.memProfile); err != nil {
			Error("写入内存性能分析文件失败: %v", err)
		}
		p.memProfile.Close()
		Info("内存性能分析已保存")
	}
}

// PrintMemStats 打印当前内存统计信息
func (p *Profiler) PrintMemStats() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	Info("内存统计信息:")
	Info("  分配的对象数: %d", m.Alloc)
	Info("  总分配量: %d", m.TotalAlloc)
	Info("  系统占用内存: %d", m.Sys)
	Info("  指针查找次数: %d", m.Lookups)
	Info("  内存分配次数: %d", m.Mallocs)
	Info("  内存释放次数: %d", m.Frees)
	Info("  当前Goroutine数量: %d", runtime.NumGoroutine())
}

// PrintDuration 打印性能分析持续时间
func (p *Profiler) PrintDuration() {
	duration := time.Since(p.startTime)
	Info("性能分析持续时间: %v", duration)
}
