package source

import (
	"net/url"
	"sync"
)

// MetaCache caches the meta information which are returned by Source for the same URL of an arbitrary Client.
// Meta are cached across multiple Open calls for the same url.
type MetaCache struct {
	client Client

	//sourceOpened map[string]bool
	metaStore *metaStore
}

var _ Client = &MetaCache{}

func NewMetaCache(c Client) *MetaCache {
	return &MetaCache{
		client:    c,
		metaStore: &metaStore{},
	}
}

// Open opens an url.URL from the underlying Client when Open is invoked for the first time for the url. Subsequent Open calls postpone the underlying open when the Source is used as an io.Reader. This stems from the assumption that a once successfully opened Source can be opened likewise in all future Open calls.
func (c *MetaCache) Open(url *url.URL) (Source, error) {
	urlStr := url.String()

	mcs := &metaCacheSource{}

	// Retrieve meta from cache or from first open of an url
	m, err := c.metaStore.Get(urlStr, func() (Meta, error) {
		// First open of url
		s, err := c.client.Open(url)
		if err != nil {
			return nil, err
		}

		mcs.source = s

		return s.Meta()
	})
	if err != nil {
		return nil, err
	}

	if mcs.source == nil {
		// Recurring open of url
		mcs.sourceFunc = func() (Source, error) {
			return c.client.Open(url)
		}
	}

	mcs.meta = m

	return mcs, nil
}

func (c *MetaCache) Schemes() []string {
	return c.client.Schemes()
}

type metaCacheSource struct {
	meta Meta

	source     Source
	sourceFunc func() (Source, error)
}

var _ Source = &metaCacheSource{}

func (c *metaCacheSource) Read(p []byte) (int, error) {
	if c.source == nil {
		s, err := c.sourceFunc()
		if err != nil {
			return 0, err
		}
		c.source = s
	}
	return c.source.Read(p)
}

func (c *metaCacheSource) Close() error {
	if c.source != nil {
		return c.source.Close()
	}
	return nil
}

func (c *metaCacheSource) Meta() (Meta, error) {
	return c.meta, nil
}

type metaStore struct {
	store  map[string]Meta
	storeM sync.Mutex
}

func (c *metaStore) Get(key string, valueFunc func() (Meta, error)) (Meta, error) {
	c.storeM.Lock()
	defer c.storeM.Unlock()

	var err error

	if c.store == nil {
		c.store = map[string]Meta{}
	}

	value, hasKey := c.store[key]
	if !hasKey || value == nil {
		value, err = valueFunc()
		if err != nil {
			return nil, err
		}
		c.store[key] = value
	}

	return value, nil
}
