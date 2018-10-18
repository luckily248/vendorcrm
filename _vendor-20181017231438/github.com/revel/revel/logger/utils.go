package logger

import (
	"fmt"
	"github.com/revel/config"
	"github.com/revel/log15"
	"gopkg.in/stack.v0"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Utility package to make existing logging backwards compatible
var (
	// Convert the string to LogLevel
	toLevel = map[string]LogLevel{"debug": LogLevel(log15.LvlDebug),
		"info": LogLevel(log15.LvlInfo), "request": LogLevel(log15.LvlInfo), "warn": LogLevel(log15.LvlWarn),
		"error": LogLevel(log15.LvlError), "crit": LogLevel(log15.LvlCrit),
		"trace": LogLevel(log15.LvlDebug), // TODO trace is deprecated, replaced by debug
	}
)
const (
	// The test mode flag overrides the default log level and shows only errors
	TEST_MODE_FLAG = "testModeFlag"
	// The special use flag enables showing messages when the logger is setup
	SPECIAL_USE_FLAG = "specialUseFlag"

// Returns the logger for the name
)

func GetLogger(name string, logger MultiLogger) (l *log.Logger) {
	switch name {
	case "trace": // TODO trace is deprecated, replaced by debug
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlDebug}, "", 0)
	case "debug":
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlDebug}, "", 0)
	case "info":
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlInfo}, "", 0)
	case "warn":
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlWarn}, "", 0)
	case "error":
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlError}, "", 0)
	case "request":
		l = log.New(loggerRewrite{Logger: logger, Level: log15.LvlInfo}, "", 0)
	}

	return l

}

// Get all handlers based on the Config (if available)
func InitializeFromConfig(basePath string, config *config.Context) (c *CompositeMultiHandler) {
	// If running in test mode suppress anything that is not an error
	if  config!=nil && config.BoolDefault(TEST_MODE_FLAG,false) {
		// Preconfigure all the options
		config.SetOption("log.info.output","none")
		config.SetOption("log.debug.output","none")
		config.SetOption("log.warn.output","none")
		config.SetOption("log.error.output","stderr")
		config.SetOption("log.crit.output","stderr")
	}


	// If the configuration has an all option we can skip some
	c, _ = NewCompositeMultiHandler()

	// Filters are assigned first, non filtered items override filters
	if config!=nil && !config.BoolDefault(TEST_MODE_FLAG,false) {
		initAllLog(c, basePath, config)
	}
	initLogLevels(c, basePath, config)
	if c.CriticalHandler == nil && c.ErrorHandler != nil {
		c.CriticalHandler = c.ErrorHandler
	}
	if config!=nil && !config.BoolDefault(TEST_MODE_FLAG,false) {
		initFilterLog(c, basePath, config)
		if c.CriticalHandler == nil && c.ErrorHandler != nil {
			c.CriticalHandler = c.ErrorHandler
		}
		initRequestLog(c, basePath, config)
	}

	return c
}

// Init the log.all configuration options
func initAllLog(c *CompositeMultiHandler, basePath string, config *config.Context) {
	if config != nil {
		extraLogFlag := config.BoolDefault(SPECIAL_USE_FLAG, false)
		if output, found := config.String("log.all.output"); found {
			// Set all output for the specified handler
			if extraLogFlag {
				log.Printf("Adding standard handler for levels to >%s< ", output)
			}
			initHandlerFor(c, output, basePath, NewLogOptions(config, true, nil, LvlAllList...))
		}
	}
}

// Used by the initFilterLog to handle the filters
var logFilterList = []struct {
	LogPrefix, LogSuffix string
	parentHandler        func(map[string]interface{}) ParentLogHandler
}{{
	"log.", ".filter",
	func(keyMap map[string]interface{}) ParentLogHandler {
		return NewParentLogHandler(func(child LogHandler) LogHandler {
			return MatchMapHandler(keyMap, child)
		})

	},
}, {
	"log.", ".nfilter",
	func(keyMap map[string]interface{}) ParentLogHandler {
		return NewParentLogHandler(func(child LogHandler) LogHandler {
			return NotMatchMapHandler(keyMap, child)
		})
	},
}}

// Init the filter options
// log.all.filter ....
// log.error.filter ....
func initFilterLog(c *CompositeMultiHandler, basePath string, config *config.Context) {

	if config != nil {
		extraLogFlag := config.BoolDefault(SPECIAL_USE_FLAG, false)

		// The commands to use
		logFilterList := []struct {
			LogPrefix, LogSuffix string
			parentHandler        func(map[string]interface{}) ParentLogHandler
		}{{
			"log.", ".filter",
			func(keyMap map[string]interface{}) ParentLogHandler {
				return NewParentLogHandler(func(child LogHandler) LogHandler {
					return MatchMapHandler(keyMap, child)
				})

			},
		}, {
			"log.", ".nfilter",
			func(keyMap map[string]interface{}) ParentLogHandler {
				return NewParentLogHandler(func(child LogHandler) LogHandler {
					return NotMatchMapHandler(keyMap, child)
				})
			},
		}}

		for _, logFilter := range logFilterList {
			// Init for all filters
			for _, name := range []string{"all", "debug", "info", "warn", "error", "crit",
				"trace", // TODO trace is deprecated
			} {
				optionList := config.Options(logFilter.LogPrefix + name + logFilter.LogSuffix)
				for _, option := range optionList {
					splitOptions := strings.Split(option, ".")
					keyMap := map[string]interface{}{}
					for x := 3; x < len(splitOptions); x += 2 {
						keyMap[splitOptions[x]] = splitOptions[x+1]
					}
					phandler := logFilter.parentHandler(keyMap)
					if extraLogFlag {
						log.Printf("Adding key map handler %s %s output %s", option, name, config.StringDefault(option, ""))
					}

					if name == "all" {
						initHandlerFor(c, config.StringDefault(option, ""), basePath, NewLogOptions(config, false, phandler))
					} else {
						initHandlerFor(c, config.StringDefault(option, ""), basePath, NewLogOptions(config, false, phandler, toLevel[name]))
					}
				}
			}
		}
	}
}

// Init the log.error, log.warn etc configuration options
func initLogLevels(c *CompositeMultiHandler, basePath string, config *config.Context) {
	for _, name := range []string{"debug", "info", "warn", "error", "crit",
		"trace", // TODO trace is deprecated
	} {
		if config != nil {
			extraLogFlag := config.BoolDefault(SPECIAL_USE_FLAG, false)
			output, found := config.String("log." + name + ".output")
			if found {
				if extraLogFlag {
					log.Printf("Adding standard handler %s output %s", name, output)
				}
				initHandlerFor(c, output, basePath, NewLogOptions(config, true, nil, toLevel[name]))
			}
			// Gets the list of options with said prefix
		} else {
			initHandlerFor(c, "stderr", basePath, NewLogOptions(config, true, nil, toLevel[name]))
		}
	}
}

// Init the request log options
func initRequestLog(c *CompositeMultiHandler, basePath string, config *config.Context) {
	// Request logging to a separate output handler
	// This takes the InfoHandlers and adds a MatchAbHandler handler to it to direct
	// context with the word "section=requestlog" to that handler.
	// Note if request logging is not enabled the MatchAbHandler will not be added and the
	// request log messages will be sent out the INFO handler
	outputRequest := "stdout"
	if config != nil {
		outputRequest = config.StringDefault("log.request.output", "")
	}
	oldInfo := c.InfoHandler
	c.InfoHandler = nil
	if outputRequest != "" {
		initHandlerFor(c, outputRequest, basePath, NewLogOptions(config, false, nil, LvlInfo))
	}
	if c.InfoHandler != nil || oldInfo != nil {
		if c.InfoHandler == nil {
			c.InfoHandler = oldInfo
		} else {
			c.InfoHandler = MatchAbHandler("section", "requestlog", c.InfoHandler, oldInfo)
		}
	}
}

// Returns a handler for the level using the output string
// Accept formats for output string are
// LogFunctionMap[value] callback function
// `stdout` `stderr` `full/file/path/to/location/app.log` `full/file/path/to/location/app.json`
func initHandlerFor(c *CompositeMultiHandler, output, basePath string, options *LogOptions) {
	if options.Ctx != nil {
		options.SetExtendedOptions(
			"noColor", !options.Ctx.BoolDefault("log.colorize", true),
			"smallDate", options.Ctx.BoolDefault("log.smallDate", true),
			"maxSize", options.Ctx.IntDefault("log.maxsize", 1024*10),
			"maxAge", options.Ctx.IntDefault("log.maxage", 14),
			"maxBackups", options.Ctx.IntDefault("log.maxbackups", 14),
			"compressBackups", !options.Ctx.BoolDefault("log.compressBackups", true),
		)
	}

	output = strings.TrimSpace(output)
	if funcHandler, found := LogFunctionMap[output]; found {
		funcHandler(c, options)
	} else {
		switch output {
		case "":
			fallthrough
		case "off":
			// No handler, discard data
		default:
			// Write to file specified
			if !filepath.IsAbs(output) {
				output = filepath.Join(basePath, output)
			}

			if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
				log.Panic(err)
			}

			if strings.HasSuffix(output, "json") {
				c.SetJsonFile(output, options)
			} else {
				// Override defaults for a terminal file
				options.SetExtendedOptions("noColor", true)
				options.SetExtendedOptions("smallDate", false)
				c.SetTerminalFile(output, options)
			}
		}
	}
	return
}

// This structure and method will handle the old output format and log it to the new format
type loggerRewrite struct {
	Logger         MultiLogger
	Level          log15.Lvl
	hideDeprecated bool
}

// The message indicating that a logger is using a deprecated log mechanism
var log_deprecated = []byte("* LOG DEPRECATED * ")

// Implements the Write of the logger
func (lr loggerRewrite) Write(p []byte) (n int, err error) {
	if !lr.hideDeprecated {
		p = append(log_deprecated, p...)
	}
	n = len(p)
	if len(p) > 0 && p[n-1] == '\n' {
		p = p[:n-1]
		n--
	}

	switch lr.Level {
	case log15.LvlInfo:
		lr.Logger.Info(string(p))
	case log15.LvlDebug:
		lr.Logger.Debug(string(p))
	case log15.LvlWarn:
		lr.Logger.Warn(string(p))
	case log15.LvlError:
		lr.Logger.Error(string(p))
	case log15.LvlCrit:
		lr.Logger.Crit(string(p))
	}

	return
}

// For logging purposes the call stack can be used to record the stack trace of a bad error
// simply pass it as a context field in your log statement like
// `controller.Log.Critc("This should not occur","stack",revel.NewCallStack())`
func NewCallStack() interface{} {
	return stack.Trace()
}

// Currently the only requirement for the callstack is to support the Formatter method
// which stack.Call does
type CallStack interface {
	fmt.Formatter

}
// Call to define a handler function for the logs
func HandlerFunc(log func(message string, time time.Time, level LogLevel, call CallStack, context ContextMap) error) LogHandler {
	return callHandler(log)
}

// The function wrapper to implement the callback
type callHandler func(message string, time time.Time, level LogLevel, call CallStack, context ContextMap) error

// Log implementation, reads the record and extracts the details from the log record
// Hiding the implementation.
func (c callHandler) Log(log *log15.Record) error {
	ctx := log.Ctx
	var ctxMap ContextMap
	if len(ctx) > 0 {
		ctxMap = make(ContextMap, len(ctx)/2)

		for i := 0; i < len(ctx); i += 2 {
			v := ctx[i]
			key, ok := v.(string)
			if !ok {
				key = fmt.Sprintf("LOGGER_INVALID_KEY %v", v)
			}
			var value interface{}
			if len(ctx) > i+1 {
				value = ctx[i+1]
			} else {
				value = "LOGGER_VALUE_MISSING"
			}
			ctxMap[key] = value
		}
	}

	return c(log.Msg, log.Time, LogLevel(log.Lvl), CallStack(log.Call), ctxMap)
}

// Internally used contextMap, allows conversion of map to map[string]string
type ContextMap map[string]interface{}

// Convert the context map to be string only values, any non string values are ignored
func (m ContextMap) StringMap() (newMap map[string]string) {
	if m != nil {
		newMap = map[string]string{}
		for key, value := range m {
			if svalue, isstring := value.(string); isstring {
				newMap[key] = svalue
			}
		}
	}
	return
}
