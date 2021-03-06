package common

import (
	"io"
)

// CombinedReadWriter just combines a Reader with a writer and
// therefore implements the ReadWriteCloser interface.
// All actions are just passed to the respective implementation.
type CombinedReadWriter struct {
	In  io.ReadCloser
	Out io.WriteCloser
}

func (s CombinedReadWriter) Read(p []byte) (n int, err error) {
	return s.In.Read(p)
}

func (s CombinedReadWriter) Write(p []byte) (n int, err error) {
	return s.Out.Write(p)
}

func (s CombinedReadWriter) Close() error {
	err := s.Out.Close()
	if err != nil {
		return err
	}
	return s.In.Close()
}
