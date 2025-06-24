package log

import "net/http"

/* Panic, PanicF, PanicDF, PanicR, PanicRF, PanicRDF */

func (c *Configuration) Panic(err error) {
	c._init()
	c.PanicRDF(nil, 1, "%v", err)
}

func (c *Configuration) PanicF(format string, argv ...interface{}) {
	c._init()
	c.PanicRDF(nil, 1, format, argv...)
}

func (c *Configuration) PanicDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.PanicRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) PanicR(r *http.Request, err error) {
	c._init()
	c.PanicRDF(r, 1, "%v", err)
}

func (c *Configuration) PanicRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.PanicRDF(r, depth+1, format, argv...)
}

func (c *Configuration) PanicRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Panicf, c.prefixLogEntry(depth+1, format, r), argv...)
}
