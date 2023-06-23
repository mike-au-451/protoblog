/*
Cache files in a directory
*/

package cache

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog/log"
)

type Cache struct {
	dir string
	maxcachesize int
	pages map[string]cachePage
}

type cachePage struct {
	page []byte
	lastused int64
}

func New(dir string) *Cache {
	return &Cache{
		dir: dir,
		pages: make(map[string]cachePage),
	}
}

func (c *Cache) SetSize(size int) *Cache {
		c.maxcachesize = size
		return c
}

func (c *Cache) Get(key string) ([]byte, bool) {
	// log.Trace().Msg(fmt.Sprintf("Get(%s)", key))
	page, ok := c.pages[key]
	if !ok {
		if !c.read(key) {
			log.Error().Msg(fmt.Sprintf("failed to get %s", key))
			return nil, false
		}

		page = c.pages[key]
	}

	page.lastused = time.Now().Unix()
	return page.page, true
}

func (c *Cache) read(key string) bool {
	// log.Trace().Msg(fmt.Sprintf("read(%s)", key))

	fh, err := os.Open(c.dir + "/" + key)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to open %s: %s", key, err))
		return false
	}
	defer fh.Close()

	raw, err := io.ReadAll(fh)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("failed to read %s: %s", key, err))
		return false
	}

	if len(c.pages) >= c.maxcachesize {
		c.flush()
	}
	cp := c.pages[key]
	cp.page = raw
	c.pages[key] = cp

	return true
}

func (c *Cache) flush() {
	// log.Trace().Msg(fmt.Sprintf("flush()"))

	if len(c.pages) == 0 {
		// first time through
		return
	}

	leastrecentkey := ""
	leastrecenttime := time.Now().Unix()
	for key, cp := range c.pages {
		if cp.lastused < leastrecenttime {
			leastrecenttime = cp.lastused
			leastrecentkey = key
		}
	}

	log.Trace().Msg(fmt.Sprintf("...flushing %s", leastrecentkey))
	delete(c.pages, leastrecentkey)
}