/*
The cache uses file system links and hashing to store files with identical content in a single file.
The layout is:

cache-dir
	.content
	sym-link --> .content/file-content-hash

Useful for images, but useless for blog entries, which have distinct content.
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

const contentDir = ".content"

type Cache struct {
	path string 					// root of the cache directory
	maxsize int 					//
	cache map[string]*cacheEntry	//
}

type cacheEntry struct {
	page []byte
	lastused int64
}

func New(path string, maxsize int) *Cache {
	if len(path) == 0 {
		log.Error("empty path")
		return nil
	}
	if path[len(path)-1] != '/' {
		path += "/"
	}
	return &Cache{ path: path, maxsize: maxsize, cache: make(map[string]*cacheEntry) }
}

func (c *Cache) Get(key string) []byte {
	// log.Trace("cache.Get(%s)", key)
	
	// key is a symlink to the actual content
	entry, ok := c.cache[key]
	if !ok {
		entry = c.read(key)
		if entry == nil {
			log.Error("failed to read %s", key)
			return nil
		}
	}

	return entry.page
}

func (c *Cache) Put(key string, content []byte) error {
	// log.Trace("cache.Put(%s)", key)

	sum := fmt.Sprintf("%x", md5.Sum(content))
	contentPath := c.path + contentDir + "/" + sum
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

	fh, err := os.Open(c.path + contentDir + "/" + key)
	if err != nil {
		log.Error("failed to open %s, %s", c.path + key, err)
		return nil
	}
	defer fh.Close()

	page, err := io.ReadAll(fh)
	if err != nil {
		log.Error("failed to read %s, %s", c.path + key, err)
		return nil
	}

	entry := cacheEntry{ page: page, lastused: time.Now().Unix() }
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

