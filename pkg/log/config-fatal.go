package log

import "net/http"

/* Fatal, FatalF, FatalDF, FatalR, FatalRF, FatalRDF */

func (c *Configuration) Fatal(err error) {
	c._init()
	c.FatalRDF(nil, 1, "%v", err)
}

func (c *Configuration) FatalF(format string, argv ...interface{}) {
	c._init()
	c.FatalRDF(nil, 1, format, argv...)
}

func (c *Configuration) FatalDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.FatalRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) FatalR(r *http.Request, err error) {
	c._init()
	c.FatalRDF(r, 1, "%v", err)
}

func (c *Configuration) FatalRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.FatalRDF(r, depth+1, format, argv...)
}

func (c *Configuration) FatalRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Fatalf, c.prefixLogEntry(depth+1, format, r), argv...)
}
