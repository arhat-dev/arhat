package connectivityutils

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
)

// nolint:gocyclo
func ReadAfterWrite(r io.Reader, w io.Writer, params map[string]string) ([][]byte, error) {
	var (
		input      []byte
		count      = 1
		readOutput func() ([][]byte, error)
		err        error
	)

	for k, v := range params {
		switch k {
		case "input_text":
			input, err = []byte(v), nil
		case "input_hex":
			input, err = hex.DecodeString(v)
		case "output_count":
			var n int
			n, err = strconv.Atoi(v)
			if err != nil {
				return nil, fmt.Errorf("invalid output_count %q: %w", v, err)
			}
			count = n
		case "output_bytes":
			if readOutput != nil {
				return nil, fmt.Errorf("invalid multiple output spec %q: %q", k, v)
			}

			switch v {
			case "all":
				readOutput = func() ([][]byte, error) {
					data, err2 := ioutil.ReadAll(r)
					if err2 != nil {
						return nil, err2
					}

					return [][]byte{data}, nil
				}
			default:
				n, err2 := strconv.Atoi(v)
				if err2 != nil {
					return nil, fmt.Errorf("incalid output_size %q: %w", v, err2)
				}

				if n == 0 {
					readOutput = func() ([][]byte, error) {
						return nil, nil
					}
				} else {
					readOutput = func() ([][]byte, error) {
						var result [][]byte
						for i := 0; i < count; i++ {
							buf := make([]byte, n)
							_, err = io.ReadFull(r, buf)
							if err != nil {
								return nil, err
							}
							result = append(result, buf)
						}

						return result, nil
					}
				}
			}
		case "output_delim":
			if readOutput != nil {
				return nil, fmt.Errorf("invalid multiple output spec %q: %q", k, v)
			}

			switch v {
			case "line":
				r := bufio.NewReader(r)
				readOutput = func() ([][]byte, error) {
					var result [][]byte

					for i := 0; i < count; i++ {
						var buf []byte
						for data, more, err2 := r.ReadLine(); more; data, more, err = r.ReadLine() {
							if err2 != nil {
								return nil, err2
							}
							buf = append(buf, data...)
						}
						result = append(result, buf)
					}

					return result, nil
				}
			default:
				if len(v) != 1 {
					return nil, fmt.Errorf("invalid output_delim %q", v)
				}

				r := bufio.NewReader(r)
				delim := v[0]
				readOutput = func() ([][]byte, error) {
					var result [][]byte
					for i := 0; i < count; i++ {
						data, err2 := r.ReadBytes(delim)
						if err2 != nil {
							return nil, err2
						}

						buf := make([]byte, len(data))
						_ = copy(buf, data)

						result = append(result, buf)
					}

					return result, nil
				}
			}
		}
	}

	if readOutput == nil {
		return nil, fmt.Errorf("no output collected")
	}

	if err != nil {
		return nil, err
	}

	for size := len(input); size > 0 && err == nil; size = len(input) {
		var n int
		n, err = w.Write(input)

		end := size
		if n < end {
			end = n
		}
		input = input[end:]
	}

	if err != nil {
		return nil, err
	}

	return readOutput()
}
