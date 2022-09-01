package source

import (
	"crypto/md5"
	"encoding/base64"
	"io"
	"net/url"
	"os"
)

const FileScheme = "file"

const fileClient = "file"

type File struct {
}

var _ Client = &File{}

var ErrFileClient = &Error{msg: "file client error"}

func NewFileClient() *File {
	return &File{}
}

func (c *File) Open(u *url.URL) (Source, error) {
	name := u.Host + u.Path

	fileInfo, err := os.Stat(name)
	if err != nil {
		return nil, ErrFileClient.WithCause(err)
	}

	file, err := os.Open(name)
	if err != nil {
		return nil, ErrFileClient.WithCause(err)
	}

	fs := &fileSource{
		url:      u.String(),
		name:     name,
		file:     file,
		fileInfo: fileInfo,
	}

	md5sum := md5.New()
	hr := newHashReader(md5sum, file)
	fs.readCloser = &readCloser{
		reader: readerFunc(func(p []byte) (n int, err error) {
			n, err = hr.Read(p)
			if err == io.EOF {
				if fs.md5sum == nil {
					fs.md5sum = md5sum.Sum(nil)
				}
			}
			return
		}),
		closer: hr,
	}

	return fs, nil
}

func (c *File) Schemes() []string {
	return []string{
		FileScheme,
	}
}

type fileSource struct {
	url string

	name     string
	file     *os.File
	fileInfo os.FileInfo
	md5sum   []byte

	readCloser io.ReadCloser
}

var _ Source = &fileSource{}

func (s *fileSource) Read(p []byte) (n int, err error) {
	return s.readCloser.Read(p)
}

func (s *fileSource) Close() error {
	return s.readCloser.Close()
}

// Meta returns tags which includes TagMd5.
// Meta must not be called simultaneously with Read.
func (s *fileSource) Meta() (Meta, error) {
	m := &httpMeta{
		meta{
			client: fileClient,
			url:    s.url,
			size:   s.fileInfo.Size(),
		},
	}

	if s.md5sum == nil {
		// Read has not been completed yet
		// Therefore, read the entire file separately to calculate the hash
		file, err := os.Open(s.name)
		if err != nil {
			return nil, ErrFileClient.WithCause(err)
		}

		md5sum, err := hashFromReader(md5.New(), file)
		if err != nil {
			return nil, ErrFileClient.WithCause(err)
		}

		s.md5sum = md5sum
	}

	// Use base64 encoded md5 sum as entity tag
	m.etag = base64.StdEncoding.EncodeToString(s.md5sum)

	return m, nil
}

type fileMeta struct {
	meta
}

var _ Meta = &fileMeta{}
