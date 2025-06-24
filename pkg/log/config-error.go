package log

import "net/http"

/* Error, ErrorF, ErrorDF, ErrorR, ErrorRF, ErrorRDF */

func (c *Configuration) Error(err error) {
	c._init()
	c.ErrorRDF(nil, 1, "%v", err)
}

func (c *Configuration) ErrorF(format string, argv ...interface{}) {
	c._init()
	c.ErrorRDF(nil, 1, format, argv...)
}

func (c *Configuration) ErrorDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.ErrorRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) ErrorR(r *http.Request, err error) {
	c._init()
	c.ErrorRDF(r, 1, "%v", err)
}

func (c *Configuration) ErrorRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.ErrorRDF(r, depth+1, format, argv...)
}

func (c *Configuration) ErrorRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Errorf, c.prefixLogEntry(depth+1, format, r), argv...)
}
