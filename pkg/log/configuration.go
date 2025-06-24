package log

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/go-enjin/be/pkg/request"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/go-corelibs/context"
)

type Configuration struct {
	logger *logrus.Logger
	writer *log.Logger

	DisableTimestamp bool
	TimestampFormat  string
	LoggingFormat    Format
	LogLevel         Level

	LogHook  string
	LogHooks []logrus.Hook

	AppName    string
	RemoteHost string
	RemotePort int
	LogTag     string

	sync.RWMutex
}

func NewConfig() (c *Configuration) {
	c = &Configuration{
		DisableTimestamp: false,
		TimestampFormat:  DefaultTimestampFormat,
		LoggingFormat:    FormatPretty,
		LogLevel:         LevelInfo,
		LogHook:          "stdout",
		AppName:          "",
		RemoteHost:       "",
		RemotePort:       0,
		LogTag:           "",
	}
	c.logger = logrus.New()
	c.writer = log.New(c.logger.Writer(), c.AppName, 0)
	return
}

func (c *Configuration) Apply() {
	c.Lock()

	if c.logger == nil {
		c.logger = logrus.New()
	}

	switch c.LoggingFormat {
	case FormatJson:
		c.logger.SetFormatter(&logrus.JSONFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
		})
	case FormatText:
		c.logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
			DisableSorting:   true,
			DisableColors:    true,
			FullTimestamp:    !c.DisableTimestamp,
		})
	case FormatPretty:
		fallthrough
	default:
		c.logger.SetFormatter(&prefixed.TextFormatter{
			DisableTimestamp: c.DisableTimestamp,
			TimestampFormat:  c.TimestampFormat,
			ForceFormatting:  true,
			FullTimestamp:    true,
			DisableSorting:   true,
			DisableColors:    true,
		})
	}

	switch c.LogLevel {
	case LevelTrace:
		c.logger.SetLevel(logrus.TraceLevel)
	case LevelDebug:
		c.logger.SetLevel(logrus.DebugLevel)
	case LevelInfo:
		c.logger.SetLevel(logrus.InfoLevel)
	case LevelWarn:
		c.logger.SetLevel(logrus.WarnLevel)
	case LevelError:
		fallthrough
	default:
		c.logger.SetLevel(logrus.ErrorLevel)
	}

	if len(c.logger.Hooks) > 0 {
		// free existing hooks, we're about to replace them
		c.logger.Hooks = make(logrus.LevelHooks)
	}

	for _, h := range c.LogHooks {
		c.logger.AddHook(h)
	}

	c.writer = log.New(c.logger.Writer(), c.AppName, 0)
	c.Unlock()
	InfoF("%s log hook initialized", c.LogHook)
}

func (c *Configuration) Logrus() *logrus.Logger {
	c.RLock()
	defer c.RUnlock()
	return c.logger
}

func (c *Configuration) Logger() *log.Logger {
	c.RLock()
	defer c.RUnlock()
	if c.writer == nil {
		c.writer = log.New(c.writer.Writer(), "", 0)
	}
	return c.writer
}

func (c *Configuration) PrefixedLogger(prefix string) *log.Logger {
	c.RLock()
	defer c.RUnlock()
	return log.New(c.logger.Writer(), prefix, 0)
}

func (c *Configuration) _init() {
	if c.logger == nil {
		panic("internal logger should not be nil!")
	}
}

func (c *Configuration) LogWith(r *http.Request, ctx context.Context) (entry *logrus.Entry) {
	if ctx == nil {
		ctx = context.Context{}
	}
	if r != nil {

		ctx = context.Context{
			"rid":      "nil",
			"username": "-",
			"enjin-id": "-",
		}

		if r.URL.User != nil {
			if name := r.URL.User.Username(); name != "" {
				ctx["username"] = name
			}
		}

		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		ctx["host"] = host

		uri := r.RequestURI

		// Requests using the CONNECT method over HTTP/2.0 must use
		// the authority field (aka r.Host) to identify the target.
		// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
		if r.ProtoMajor == 2 && r.Method == "CONNECT" {
			uri = r.Host
		}
		if uri == "" {
			uri = r.URL.RequestURI()
		}
		ctx["uri"] = uri

		if rid := request.GetRequestID(r); rid != "" {
			ctx["rid"] = rid
		}

		ctx["req-host"] = r.Host
		if v := request.GetEnjinID(r); v != "" {
			ctx["enjin-id"] = v
		}

		if username, _, ok := r.BasicAuth(); ok {
			ctx["basic-auth"] = username
		}

		ctx["method"] = r.Method
		ctx["proto"] = r.Proto

	}
	return c.logger.WithFields(logrus.Fields(ctx))
}

func (c *Configuration) writeLogEntry(fn func(format string, args ...interface{}), prefixed string, argv ...interface{}) {
	if len(argv) == 0 {
		fn("%s", prefixed)
		return
	}
	fn(prefixed, argv...)
}

func (c *Configuration) prefixLogEntry(depth int, format string, r *http.Request) string {
	prefix := c.getLogPrefix(depth+1, r)
	return fmt.Sprintf("%v %v", prefix, format)
}
