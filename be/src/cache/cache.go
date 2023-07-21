/*
The cache uses file system links and hashing to store files with identical content in a single file.
The layout is:

cache-dir
	.content
	sym-link --> .content/file-content-hash

Useful for images, but useless for blog entries, which have distinct content.

TODO:
1.  The blog content is markdown on disk, which gets rendered into html.
    Its is the html that should ne cached, not the markdown,
*/
package cache

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"time"

	log "main/logger"
)

// const contentDir = ".content"

type Cache struct {
	path string 					// root of the cache directory
	maxsize int 					// start flushing when maxsize pages are in the cache
	cache map[string]*cacheEntry	// the actual in-memory cache
}

type cacheEntry struct {
	raw []byte
	rendered []byte
	updated bool
	lastused int64
}

func New(path string, maxsize int) *Cache {
	if len(path) == 0 {
		log.Errorf("empty path")
		return nil
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	return &Cache{ path: path, maxsize: maxsize, cache: make(map[string]*cacheEntry) }
}

func (c *Cache) Get(key string) []byte {
	// log.Tracef("cache.Get(%s)", key)
	
	// key is a symlink to the actual content
	entry, ok := c.cache[key]
	if !ok {
		entry = c.read(key)
		if entry == nil {
			log.Errorf("failed to read %s", key)
			return nil
		}
	}

	if entry.updated {
		return entry.rendered
	} else {
		return entry.raw
	}
}

func (c *Cache) Updated(key string) bool {
	// log.Tracef("cache.Updated(%s)", key)

	entry, ok := c.cache[key]
	return ok && entry.updated
}

func (c *Cache) Update(key string, content []byte) error {
	// log.Tracef("cache.Update(%s)", key)

	entry, ok := c.cache[key]
	if !ok {
		return fmt.Errorf("not in cache")
	}

	entry.rendered = content
	entry.raw = []byte{}	// hopefully free the underlaying raw bytes
	entry.updated = true

	return nil
}

func (c *Cache) Put(key string, content []byte) error {
	// log.Tracef("cache.Put(%s)", key)

	sum := fmt.Sprintf("%x", md5.Sum(content))
	// contentPath := c.path + contentDir + "/" + sum
	contentPath := c.path + "/" + sum + ".md"
	keyPath := c.path + key

	contentExists := exists(contentPath)
	filenameExists := exists(keyPath)

	/*
	four cases:
	content new, file name new 			create content, create link
	content new, file name exists 		error, would trash
	content exists, file name new 		new name for existing content, create link
	content exists, file name exists 	content and name match, do nothing
	*/
	if filenameExists {
		if !contentExists {
			return fmt.Errorf("would trash %s", key)
		}
		return nil
	}

	if !contentExists {
		fh, err := os.Create(contentPath)
		if err != nil {
			return err
		}
		fh.Write(content)
		fh.Close()

		// the file system needs time to update (a known issue)
		time.Sleep(1 * time.Second)
	}

	return os.Symlink(contentPath, keyPath)
}

// file exists
func exists(path string) bool {
	fh, err := os.Open(path)
	exists := err == nil
	if exists {
		fh.Close()
	}
	return exists
}

func (c *Cache) read(key string) *cacheEntry {
	if len(c.cache) >= c.maxsize {
		c.flush()
	}

	// fh, err := os.Open(c.path + contentDir + "/" + key)
	fh, err := os.Open(c.path + "/" + key)
	if err != nil {
		log.Errorf("failed to open %s, %s", c.path + key, err)
		return nil
	}
	defer fh.Close()

	raw, err := io.ReadAll(fh)
	if err != nil {
		log.Errorf("failed to read %s, %s", c.path + key, err)
		return nil
	}

	entry := cacheEntry{ raw: raw, lastused: time.Now().Unix() }
	c.cache[key] = &entry

	return c.cache[key]
}

func (c *Cache) flush() {
	lastusedkey := ""
	lastused := time.Now().Unix()
	for key, entry := range c.cache {
		if entry.lastused < lastused {
			lastusedkey = key
			lastused = entry.lastused
		}
	}

	delete(c.cache, lastusedkey)
}

