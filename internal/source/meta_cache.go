package source

type Meta interface {
	Client() string
	Url() string
	Size() int64
	ETag() string
}

type meta struct {
	client string
	url    string
	size   int64
	etag   string
}

var _ Meta = &meta{}

func (m *meta) Client() string {
	return m.client
}

func (m *meta) Url() string {
	return m.url
}

func (m *meta) Size() int64 {
	return m.size
}

func (m *meta) ETag() string {
	return m.etag
}
