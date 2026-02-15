package filter

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
)

type Stats struct {
	TotalLines    int
	FilteredLines int
}

func InsertFilter(r io.Reader, w io.Writer, skipTables []string, maxLineBytes int) (Stats, error) {
	patterns := make([]*regexp.Regexp, 0, len(skipTables))
	for _, pat := range skipTables {
		re, err := regexp.Compile(pat)
		if err != nil {
			return Stats{}, fmt.Errorf("invalid pattern %q: %w", pat, err)
		}
		patterns = append(patterns, re)
	}

	reInsert := regexp.MustCompile(`^INSERT INTO ` + "`?" + `([^` + "`" + ` ]*)` + "`?")
	reader := bufio.NewReaderSize(r, 64*1024)
	writer := bufio.NewWriterSize(w, 64*1024)
	defer writer.Flush()

	var stats Stats
	insideInsertBlock := false

	for {
		line, err := readLine(reader, maxLineBytes)
		if err == io.EOF {
			break
		}
		if err != nil {
			return stats, err
		}

		stats.TotalLines++
		trimmed := bytes.TrimSuffix(line, []byte("\n"))
		trimmed = bytes.TrimSuffix(trimmed, []byte("\r"))

		if matches := reInsert.FindSubmatch(trimmed); matches != nil {
			tableName := string(matches[1])
			skipThis := false
			for _, re := range patterns {
				if re.MatchString(tableName) {
					skipThis = true
					break
				}
			}
			if skipThis {
				stats.FilteredLines++
				insideInsertBlock = !bytes.HasSuffix(trimmed, []byte(";"))
				continue
			}
		}

		if insideInsertBlock {
			stats.FilteredLines++
			if bytes.HasSuffix(trimmed, []byte(";")) {
				insideInsertBlock = false
			}
			continue
		}

		if _, err := writer.Write(line); err != nil {
			return stats, fmt.Errorf("write output: %w", err)
		}
	}

	return stats, nil
}

func readLine(r *bufio.Reader, maxLineBytes int) ([]byte, error) {
	var buf []byte
	for {
		chunk, err := r.ReadSlice('\n')
		buf = append(buf, chunk...)
		if len(buf) > maxLineBytes {
			return nil, fmt.Errorf("line exceeds MAX_LINE_BYTES=%d", maxLineBytes)
		}
		if err == nil {
			return buf, nil
		}
		if err == bufio.ErrBufferFull {
			continue
		}
		if err == io.EOF {
			if len(buf) == 0 {
				return nil, io.EOF
			}
			return buf, nil
		}
		return nil, fmt.Errorf("read line: %w", err)
	}
}
