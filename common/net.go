package common

import "io"

const COPY_BUF_SIZE int = 4096

func FastCopy(src io.Reader, dst io.Writer) (int, error) {
	buf := make([]byte, COPY_BUF_SIZE)
	for {
		n, err := src.Read(buf)
		if err != nil {
			return n, err
		}
		n, err = dst.Write(buf[:n])
		if err != nil {
			return n, err
		}
	}
}
