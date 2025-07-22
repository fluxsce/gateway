package pprof

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"gateway/pkg/logger"
)

// Analyzer pprof分析器
type Analyzer struct {
	config *Config
}

// NewAnalyzer 创建分析器
func NewAnalyzer(config *Config) *Analyzer {
	return &Analyzer{
		config: config,
	}
}

// RunAnalysis 运行分析
func (a *Analyzer) RunAnalysis() error {
	if !a.config.AutoAnalysis.Enabled {
		return fmt.Errorf("自动分析未启用")
	}

	// 创建时间戳目录
	timestamp := time.Now().Format("20060102_150405")
	outputDir := filepath.Join(a.config.AutoAnalysis.OutputDir, timestamp)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	logger.Info("开始性能分析", "output_dir", outputDir)

	// 收集各种profile数据
	if err := a.collectProfiles(outputDir); err != nil {
		logger.Error("收集profile数据失败", "error", err)
		return err
	}

	// 生成分析报告
	if err := a.generateReports(outputDir); err != nil {
		logger.Error("生成分析报告失败", "error", err)
		return err
	}

	// 清理旧数据
	if a.config.AutoAnalysis.SaveHistory {
		if err := a.cleanupOldData(); err != nil {
			logger.Warn("清理旧数据失败", "error", err)
		}
	}

	logger.Info("性能分析完成", "output_dir", outputDir)
	return nil
}

// collectProfiles 收集profile数据
func (a *Analyzer) collectProfiles(outputDir string) error {
	baseURL := fmt.Sprintf("http://localhost%s/debug/pprof", a.config.Listen)

	profiles := map[string]string{
		"cpu":          fmt.Sprintf("%s/profile?seconds=%d", baseURL, int(a.config.AutoAnalysis.CPUSampleDuration.Seconds())),
		"heap":         fmt.Sprintf("%s/heap", baseURL),
		"goroutine":    fmt.Sprintf("%s/goroutine", baseURL),
		"allocs":       fmt.Sprintf("%s/allocs", baseURL),
		"block":        fmt.Sprintf("%s/block", baseURL),
		"mutex":        fmt.Sprintf("%s/mutex", baseURL),
		"threadcreate": fmt.Sprintf("%s/threadcreate", baseURL),
	}

	for profileType, url := range profiles {
		filename := filepath.Join(outputDir, fmt.Sprintf("%s.prof", profileType))

		if err := a.downloadProfile(url, filename); err != nil {
			logger.Error("下载profile失败", "type", profileType, "error", err)
			continue
		}

		logger.Debug("收集profile成功", "type", profileType, "file", filename)
	}

	return nil
}

// downloadProfile 下载profile数据
func (a *Analyzer) downloadProfile(url, filename string) error {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP状态码: %d", resp.StatusCode)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// generateReports 生成分析报告
func (a *Analyzer) generateReports(outputDir string) error {
	// 检查go命令是否可用
	if !a.isGoAvailable() {
		logger.Warn("未找到go命令，跳过报告生成")
		return nil
	}

	reports := map[string]string{
		"cpu":       "cpu.prof",
		"heap":      "heap.prof",
		"goroutine": "goroutine.prof",
		"allocs":    "allocs.prof",
		"block":     "block.prof",
		"mutex":     "mutex.prof",
	}

	for reportType, profileFile := range reports {
		profilePath := filepath.Join(outputDir, profileFile)

		// 检查profile文件是否存在
		if _, err := os.Stat(profilePath); os.IsNotExist(err) {
			continue
		}

		// 生成top报告
		if err := a.generateTopReport(profilePath, outputDir, reportType); err != nil {
			logger.Error("生成top报告失败", "type", reportType, "error", err)
		}

		// 生成详细报告
		if err := a.generateDetailedReport(profilePath, outputDir, reportType); err != nil {
			logger.Error("生成详细报告失败", "type", reportType, "error", err)
		}
	}

	// 生成系统信息报告
	if err := a.generateSystemReport(outputDir); err != nil {
		logger.Error("生成系统信息报告失败", "error", err)
	}

	return nil
}

// generateTopReport 生成top报告
func (a *Analyzer) generateTopReport(profilePath, outputDir, reportType string) error {
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_top.txt", reportType))

	cmd := exec.Command("go", "tool", "pprof", "-top", profilePath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, output, 0644)
}

// generateDetailedReport 生成详细报告
func (a *Analyzer) generateDetailedReport(profilePath, outputDir, reportType string) error {
	outputFile := filepath.Join(outputDir, fmt.Sprintf("%s_list.txt", reportType))

	cmd := exec.Command("go", "tool", "pprof", "-list", ".*", profilePath)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	return os.WriteFile(outputFile, output, 0644)
}

// generateSystemReport 生成系统信息报告
func (a *Analyzer) generateSystemReport(outputDir string) error {
	outputFile := filepath.Join(outputDir, "system_info.txt")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	info := fmt.Sprintf(`系统信息报告
生成时间: %s
Go版本: %s
操作系统: %s
架构: %s
CPU核心数: %d
协程数量: %d

内存统计:
  分配的内存: %d KB
  系统内存: %d KB
  堆内存: %d KB
  栈内存: %d KB
  下次GC目标: %d KB
  GC次数: %d
  最后GC时间: %s

分析配置:
  采样间隔: %s
  CPU采样时间: %s
  输出目录: %s
  保存历史: %t
`,
		time.Now().Format("2006-01-02 15:04:05"),
		runtime.Version(),
		runtime.GOOS,
		runtime.GOARCH,
		runtime.NumCPU(),
		runtime.NumGoroutine(),
		memStats.Alloc/1024,
		memStats.Sys/1024,
		memStats.HeapAlloc/1024,
		memStats.StackInuse/1024,
		memStats.NextGC/1024,
		memStats.NumGC,
		time.Unix(0, int64(memStats.LastGC)).Format("2006-01-02 15:04:05"),
		a.config.AutoAnalysis.Interval,
		a.config.AutoAnalysis.CPUSampleDuration,
		a.config.AutoAnalysis.OutputDir,
		a.config.AutoAnalysis.SaveHistory,
	)

	return os.WriteFile(outputFile, []byte(info), 0644)
}

// isGoAvailable 检查go命令是否可用
func (a *Analyzer) isGoAvailable() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// cleanupOldData 清理旧数据
func (a *Analyzer) cleanupOldData() error {
	if a.config.AutoAnalysis.HistoryRetentionDays <= 0 {
		return nil
	}

	cutoffTime := time.Now().AddDate(0, 0, -a.config.AutoAnalysis.HistoryRetentionDays)

	entries, err := os.ReadDir(a.config.AutoAnalysis.OutputDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// 解析目录名中的时间戳
		dirTime, err := time.Parse("20060102_150405", entry.Name())
		if err != nil {
			continue
		}

		if dirTime.Before(cutoffTime) {
			dirPath := filepath.Join(a.config.AutoAnalysis.OutputDir, entry.Name())
			if err := os.RemoveAll(dirPath); err != nil {
				logger.Error("删除旧数据目录失败", "path", dirPath, "error", err)
			} else {
				logger.Info("已删除旧数据目录", "path", dirPath)
			}
		}
	}

	return nil
}

// GetAnalysisHistory 获取分析历史
func (a *Analyzer) GetAnalysisHistory() ([]string, error) {
	entries, err := os.ReadDir(a.config.AutoAnalysis.OutputDir)
	if err != nil {
		return nil, err
	}

	var history []string
	for _, entry := range entries {
		if entry.IsDir() {
			// 验证是否为有效的时间戳格式
			if _, err := time.Parse("20060102_150405", entry.Name()); err == nil {
				history = append(history, entry.Name())
			}
		}
	}

	return history, nil
}

// GetAnalysisReport 获取指定时间的分析报告
func (a *Analyzer) GetAnalysisReport(timestamp string) (map[string]string, error) {
	reportDir := filepath.Join(a.config.AutoAnalysis.OutputDir, timestamp)

	if _, err := os.Stat(reportDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("分析报告不存在: %s", timestamp)
	}

	reports := make(map[string]string)

	// 读取所有报告文件
	entries, err := os.ReadDir(reportDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if filepath.Ext(filename) == ".txt" {
			content, err := os.ReadFile(filepath.Join(reportDir, filename))
			if err != nil {
				logger.Error("读取报告文件失败", "file", filename, "error", err)
				continue
			}
			reports[filename] = string(content)
		}
	}

	return reports, nil
}
