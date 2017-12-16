// +build linux

package gopjit

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"strings"
)

// JITCompiler is jit compiler for golang using plugin.
type JITCompiler struct {
	buildDir string
}

// NewJIT creates JITCompiler.
func NewJIT() JITCompiler {
	parent := os.Getenv("GOPJITBUILDDIR")
	if parent == "" {
		parent = os.TempDir()
	}
	dir, err := ioutil.TempDir(parent, "gopjit")
	if err != nil {
		panic(err)
	}
	return JITCompiler{buildDir: dir}
}

// Build builds ast.Node to go plugin and load function named F0 from it.
func (jit *JITCompiler) Build(n *ast.Node) (interface{}, error) {
	var b bytes.Buffer
	if err := format.Node(&b, nil, n); err != nil {
		return nil, err
	}

	return jit.BuildSrc(b.String())
}

func makeTempDir(dir string) error {
	f, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		return nil
	}
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return fmt.Errorf("%s is not directory", dir)
	}
	return nil
}

func saveToFile(dir, src string) (p string, err error) {
	f, err := ioutil.TempFile(dir, "goplugin-")
	if err != nil {
		return "", err
	}
	defer func() {
		if cerr := f.Close(); cerr != nil {
			if err != nil {
				err = fmt.Errorf("multiple error has occurred: %s and %s", err, cerr)
			} else {
				err = cerr
			}
			return
		}
		name := f.Name()
		p = name + ".go"
		err = os.Rename(name, p)
	}()

	_, err = fmt.Fprint(f, src)
	return
}

// BuildSrc compile src to go plugin and load function named F0 from it.
func (jit *JITCompiler) BuildSrc(src string) (interface{}, error) {
	if err := makeTempDir(jit.buildDir); err != nil {
		return nil, err
	}

	p, err := saveToFile(jit.buildDir, src)
	if err != nil {
		return nil, err
	}

	name := filepath.Base(p)
	cmd := exec.Command("go", "build", "-buildmode=plugin", name)
	cmd.Env = os.Environ()
	cmd.Dir = jit.buildDir
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	so, err := plugin.Open(strings.TrimSuffix(p, filepath.Ext(p)) + ".so")
	if err != nil {
		return nil, err
	}

	return so.Lookup("F0")
}
