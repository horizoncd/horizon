package log

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

const (
	_info   = "Info"
	_infof  = "Infof"
	_error  = "Error"
	_errorf = "Errorf"
)

func TestWithLogContext(t *testing.T) {
	dir, err := ioutil.TempDir("", "log")
	if err != nil {
		log.Fatal(err)
	}
	infoRegexp, err := regexp.Compile(`I.+log_test.go.+\] \[hello\] log message`)
	if err != nil {
		t.Fatal(err)
	}
	errRegexp, err := regexp.Compile(`E.+log_test.go.+\] \[hello\] log message`)
	if err != nil {
		t.Fatal(err)
	}

	tt := []struct {
		method string
		val    string
		msg    string
		wanted *regexp.Regexp
	}{
		{
			method: _info,
			val:    "hello",
			msg:    "log message",
			wanted: infoRegexp,
		},
		{
			method: _infof,
			val:    "hello",
			msg:    "log message",
			wanted: infoRegexp,
		},
		{
			method: _error,
			val:    "hello",
			msg:    "log message",
			wanted: errRegexp,
		},
		{
			method: _errorf,
			val:    "hello",
			msg:    "log message",
			wanted: errRegexp,
		},
	}

	parent := context.Background()
	for i := range tt {
		os.Stderr, err = os.Create(filepath.Join(dir, tt[i].val))
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tt[i].val, func(t *testing.T) {
			ctx := WithContext(parent, tt[i].val)
			switch tt[i].method {
			case _info:
				Info(ctx, tt[i].msg)
			case _infof:
				Infof(ctx, "%v", tt[i].msg)
			case _error:
				Error(ctx, tt[i].msg)
			case _errorf:
				Errorf(ctx, "%v", tt[i].msg)
			default:
				t.Fatal("not allowed log method")
			}
		})

		Flush()
		output, err := ioutil.ReadFile(filepath.Join(dir, tt[i].val))
		if err != nil {
			t.Fatal(err)
		}
		if !tt[i].wanted.Match(output) {
			t.Fatalf("info output<%v> does match regexp: <%v>", string(output), tt[i].wanted)
		}
	}
}

func TestWithoutLogContext(t *testing.T) {
	dir, err := ioutil.TempDir("", "log")
	if err != nil {
		log.Fatal(err)
	}

	infoRegexp, err := regexp.Compile(`I.+log_test.go.+\] log message`)
	if err != nil {
		t.Fatal(err)
	}
	errRegexp, err := regexp.Compile(`E.+log_test.go.+\] log message`)
	if err != nil {
		t.Fatal(err)
	}
	tt := []struct {
		method string
		msg    string
		wanted *regexp.Regexp
	}{
		{
			method: _info,
			msg:    "log message",
			wanted: infoRegexp,
		},
		{
			method: _infof,
			msg:    "log message",
			wanted: infoRegexp,
		},
		{
			method: _error,
			msg:    "log message",
			wanted: errRegexp,
		},
		{
			method: _errorf,
			msg:    "log message",
			wanted: errRegexp,
		},
	}

	for _, tc := range tt {
		os.Stderr, err = os.Create(filepath.Join(dir, tc.method))
		if err != nil {
			t.Fatal(err)
		}

		t.Run(tc.method, func(t *testing.T) {
			ctx := context.Background()
			switch tc.method {
			case _info:
				Info(ctx, tc.msg)
			case _infof:
				Infof(ctx, "%v", tc.msg)
			case _error:
				Error(ctx, tc.msg)
			case _errorf:
				Errorf(ctx, "%v", tc.msg)
			default:
				t.Fatal("not allowed log method")
			}
		})

		Flush()
		output, err := ioutil.ReadFile(filepath.Join(dir, tc.method))
		if err != nil {
			t.Fatal(err)
		}
		if !tc.wanted.Match(output) {
			t.Fatalf("info output<%v> does match regexp: <%v>", string(output), tc.wanted)
		}
	}
}

func TestRegexp(t *testing.T) {
	r, err := regexp.Compile(`I.+\] \[hello\] log message`)
	if err != nil {
		t.Fatal(err)
	}
	if !r.MatchString("I0914 13:38:11.128646   46433 log.go:23] [hello] log message\n") {
		t.Fatal("bad regexp")
	}
}
