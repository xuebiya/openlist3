package logger

import (
	"fmt"
	"strings"
	"time"
)

// TextFormatter 文本格式化器
type TextFormatter struct {
	TimestampFormat string
	ShowColors      bool
}

func NewTextFormatter() *TextFormatter {
	return &TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		ShowColors:      true,
	}
}

func (f *TextFormatter) Format(entry *LogEntry) string {
	var builder strings.Builder
	
	// 时间戳
	timestamp := entry.Timestamp.Format(f.TimestampFormat)
	builder.WriteString(fmt.Sprintf("[%s]", timestamp))
	
	// 日志类型和级别
	if f.ShowColors {
		builder.WriteString(f.colorizeType(entry.Type))
	} else {
		builder.WriteString(fmt.Sprintf(" [%s]", strings.ToUpper(string(entry.Type))))
	}
	
	// 根据日志类型格式化内容
	switch entry.Type {
	case TypeAccess:
		f.formatAccess(&builder, entry)
	case TypeMedia:
		f.formatMedia(&builder, entry)
	case TypeError:
		f.formatError(&builder, entry)
	case TypeSystem:
		f.formatSystem(&builder, entry)
	default:
		builder.WriteString(fmt.Sprintf(" %s", entry.Message))
	}
	
	return builder.String()
}

func (f *TextFormatter) colorizeType(logType LogType) string {
	switch logType {
	case TypeAccess:
		return " \033[36m[ACCESS]\033[0m" // 青色
	case TypeMedia:
		return " \033[35m[MEDIA]\033[0m"  // 紫色
	case TypeError:
		return " \033[31m[ERROR]\033[0m"  // 红色
	case TypeSystem:
		return " \033[32m[SYSTEM]\033[0m" // 绿色
	default:
		return fmt.Sprintf(" [%s]", strings.ToUpper(string(logType)))
	}
}

func (f *TextFormatter) formatAccess(builder *strings.Builder, entry *LogEntry) {
	// IP地址:X.X.X.X 方法:GET 路径:/api/test 状态:200 耗时:15ms
	builder.WriteString(fmt.Sprintf(" IP地址:%s", entry.IP))
	builder.WriteString(fmt.Sprintf(" 方法:%s", entry.Method))
	builder.WriteString(fmt.Sprintf(" 路径:%s", entry.Path))
	
	if entry.StatusCode > 0 {
		builder.WriteString(fmt.Sprintf(" 状态:%d", entry.StatusCode))
	}
	
	if entry.ResponseTime > 0 {
		builder.WriteString(fmt.Sprintf(" 耗时:%dms", entry.ResponseTime))
	}
	
	if entry.Username != "" {
		builder.WriteString(fmt.Sprintf(" 用户:%s", entry.Username))
	}
}

func (f *TextFormatter) formatMedia(builder *strings.Builder, entry *LogEntry) {
	// 时间:2025年1月X日 IP地址:X.X.X.X 用户:XXX 访问路径:XXX 文件大小:XXXMb
	date := entry.Timestamp.Format("2006年1月2日")
	builder.WriteString(fmt.Sprintf(" 时间:%s", date))
	builder.WriteString(fmt.Sprintf(" IP地址:%s", entry.IP))
	
	if entry.Username != "" {
		builder.WriteString(fmt.Sprintf(" 用户:%s", entry.Username))
	} else {
		builder.WriteString(" 用户:匿名")
	}
	
	builder.WriteString(fmt.Sprintf(" 访问路径:%s", entry.Path))
	
	if entry.FileSize > 0 {
		size := f.formatFileSize(entry.FileSize)
		builder.WriteString(fmt.Sprintf(" 文件大小:%s", size))
	}
}

func (f *TextFormatter) formatError(builder *strings.Builder, entry *LogEntry) {
	builder.WriteString(fmt.Sprintf(" IP地址:%s", entry.IP))
	builder.WriteString(fmt.Sprintf(" 路径:%s", entry.Path))
	
	if entry.Message != "" {
		builder.WriteString(fmt.Sprintf(" 错误:%s", entry.Message))
	}
}

func (f *TextFormatter) formatSystem(builder *strings.Builder, entry *LogEntry) {
	if entry.Message != "" {
		builder.WriteString(fmt.Sprintf(" %s", entry.Message))
	}
	
	// 添加额外信息
	if entry.Extra != nil && len(entry.Extra) > 0 {
		for key, value := range entry.Extra {
			builder.WriteString(fmt.Sprintf(" %s:%v", key, value))
		}
	}
}

func (f *TextFormatter) formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"KB", "MB", "GB", "TB"}
	if exp >= len(units) {
		exp = len(units) - 1
	}
	
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// JSONFormatter JSON格式化器
type JSONFormatter struct{}

func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

func (f *JSONFormatter) Format(entry *LogEntry) string {
	// 简单的JSON序列化，实际项目中可以使用json包
	return fmt.Sprintf(`{"timestamp":"%s","type":"%s","ip":"%s","method":"%s","path":"%s","status_code":%d,"response_time":%d,"username":"%s","message":"%s"}%s`,
		entry.Timestamp.Format(time.RFC3339),
		entry.Type,
		entry.IP,
		entry.Method,
		entry.Path,
		entry.StatusCode,
		entry.ResponseTime,
		entry.Username,
		entry.Message,
		"\n")
}
