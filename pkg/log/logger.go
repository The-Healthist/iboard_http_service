package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// 日志级别
type LogLevel int

const (
	// 日志级别常量
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL

	// 默认配置
	DefaultLogDir      = "logs"
	DefaultMaxFileSize = 1 * 1024 * 1024 // 1MB
	DefaultMaxFiles    = 365             // 保留365天的日志
	DefaultTimeFormat  = "2006-01-02"    // 日志文件日期格式
)

// 日志级别对应的字符串
var levelNames = map[LogLevel]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// Logger 日志记录器
type Logger struct {
	level        LogLevel    // 日志级别
	logDir       string      // 日志目录
	maxFileSize  int64       // 单个日志文件最大大小
	maxFiles     int         // 最大保留的日志文件数量
	timeFormat   string      // 日志文件时间格式
	logFile      *os.File    // 当前日志文件
	stdLogger    *log.Logger // 标准库日志记录器
	multiWriters []io.Writer // 多输出目标
	logFileName  string      // 当前日志文件名
}

// 日志配置选项
type Option func(*Logger)

// 设置日志级别
func WithLevel(level LogLevel) Option {
	return func(l *Logger) {
		l.level = level
	}
}

// GetProjectRoot 获取项目根目录
func GetProjectRoot() string {
	workDir, _ := os.Getwd()
	// 如果当前在cmd/server目录下，则需要回到项目根目录
	if strings.HasSuffix(workDir, "cmd/server") {
		return filepath.Join(workDir, "../..")
	}
	return workDir
}

// 设置日志目录
func WithLogDir(dir string) Option {
	return func(l *Logger) {
		// 如果是相对路径，则相对于项目根目录
		if !filepath.IsAbs(dir) {
			l.logDir = filepath.Join(GetProjectRoot(), dir)
		} else {
			l.logDir = dir
		}
	}
}

// 设置单个日志文件最大大小
func WithMaxFileSize(size int64) Option {
	return func(l *Logger) {
		l.maxFileSize = size
	}
}

// 设置最大保留的日志文件数量
func WithMaxFiles(count int) Option {
	return func(l *Logger) {
		l.maxFiles = count
	}
}

// 设置日志文件时间格式
func WithTimeFormat(format string) Option {
	return func(l *Logger) {
		l.timeFormat = format
	}
}

// 添加输出目标
func WithWriter(writer io.Writer) Option {
	return func(l *Logger) {
		l.multiWriters = append(l.multiWriters, writer)
	}
}

// NewLogger 创建新的日志记录器
func NewLogger(options ...Option) (*Logger, error) {
	logger := &Logger{
		level:       INFO,
		logDir:      DefaultLogDir,
		maxFileSize: DefaultMaxFileSize,
		maxFiles:    DefaultMaxFiles,
		timeFormat:  DefaultTimeFormat,
	}

	// 应用选项
	for _, option := range options {
		option(logger)
	}

	// 确保日志目录存在
	if err := os.MkdirAll(logger.logDir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 清理旧日志文件
	if err := logger.cleanupOldLogs(); err != nil {
		return nil, fmt.Errorf("清理旧日志失败: %v", err)
	}

	// 创建日志文件
	if err := logger.createLogFile(); err != nil {
		return nil, fmt.Errorf("创建日志文件失败: %v", err)
	}

	// 设置多输出目标
	writers := []io.Writer{os.Stdout}
	if logger.logFile != nil {
		writers = append(writers, logger.logFile)
	}
	for _, w := range logger.multiWriters {
		writers = append(writers, w)
	}
	multiWriter := io.MultiWriter(writers...)

	// 创建标准库日志记录器
	logger.stdLogger = log.New(multiWriter, "", log.LstdFlags)

	return logger, nil
}

// 创建日志文件
func (l *Logger) createLogFile() error {
	currentTime := time.Now().Format(l.timeFormat)
	l.logFileName = fmt.Sprintf("%s/app_%s.log", l.logDir, currentTime)

	// 检查当前日志文件是否存在及其大小
	if info, err := os.Stat(l.logFileName); err == nil {
		if info.Size() >= l.maxFileSize {
			// 如果文件存在且超过大小限制，创建带时间戳的新文件
			l.logFileName = fmt.Sprintf("%s/app_%s_%d.log", l.logDir, currentTime, time.Now().Unix())
		}
	}

	// 打开日志文件
	logFile, err := os.OpenFile(l.logFileName, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 如果之前有日志文件，关闭它
	if l.logFile != nil {
		l.logFile.Close()
	}

	l.logFile = logFile
	return nil
}

// 清理旧日志文件
func (l *Logger) cleanupOldLogs() error {
	files, err := os.ReadDir(l.logDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 目录不存在，不需要清理
		}
		return fmt.Errorf("读取日志目录失败: %v", err)
	}

	var logFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasPrefix(file.Name(), "app_") {
			logFiles = append(logFiles, filepath.Join(l.logDir, file.Name()))
		}
	}

	// 如果日志文件数量超过最大限制
	if len(logFiles) >= l.maxFiles {
		// 按修改时间排序（最旧的在前）
		sort.Slice(logFiles, func(i, j int) bool {
			iInfo, _ := os.Stat(logFiles[i])
			jInfo, _ := os.Stat(logFiles[j])
			return iInfo.ModTime().Before(jInfo.ModTime())
		})

		// 删除最旧的文件，直到数量低于限制
		for i := 0; i < len(logFiles)-l.maxFiles+1; i++ {
			if err := os.Remove(logFiles[i]); err != nil {
				return fmt.Errorf("删除旧日志文件 %s 失败: %v", logFiles[i], err)
			}
		}
	}

	return nil
}

// Close 关闭日志文件
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

// 检查日志文件大小，如果超过限制则创建新文件
func (l *Logger) checkRotate() error {
	if l.logFile == nil {
		return nil
	}

	info, err := os.Stat(l.logFileName)
	if err != nil {
		return fmt.Errorf("获取日志文件信息失败: %v", err)
	}

	if info.Size() >= l.maxFileSize {
		return l.createLogFile()
	}

	return nil
}

// log 记录日志的通用方法
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// 检查是否需要轮转日志文件
	if err := l.checkRotate(); err != nil {
		fmt.Fprintf(os.Stderr, "日志轮转失败: %v\n", err)
	}

	// 格式化日志级别前缀
	prefix := fmt.Sprintf("[%s] ", levelNames[level])

	// 格式化日志消息
	var msg string
	if format == "" {
		msg = fmt.Sprint(args...)
	} else {
		msg = fmt.Sprintf(format, args...)
	}

	// 记录日志
	l.stdLogger.Print(prefix + msg)

	// 如果是致命错误，退出程序
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug 记录调试级别日志
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info 记录信息级别日志
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn 记录警告级别日志
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error 记录错误级别日志
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal 记录致命错误日志并退出程序
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// 全局日志记录器实例
var defaultLogger *Logger

// 初始化默认日志记录器
func init() {
	var err error
	defaultLogger, err = NewLogger()
	if err != nil {
		log.Fatalf("初始化默认日志记录器失败: %v", err)
	}
}

// 全局日志函数

// Debug 记录调试级别日志
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info 记录信息级别日志
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn 记录警告级别日志
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error 记录错误级别日志
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal 记录致命错误日志并退出程序
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// SetLevel 设置默认日志记录器的级别
func SetLevel(level LogLevel) {
	defaultLogger.level = level
}

// GetLogger 获取默认日志记录器
func GetLogger() *Logger {
	return defaultLogger
}

// SetLogger 设置默认日志记录器
func SetLogger(logger *Logger) {
	if logger != nil {
		defaultLogger = logger
	}
}

// InitLogger 初始化默认日志记录器
func InitLogger(options ...Option) error {
	// 先关闭现有的日志记录器（如果存在）
	if defaultLogger != nil && defaultLogger.logFile != nil {
		defaultLogger.logFile.Close()
	}

	var err error
	defaultLogger, err = NewLogger(options...)
	if err != nil {
		return err
	}

	// 输出日志目录信息
	absLogDir, _ := filepath.Abs(defaultLogger.logDir)
	Info("日志系统初始化成功，日志将保存在: %s", absLogDir)

	return nil
}
