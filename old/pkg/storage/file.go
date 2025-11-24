package storage

import "io"

type File struct {
	Reader      io.Reader
	Filename    string
	Size        int64
	ContentType string
}

type FileDownload struct {
	Content     io.ReadCloser
	ContentType string
	Size        int64
}

type FileResult struct {
	Key         string
	Size        int64
	ContentType string
	Etag        string
	Url         string
}
