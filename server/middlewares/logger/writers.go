package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ConsoleWriter 控制台输出写入器
type ConsoleWriter struct {
	mu        sync.Mutex
	output    io.Writer
	formatter LogFormatter
}

func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		output:    os.Stdout,
		formatter: NewTextFormatter(),
	}
}

func (w *ConsoleWriter) SetFormatter(formatter LogFormatter) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.formatter = formatter
}

func (w *ConsoleWriter) Write(entry *LogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.formatter != nil {
		formatted := w.formatter.Format(entry)
		_, err := fmt.Fprint(w.output, formatted)
		return err
	}
	
	// 默认格式化
	_, err := fmt.Fprintf(w.output, "[%s] %s %s\n", 
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		strings.ToUpper(string(entry.Type)),
		entry.Message)
	return err
}

func (w *ConsoleWriter) Close() error {
	return nil
}

// FileWriter 文件输出写入器
type FileWriter struct {
	mu           sync.Mutex
	file         *os.File
	filename     string
	maxSize      int64  // 最大文件大小（字节）
	maxFiles     int    // 最大文件数量
	currentSize  int64
}

func NewFileWriter(filename string, maxSize int64, maxFiles int) (*FileWriter, error) {
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %v", err)
	}
	
	// 获取当前文件大小
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}
	
	return &FileWriter{
		file:        file,
		filename:    filename,
		maxSize:     maxSize,
		maxFiles:    maxFiles,
		currentSize: info.Size(),
	}, nil
}

func (w *FileWriter) Write(entry *LogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// 使用默认格式化
	content := fmt.Sprintf("[%s] %s IP:%s Method:%s Path:%s Status:%d User:%s\n",
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		strings.ToUpper(string(entry.Type)),
		entry.IP,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.Username)
	
	// 检查是否需要轮转文件
	if w.currentSize+int64(len(content)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return fmt.Errorf("文件轮转失败: %v", err)
		}
	}
	
	n, err := w.file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}
	
	w.currentSize += int64(n)
	return w.file.Sync()
}

func (w *FileWriter) rotate() error {
	// 关闭当前文件
	w.file.Close()
	
	// 轮转文件名
	for i := w.maxFiles - 1; i > 0; i-- {
		oldName := fmt.Sprintf("%s.%d", w.filename, i)
		newName := fmt.Sprintf("%s.%d", w.filename, i+1)
		
		if i == w.maxFiles-1 {
			// 删除最老的文件
			os.Remove(newName)
		}
		
		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}
	
	// 重命名当前文件
	os.Rename(w.filename, fmt.Sprintf("%s.1", w.filename))
	
	// 创建新文件
	file, err := os.OpenFile(w.filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	w.file = file
	w.currentSize = 0
	return nil
}

func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// MultiWriter 多写入器 - 同时写入多个目标
type MultiWriter struct {
	writers []LogWriter
}

func NewMultiWriter(writers ...LogWriter) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

func (w *MultiWriter) AddWriter(writer LogWriter) {
	w.writers = append(w.writers, writer)
}

func (w *MultiWriter) Write(entry *LogEntry) error {
	for _, writer := range w.writers {
		if err := writer.Write(entry); err != nil {
			// 记录错误但继续写入其他writer
			fmt.Printf("Writer error: %v\n", err)
		}
	}
	return nil
}

func (w *MultiWriter) Close() error {
	var lastErr error
	for _, writer := range w.writers {
		if err := writer.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// AsyncWriter 异步写入器 - 在后台异步写入日志
type AsyncWriter struct {
	writer   LogWriter
	ch       chan *LogEntry
	done     chan struct{}
	wg       sync.WaitGroup
	bufSize  int
}

func NewAsyncWriter(writer LogWriter, bufSize int) *AsyncWriter {
	w := &AsyncWriter{
		writer:  writer,
		ch:      make(chan *LogEntry, bufSize),
		done:    make(chan struct{}),
		bufSize: bufSize,
	}
	
	w.wg.Add(1)
	go w.worker()
	
	return w
}

func (w *AsyncWriter) Write(entry *LogEntry) error {
	select {
	case w.ch <- entry:
		return nil
	default:
		// 缓冲区满时直接返回错误（或者可以选择阻塞）
		return fmt.Errorf("日志缓冲区已满")
	}
}

func (w *AsyncWriter) worker() {
	defer w.wg.Done()
	
	for {
		select {
		case entry := <-w.ch:
			if err := w.writer.Write(entry); err != nil {
				fmt.Printf("Async writer error: %v\n", err)
			}
		case <-w.done:
			// 处理剩余的日志条目
			for {
				select {
				case entry := <-w.ch:
					w.writer.Write(entry)
				default:
					return
				}
			}
		}
	}
}

func (w *AsyncWriter) Close() error {
	close(w.done)
	w.wg.Wait()
	return w.writer.Close()
}

// RotatingFileWriter 基于时间的轮转文件写入器
type RotatingFileWriter struct {
	mu            sync.Mutex
	file          *os.File
	filenamePattern string
	currentDate   string
	maxFiles      int
}

func NewRotatingFileWriter(filenamePattern string, maxFiles int) (*RotatingFileWriter, error) {
	w := &RotatingFileWriter{
		filenamePattern: filenamePattern,
		maxFiles:        maxFiles,
	}
	
	if err := w.openTodayFile(); err != nil {
		return nil, err
	}
	
	return w, nil
}

func (w *RotatingFileWriter) openTodayFile() error {
	today := time.Now().Format("2006-01-02")
	
	if w.currentDate == today && w.file != nil {
		return nil
	}
	
	// 关闭旧文件
	if w.file != nil {
		w.file.Close()
	}
	
	// 构造今天的文件名
	filename := fmt.Sprintf(w.filenamePattern, today)
	
	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}
	
	// 打开文件
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}
	
	w.file = file
	w.currentDate = today
	
	// 清理旧文件
	go w.cleanOldFiles()
	
	return nil
}

func (w *RotatingFileWriter) cleanOldFiles() {
	// 简化的清理逻辑，实际实现应该更完善
	// 这里只是示例
}

func (w *RotatingFileWriter) Write(entry *LogEntry) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	// 检查是否需要切换到新的日期文件
	if err := w.openTodayFile(); err != nil {
		return err
	}
	
	content := fmt.Sprintf("[%s] %s IP:%s Method:%s Path:%s Status:%d User:%s\n",
		entry.Timestamp.Format("2006-01-02 15:04:05"),
		strings.ToUpper(string(entry.Type)),
		entry.IP,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.Username)
	_, err := w.file.WriteString(content)
	if err != nil {
		return fmt.Errorf("写入日志失败: %v", err)
	}
	
	return w.file.Sync()
}

func (w *RotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
