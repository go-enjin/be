package log

import "net/http"

/* Debug, DebugF, DebugDF, DebugR, DebugRF, DebugRDF */

func (c *Configuration) Debug(err error) {
	c._init()
	c.DebugRDF(nil, 1, "%v", err)
}

func (c *Configuration) DebugF(format string, argv ...interface{}) {
	c._init()
	c.DebugRDF(nil, 1, format, argv...)
}

func (c *Configuration) DebugDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.DebugRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) DebugR(r *http.Request, err error) {
	c._init()
	c.DebugRDF(r, 1, "%v", err)
}

func (c *Configuration) DebugRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.DebugRDF(r, depth+1, format, argv...)
}

func (c *Configuration) DebugRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Debugf, c.prefixLogEntry(depth+1, format, r), argv...)
}
