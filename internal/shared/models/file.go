package models

import "io"

type InternalFile struct {
	Reader      io.Reader
	Filename    string
	Size        int64
	ContentType string
}
