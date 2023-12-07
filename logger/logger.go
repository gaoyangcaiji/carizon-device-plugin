package logger

import (
	"io"
	"log"
	"os"
	"sync/atomic"

	"gopkg.in/natefinch/lumberjack.v2"
)

// LevelType is the log level enum
type LevelType int

func (s LevelType) String() string {
	switch s {
	case LevelTypeError:
		return "Error"
	case LevelTypeWarn:
		return "Warn"
	case LevelTypeInfo:
		return "Info"
	case LevelTypeDebug:
		return "Debug"
	case LevelTypeFATAL:
		return "Debug"
	default:
		return "Other"
	}
}

// logger supports 5 levels. Default level is LevelInformational.
const (
	LevelTypeDebug = iota
	LevelTypeInfo
	LevelTypeWarn
	LevelTypeError
	LevelTypeFATAL
)

// Logger is a log wrapper
type Logger struct {
	level       int64
	w           io.Writer
	debugLogger *log.Logger
	warnLogger  *log.Logger
	infoLogger  *log.Logger
	errLogger   *log.Logger
	fatalLogger *log.Logger
}

// Wrapper is the log instance
var Wrapper Logger

func init() {

	fileWriter := &lumberjack.Logger{
		Filename:   "log/error.log",
		MaxSize:    500, // megabytes
		MaxBackups: 2,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}

	stdWriter := os.Stdout

	mw := io.MultiWriter(stdWriter, fileWriter)

	Wrapper = Logger{
		w:           stdWriter,
		level:       LevelTypeInfo,
		debugLogger: log.New(stdWriter, "[DEBUG] ", log.Lshortfile|log.Ldate|log.Lmicroseconds|log.Lmsgprefix),
		warnLogger:  log.New(stdWriter, "[WARN] ", log.Lshortfile|log.Ldate|log.Lmicroseconds|log.Lmsgprefix),
		infoLogger:  log.New(stdWriter, "[INFO] ", log.Ldate|log.Lmicroseconds|log.Lmsgprefix),
		errLogger:   log.New(mw, "[ERROR] ", log.Lshortfile|log.Ldate|log.Lmicroseconds|log.Lmsgprefix),
		fatalLogger: log.New(mw, "[FATAL] ", log.Lshortfile|log.Ldate|log.Lmicroseconds|log.Lmsgprefix),
	}
}

// SetLevel ...
func (l *Logger) SetLevel(level int64) {
	if level < LevelTypeDebug || level > LevelTypeFATAL {
		return
	}

	atomic.StoreInt64(&l.level, level)
}

// Debugln ...
func (l *Logger) Debugln(v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeDebug {
		return
	}
	l.debugLogger.Println(v...)
}

// Debugf ...
func (l *Logger) Debugf(format string, v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeDebug {
		return
	}
	l.debugLogger.Printf(format, v...)
}

// Infoln ...
func (l *Logger) Infoln(v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeInfo {
		return
	}
	l.infoLogger.Println(v...)
}

// Infof ...
func (l *Logger) Infof(format string, v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeInfo {
		return
	}
	l.infoLogger.Printf(format, v...)
}

// Warnln ...
func (l *Logger) Warnln(v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeWarn {
		return
	}
	l.warnLogger.Println(v...)
}

// Warnf ...
func (l *Logger) Warnf(format string, v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeWarn {
		return
	}
	l.warnLogger.Printf(format, v...)
}

// Errorf ...
func (l *Logger) Errorf(format string, v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeError {
		return
	}
	l.errLogger.Printf(format, v...)
}

// Errorln ...
func (l *Logger) Errorln(v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeError {
		return
	}
	l.errLogger.Println(v...)
}

// Fatalf ...
func (l *Logger) Fatalf(format string, v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeFATAL {
		return
	}
	l.fatalLogger.Fatalf(format, v...)
}

// Fatalln ...
func (l *Logger) Fatalln(v ...interface{}) {
	if atomic.LoadInt64(&l.level) > LevelTypeFATAL {
		return
	}
	l.fatalLogger.Fatalln(v...)
}
