package log

import "net/http"

/* Warn, WarnF, WarnDF, WarnR, WarnRF, WarnRDF */

func (c *Configuration) Warn(err error) {
	c._init()
	c.WarnRDF(nil, 1, "%v", err)
}

func (c *Configuration) WarnF(format string, argv ...interface{}) {
	c._init()
	c.WarnRDF(nil, 1, format, argv...)
}

func (c *Configuration) WarnDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.WarnRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) WarnR(r *http.Request, err error) {
	c._init()
	c.WarnRDF(r, 1, "%v", err)
}

func (c *Configuration) WarnRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.WarnRDF(r, depth+1, format, argv...)
}

func (c *Configuration) WarnRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Warnf, c.prefixLogEntry(depth+1, format, r), argv...)
}
