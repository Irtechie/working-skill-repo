package main

import (
	"bytes"
	"sync"
)

type cappedBuffer struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	limit    int
	overflow bool
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{limit: limit}
}

func (writer *cappedBuffer) Write(content []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	original := len(content)
	remaining := writer.limit - writer.buffer.Len()
	if remaining <= 0 {
		writer.overflow = true
		return original, nil
	}

	if len(content) > remaining {
		content = content[:remaining]
		writer.overflow = true
	}
	_, _ = writer.buffer.Write(content)
	return original, nil
}

func (writer *cappedBuffer) WriteString(content string) {
	_, _ = writer.Write([]byte(content))
}

func (writer *cappedBuffer) Bytes() []byte {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return append([]byte(nil), writer.buffer.Bytes()...)
}

func (writer *cappedBuffer) String() string {
	return string(writer.Bytes())
}

func (writer *cappedBuffer) Len() int {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.buffer.Len()
}

func (writer *cappedBuffer) Overflowed() bool {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	return writer.overflow
}
