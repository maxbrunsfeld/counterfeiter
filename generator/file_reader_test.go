package generator_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"strings"
	"testing"

	"github.com/maxbrunsfeld/counterfeiter/v6/generator"
)

func TestFileReader(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		readerCreator func(generator.Opener) generator.FileReader
		open          generator.Opener

		workingDir string
		path       string

		expectedErrMsg  string
		expectedContent string
		expectedCalls   []string
	}{
		// SimpleFileReader
		"[simple] when the filepath is empty, it's a noop": {
			readerCreator: simpleReaderCreator,
		},
		"[simple] when open returns an error, the error bubbles up": {
			readerCreator:  simpleReaderCreator,
			open:           openReturningErr("some error"),
			path:           relFile,
			expectedErrMsg: "some error",
			expectedCalls:  []string{relFile, relFile},
		},
		"[simple] when open returns a reader, the readers content is read": {
			readerCreator:   simpleReaderCreator,
			open:            openReturningReader("some content 0"),
			path:            relFile,
			expectedContent: "some content 0",
			expectedCalls:   []string{relFile, relFile},
		},
		"[simple] when the working directory is set but the filepath is absolut, the absolute path is used": {
			readerCreator:   simpleReaderCreator,
			open:            openReturningReader("some content 1"),
			workingDir:      workingDir,
			path:            absFile,
			expectedContent: "some content 1",
			expectedCalls:   []string{absFile, absFile},
		},
		"[simple] when the working directory and a relative filepath is set, the paths are combined & cleaned": {
			readerCreator:   simpleReaderCreator,
			open:            openReturningReader("some content 2"),
			workingDir:      workingDir,
			path:            relFileUp,
			expectedContent: "some content 2",
			expectedCalls:   []string{expectedFileUp, expectedFileUp},
		},
		"[simple] when the reader's Read returns an error, the error bubbles up": {
			readerCreator:  simpleReaderCreator,
			open:           openReturningFailingReader("some read error"),
			path:           relFile,
			expectedErrMsg: "some read error",
			expectedCalls:  []string{relFile, relFile},
		},

		// CachedFileReader
		"[cached] when the filepath is empty, it's a noop": {
			readerCreator: cachedReaderCreator,
		},
		"[cached] when open returns an error, the error bubbles up": {
			readerCreator:  cachedReaderCreator,
			open:           openReturningErr("some error"),
			path:           relFile,
			expectedErrMsg: "some error",
			expectedCalls:  []string{relFile, relFile}, // because on error, we don't cache
		},
		"[cached] when open returns a reader, the readers content is read": {
			readerCreator:   cachedReaderCreator,
			open:            openReturningReader("some content 3"),
			path:            relFile,
			expectedContent: "some content 3",
			expectedCalls:   []string{relFile},
		},
		"[cached] when the working directory is set but the filepath is absolut, the absolute path is used": {
			readerCreator:   cachedReaderCreator,
			open:            openReturningReader("some content 4"),
			workingDir:      workingDir,
			path:            absFile,
			expectedContent: "some content 4",
			expectedCalls:   []string{absFile},
		},
		"[cached] when the working directory and a relative filepath is set, the paths are combined & cleaned": {
			readerCreator:   cachedReaderCreator,
			open:            openReturningReader("some content 5"),
			workingDir:      workingDir,
			path:            relFileUp,
			expectedContent: "some content 5",
			expectedCalls:   []string{expectedFileUp},
		},
		"[cached] when the reader's Read returns an error, the error bubbles up": {
			readerCreator:  cachedReaderCreator,
			open:           openReturningFailingReader("some read error"),
			path:           relFile,
			expectedErrMsg: "some read error",
			expectedCalls:  []string{relFile, relFile}, // because on error, we don't cache
		},
	}

	for name, tc := range tests {
		tc := tc

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			spy := &openSpy{Func: tc.open}
			reader := tc.readerCreator(spy.Open)

			readAndCheckContent := func() {
				content, err := reader.Get(tc.workingDir, tc.path)
				checkErr(t, err, tc.expectedErrMsg)

				if a, e := content, tc.expectedContent; e != a {
					t.Errorf("Expected content to be '%s', got '%s'", e, a)
				}
			}

			// let's run the tests twice, to check on caching
			readAndCheckContent()
			readAndCheckContent()

			if a, e := spy.Calls, tc.expectedCalls; !reflect.DeepEqual(e, a) {
				t.Errorf("Expected open call args to be %#v, got %#v", e, a)
			}
		})
	}
}

func simpleReaderCreator(o generator.Opener) generator.FileReader {
	return &generator.SimpleFileReader{Open: o}
}
func cachedReaderCreator(o generator.Opener) generator.FileReader {
	return &generator.CachedFileReader{Open: o}
}

func openReturningErr(err string) generator.Opener {
	return func(_ string) (io.ReadCloser, error) {
		return nil, fmt.Errorf(err)
	}
}
func openReturningReader(content string) generator.Opener {
	return func(_ string) (io.ReadCloser, error) {
		return ioutil.NopCloser(strings.NewReader(content)), nil
	}
}
func openReturningFailingReader(err string) generator.Opener {
	return func(_ string) (io.ReadCloser, error) {
		r := &erroringReader{
			reader: ioutil.NopCloser(strings.NewReader("some random file content")),
			err:    fmt.Errorf(err),
		}
		return r, nil
	}
}

func checkErr(t *testing.T, err error, msg string) {
	t.Helper()

	if msg == "" {
		if err != nil {
			t.Errorf("Expected no error to occur, got %v", err)
		}
		return
	}

	if err == nil {
		t.Errorf("Expected error '%s', got no error", msg)
		return
	}

	if a, e := err.Error(), msg; a != e {
		t.Errorf("Expected error '%s', got: '%s'", e, a)
	}
}

type openSpy struct {
	Func  generator.Opener
	Calls []string
}

func (o *openSpy) Open(p string) (io.ReadCloser, error) {
	o.Calls = append(o.Calls, p)
	return o.Func(p)
}

var _ generator.Opener = (&openSpy{}).Open

type erroringReader struct {
	reader    io.ReadCloser
	callCount int
	err       error
}

func (r *erroringReader) Read(p []byte) (int, error) {
	r.callCount++
	if r.callCount >= 2 {
		return 0, r.err
	}
	return r.reader.Read(p)
}

func (r *erroringReader) Close() error {
	return r.reader.Close()
}

var _ io.ReadCloser = &erroringReader{}
