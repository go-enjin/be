package log

import "net/http"

/* Trace, TraceF, TraceDF, TraceR, TraceRF, TraceRDF */

func (c *Configuration) Trace(err error) {
	c._init()
	c.TraceRDF(nil, 1, "%v", err)
}

func (c *Configuration) TraceF(format string, argv ...interface{}) {
	c._init()
	c.TraceRDF(nil, 1, format, argv...)
}

func (c *Configuration) TraceDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.TraceRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) TraceR(r *http.Request, err error) {
	c._init()
	c.TraceRDF(r, 1, "%v", err)
}

func (c *Configuration) TraceRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.TraceRDF(r, depth+1, format, argv...)
}

func (c *Configuration) TraceRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Tracef, c.prefixLogEntry(depth+1, format, r), argv...)
}
