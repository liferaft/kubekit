package terraformer

import (
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

// Logger interface defines the behaviour of a log instance. Any logger requires to implement these methods.
type Logger interface {
	Printf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	// Fatalf(format string, args ...interface{})
	// Panicf(format string, args ...interface{})

	// Debug(args ...interface{})
	// Info(args ...interface{})
	// Print(args ...interface{})
	// Warn(args ...interface{})
	// Error(args ...interface{})
	// Fatal(args ...interface{})
	// Panic(args ...interface{})

	// Debugln(args ...interface{})
	// Infoln(args ...interface{})
	// Println(args ...interface{})
	// Warnln(args ...interface{})
	// Errorln(args ...interface{})
	// Fatalln(args ...interface{})
	// Panicln(args ...interface{})
}

// Level is the log level type
type Level uint8

// Different log level from higher to lower.
// Lower log levels to the level set won't be displayed.
// i.e. If LogLevelWarn is set only Warnings and Errors are displayed.
// DefLogLevel defines the log level to assing by default
const (
	LogLevelError Level = iota
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug

	DefLogLevel = LogLevelInfo
)

const envPrintTrace = "TERRAFORMER_TRACE"

var printTrace bool

// var stdLogger *StdLogger

// Won't compile if StdLogger can't be realized by a  Logger
var (
	_ Logger = &StdLogger{}
)

func init() {
	valPrintTrace, _ := os.LookupEnv(envPrintTrace)
	printTrace = (strings.ToLower(valPrintTrace) == "true")

	// stdLogger := NewLogger(os.Stdout, "TERRAFORMER", LogLevelInfo)
}

// StdLog returns the standar logger
// func StdLog() Logger {
// 	return stdLogger
// }

// IsTraceOn returns the state of the trace flag
func IsTraceOn() bool {
	return printTrace
}

// ToggleTrace turn off the trace flag if it's on, and viseversa
func ToggleTrace() {
	// TODO: this may need the use of a mutex
	printTrace = !printTrace
}

// StdLogger implements a standard Logger
type StdLogger struct {
	Prefix string
	Level  Level
	logger *log.Logger
	mu     sync.Mutex
}

// LogWriter is an implementation of a io.Writer to capture all the TF logs which are send using the "log" package
type LogWriter struct {
	Logger    Logger
	oldLogger *log.Logger
	mu        sync.Mutex
}

// NewLogWriter creates a new instance of LogWriter with the Logger l or (if nil) a new Logger
func NewLogWriter(l Logger) *LogWriter {
	lw := &LogWriter{
		Logger: l,
	}
	if l == nil {
		l := NewLogger(os.Stdout, "", DefLogLevel)
		lw.Logger = l
	}
	return lw
}

// SetLogOut set the output to the standard golang log to use the LogWriter so it's possible to get the TF logs
func (lw *LogWriter) SetLogOut() {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	log.SetOutput(lw)
}

// RestoreLogOut restore the output of the standard golang log to StdErr as it was previously (by default)
func (lw *LogWriter) RestoreLogOut() {
	lw.mu.Lock()
	defer lw.mu.Unlock()
	log.SetOutput(os.Stderr)
}

// Writer captures all the output from Terraform and use the logger to print it out
func (lw *LogWriter) Write(p []byte) (n int, err error) {
	re := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+\[(\w+)\]\s+((?s:.+))`)
	allMatch := re.FindAllStringSubmatch(string(p), -1)

	if len(allMatch) > 0 {
		match := allMatch[0]
		logMessage := strings.TrimRight(match[2], "\n")
		if len(match) == 3 {
			switch match[1] {
			case "ERROR":
				lw.Logger.Errorf("%s", logMessage)
			case "WARN":
				lw.Logger.Warnf("%s", logMessage)
			case "INFO":
				lw.Logger.Infof("%s", logMessage)
			case "DEBUG":
				lw.Logger.Debugf("%s", logMessage)
			case "TRACE":
				if printTrace {
					lw.Logger.Debugf("\x1B[95m[TRACE]\x1B[0m %s", logMessage)
				}
			default:
				// If there is something like [x] in cyan, another case may be needed
				lw.Logger.Printf("\x1B[36m[%s]\x1B[0m %s", match[1], logMessage)
			}
		} else {
			// If there is something starting with >, improve the output
			lw.Logger.Printf("> %s", p)
		}
	} else {
		reDate := regexp.MustCompile(`\d{4}/\d{2}/\d{2}\s+\d{2}:\d{2}:\d{2}\s+(.+)`)
		allMatchDate := reDate.FindAllStringSubmatch(string(p), -1)
		matchDate := allMatchDate[0]
		if len(matchDate) == 2 {
			lw.Logger.Printf("\x1B[36m->\x1B[0m %s", matchDate[1])
		} else {
			// If there is something starting with -, improve the output
			lw.Logger.Printf("- %s", p)
		}
	}

	return len(p), nil
}

// NewStdLogger creates a standard logger, sending output to stdout at info level
func NewStdLogger() *StdLogger {
	return NewLogger(os.Stdout, "TERRAFORMER", LogLevelInfo)
}

// NewLogger creates an Logger based on log with default values
func NewLogger(w io.Writer, prefix string, level Level) *StdLogger {
	l := log.New(w, "", log.LstdFlags)
	return &StdLogger{
		Prefix: prefix,
		Level:  level,
		logger: l,
	}
}

// Printf implements a standard Printf function of Logger interface
func (l *StdLogger) Printf(format string, args ...interface{}) {
	l.output("     ", format, args...)
}

// Debugf implements a standard Debugf function of Logger interface
func (l *StdLogger) Debugf(format string, args ...interface{}) {
	if l.Level < LogLevelDebug {
		return
	}
	if printTrace && strings.HasPrefix(format, "TRACE:") {
		format = strings.TrimPrefix(format, "TRACE:")
		l.output("TRACE", format, args...)
		return
	}
	l.output("DEBUG", format, args...)
}

// Infof implements a standard Infof function of Logger interface
func (l *StdLogger) Infof(format string, args ...interface{}) {
	if l.Level < LogLevelInfo {
		return
	}
	l.output("INFO ", format, args...)
}

// Warnf implements a standard Warnf function of Logger interface
func (l *StdLogger) Warnf(format string, args ...interface{}) {
	if l.Level < LogLevelWarn {
		return
	}
	l.output("WARN ", format, args...)
}

// Errorf implements a standard Errorf function of Logger interface
func (l *StdLogger) Errorf(format string, args ...interface{}) {
	if l.Level < LogLevelError {
		return
	}
	l.output("ERROR", format, args...)
}

func (l *StdLogger) output(levelStr string, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	oldp := l.logger.Prefix()
	p := l.Prefix
	if len(p) != 0 {
		p = fmt.Sprintf("] %s: ", p)
	} else {
		p = "] "
	}
	l.logger.SetPrefix(levelStr + " [ ")
	l.logger.Printf(p+format, args...)
	l.logger.SetPrefix(oldp)
}
