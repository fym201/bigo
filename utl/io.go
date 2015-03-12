package utl

import (
	"io"
)

func Pipe(in io.Reader, out io.Writer) error {
	buf := make([]byte, 2048)
	for {
		n, err := in.Read(buf)

		if n > 0 {
			_, err = out.Write(buf)
		}
		if err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
	}
	return nil
}

func PipeAndClose(in io.ReadCloser, out io.WriteCloser) error {
	defer func() {
		in.Close()
		out.Close()
	}()
	return Pipe(in, out)
}
