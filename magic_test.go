/*
 * magic_test.go
 *
 * Copyright 2013 Krzysztof Wilczynski
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package magic_test

import (
	"bytes"
	"fmt"
	// XXX(kwilczynski): Not in use at the moment, see comment below ...
	//	"os"
	//	"path"
	"reflect"
	"syscall"
	"testing"

	. "github.com/kwilczynski/go-magic"
)

func CompareStrings(this, other string) bool {
	if this == "" || other == "" {
		return false
	}
	return bytes.Equal([]byte(this), []byte(other))
}

func TestNew(t *testing.T) {
	mgc, err := New()
	if err != nil {
		t.Fatalf("unable to create new Magic type: %s", err.Error())
	}
	defer mgc.Close()

	func(v interface{}) {
		if _, ok := v.(*Magic); !ok {
			t.Fatalf("not a Magic type: %s", reflect.TypeOf(mgc).String())
		}
	}(mgc)
}

func TestMagic_Close(t *testing.T) {
	mgc, _ := New()

	var cookie reflect.Value

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()

	cookie = magic.FieldByName("cookie").Elem()
	if ok := cookie.IsValid(); !ok {
		t.Errorf("value given %v, want %v", ok, true)
	}

	mgc.Close()

	// Should be NULL (at C level) as magic_close() will free underlying Magic database.
	cookie = magic.FieldByName("cookie").Elem()
	if ok := cookie.IsValid(); ok {
		t.Errorf("value given %v, want %v", ok, false)
	}

	// Should be a no-op ...
	mgc.Close()
}

func TestMagic_String(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	magic := reflect.ValueOf(mgc).Elem().FieldByName("magic").Elem()
	path := magic.FieldByName("path")
	cookie := magic.FieldByName("cookie").Elem().Index(0).UnsafeAddr()

	// Get whatever the underlying default path is ...
	paths := make([]string, path.Len())
	for i := 0; i < path.Len(); i++ {
		paths[i] = path.Index(i).String()
	}

	v := fmt.Sprintf("Magic{flags:%d path:%s cookie:0x%x}", 0, paths, cookie)
	if ok := CompareStrings(mgc.String(), v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", mgc.String(), v)
	}
}

func TestMagic_Path(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	v, _ := mgc.Path()
	if len(v) == 0 {
		t.Fatalf("value given \"%T\", should not be empty", v)
	}

	// XXX(krzysztof): Setting "MAGIC" here breaks tests later as it will
	// be persistent between different tests, sadly needed to be disabled
	// for the time being.
	//
	//	p, err := os.Getwd()
	//	if err != nil {
	//		t.Fatal("unable to get current and/or working directory")
	//	}
	//
	//	p = path.Clean(path.Join(p, "fixtures"))
	//	if err = os.Setenv("MAGIC", p); err != nil {
	//		t.Fatalf("unable to set \"MAGIC\" environment variable to \"%s\"", p)
	//	}
	//
	//	v, _ = mgc.Path()
	//	if ok := CompareStrings(v[0], p); !ok {
	//		t.Errorf("value given \"%s\", want \"%s\"", v[0], p)
	//	}

	// TODO(kwilczynski): Test Magic.Load() affecting Magic.Path() as well. But
	// that requires working os.Clearenv() which is yet to be implemented as
	// per http://golang.org/src/pkg/syscall/env_unix.go?s=1772:1787#L101
}

func TestMagic_Flags(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	mgc.SetFlags(MIME)

	flags := MIME_TYPE | MIME_ENCODING
	if v, _ := mgc.Flags(); v != flags {
		t.Errorf("value given 0x%06x, want 0x%06x", v, flags)
	}
}

func TestMagic_SetFlags(t *testing.T) {
	mgc, _ := New()
	defer mgc.Close()

	var err error
	var actual, errno int

	var flagsTests = []struct {
		broken   bool
		errno    int
		expected int
		given    int
	}{
		// Test lower boundary limit.
		{true, 22, 0x000000, -0xffffff},
		// Genuine flags ...
		{false, 0, 0x000000, 0x000000}, // Flag: NONE
		{false, 0, 0x000010, 0x000010}, // Flag: MIME_TYPE
		{false, 0, 0x000400, 0x000400}, // Flag: MIME_ENCODING
		{false, 0, 0x000410, 0x000410}, // Flag: MIME_TYPE | MIME_ENCODING
		// Test upper boundary limit.
		{true, 22, 0x000410, 0xffffff},
	}

	for _, tt := range flagsTests {
		err = mgc.SetFlags(tt.given)
		actual, _ = mgc.Flags()
		if err != nil && tt.broken {
			errno = err.(*MagicError).Errno
			if actual != tt.expected || errno != tt.errno {
				t.Errorf("value given {0x%06x %d}, want {0x%06x %d}",
					actual, errno, tt.expected, tt.errno)
				continue
			}
		}
		if actual != tt.expected {
			t.Errorf("value given 0x%06x, want 0x%06x", actual, tt.expected)
		}
	}

}

func TestMagic_Load(t *testing.T) {
	var mgc *Magic

	var rv bool
	var err error
	var path []string

	mgc, _ = New()

	rv, err = mgc.Load()
	if !rv && err != nil {
		if ok := CompareStrings(err.Error(), ""); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), true, "")
		}
	}

	rv, err = mgc.Load("")
	if rv && err != nil {
		v := "magic: could not find any magic files!"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// XXX(krzysztof): Currently, libmagic API will *never* clear an error once
	// there is one, therefore a whole new session has to be created in order to
	// clear it. Unless upstream fixes this bad design choice, there is nothing
	// to do about it, sadly.
	mgc.Close()

	mgc, _ = New()

	rv, err = mgc.Load("fixtures/png.magic")
	if !rv && err != nil {
		if ok := CompareStrings(err.Error(), ""); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), true, "")
		}
	}

	// Current path should change accordingly ...
	path, _ = mgc.Path()

	v := "fixtures/png.magic"
	if ok := CompareStrings(path[0], v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", path[0], v)
	}

	rv, err = mgc.Load("fixtures/png-broken.magic")
	if rv && err != nil {
		v := "magic: No current entry for continuation"
		if ok := CompareStrings(err.Error(), v); !ok {
			t.Errorf("value given {%v \"%s\"}, want {%v \"%s\"}",
				rv, err.Error(), false, v)
		}
	}

	// Since there was an error, path should remain the same.
	path, _ = mgc.Path()
	if ok := CompareStrings(path[0], v); !ok {
		t.Errorf("value given \"%s\", want \"%s\"", path[0], v)
	}

	mgc.Close()
}

func TestMagic_Compile(t *testing.T) {
}

func TestMagic_Check(t *testing.T) {
}

func TestMagic_File(t *testing.T) {
}

func TestMagic_Buffer(t *testing.T) {
}

func TestMagic_Descriptor(t *testing.T) {
}

func TestOpen(t *testing.T) {
}

func TestCompile(t *testing.T) {
}

func TestCheck(t *testing.T) {
}

func TestVersion(t *testing.T) {
	// XXX(krzysztof): Attempt to circumvent lack of T.Skip() prior to Go version go1.1 ...
	f := reflect.ValueOf(t).MethodByName("Skip")
	if ok := f.IsValid(); !ok {
		f = reflect.ValueOf(t).MethodByName("Log")
	}

	v, err := Version()
	if err != nil && err.(*MagicError).Errno == int(syscall.ENOSYS) {
		f.Call([]reflect.Value{
			reflect.ValueOf("function `int magic_version(void)' is not implemented"),
		})
		return // Should not me reachable on modern Go version.
	}

	if reflect.ValueOf(v).Kind() != reflect.Int || v <= 0 {
		t.Errorf("value given {%v %d}, want {%v > 0}",
			reflect.ValueOf(v).Kind(), v, reflect.Int)
	}
}

func TestFileMime(t *testing.T) {
}

func TestFileEncoding(t *testing.T) {
}

func TestFileType(t *testing.T) {
}

func TestBufferMime(t *testing.T) {
}

func TestBufferEncoding(t *testing.T) {
}

func TestBufferType(t *testing.T) {
}
