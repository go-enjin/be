package log

import "net/http"

/* Info, InfoF, InfoDF, InfoR, InfoRF, InfoRDF */

func (c *Configuration) Info(err error) {
	c._init()
	c.InfoRDF(nil, 1, "%v", err)
}

func (c *Configuration) InfoF(format string, argv ...interface{}) {
	c._init()
	c.InfoRDF(nil, 1, format, argv...)
}

func (c *Configuration) InfoDF(depth int, format string, argv ...interface{}) {
	c._init()
	c.InfoRDF(nil, depth+1, format, argv...)
}

func (c *Configuration) InfoR(r *http.Request, err error) {
	c._init()
	c.InfoRDF(r, 1, "%v", err)
}

func (c *Configuration) InfoRF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.InfoRDF(r, depth+1, format, argv...)
}

func (c *Configuration) InfoRDF(r *http.Request, depth int, format string, argv ...interface{}) {
	c._init()
	c.RLock()
	defer c.RUnlock()
	c.writeLogEntry(c.LogWith(r, nil).Infof, c.prefixLogEntry(depth+1, format, r), argv...)
}
