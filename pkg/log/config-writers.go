package log

import (
	"io"

	"github.com/sirupsen/logrus"
)

func (c *Configuration) ErrorWriter() *io.PipeWriter {
	c.RLock()
	defer c.RUnlock()
	return c.logger.WriterLevel(logrus.ErrorLevel)
}

func (c *Configuration) WarnWriter() *io.PipeWriter {
	c.RLock()
	defer c.RUnlock()
	return c.logger.WriterLevel(logrus.WarnLevel)
}

func (c *Configuration) InfoWriter() *io.PipeWriter {
	c.RLock()
	defer c.RUnlock()
	return c.logger.WriterLevel(logrus.InfoLevel)
}

func (c *Configuration) DebugWriter() *io.PipeWriter {
	c.RLock()
	defer c.RUnlock()
	return c.logger.WriterLevel(logrus.DebugLevel)
}

func (c *Configuration) TraceWriter() *io.PipeWriter {
	c.RLock()
	defer c.RUnlock()
	return c.logger.WriterLevel(logrus.TraceLevel)
}
