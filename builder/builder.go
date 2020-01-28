package builder

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"Gtkmmcargo/shared"
	"Gtkmmcargo/tr"
)

var (
	gtkmmCompilerFlags []string
	gtkmmLinkerFlags   []string
)

type Builder struct {
	projectDir string // Path of root project directory (must be finished with /)
	workingDir string
	files      []string
	objects    []string
	extObjects []string
}

func init() {
	fetchGtkmmCompilerFlags()
	fetchGtkmmLinkerFlags()
}

func New(projectDir string) *Builder {
	if gtkmmCompilerFlags != nil {
		if gtkmmLinkerFlags != nil {
			workingDir := filepath.Join(projectDir, ".gtkmmcargo")
			if shared.CreateDirIfNeeded(workingDir) {
				return &Builder{projectDir: projectDir, workingDir: workingDir}
			}
		}
	}
	return nil
}

func (b *Builder) AddFile(fname string) {
	b.files = append(b.files, filepath.Join(b.projectDir, fname))
}

func (b *Builder) PrintFilesToCompile() {
	print("files to compile", b.files)
}

func (b *Builder) PrintGtkmmFlags() {
	display("gtkmm builder flags", gtkmmCompilerFlags)
	display("gtkmm linker flags", gtkmmLinkerFlags)

}

// g++ test.o -o test `pkg-config gtkmm-3.0 --libs`
func (b *Builder) Link(binName string) (bool, string, string) {
	var outBuffer, errBuffer bytes.Buffer

	binPath := filepath.Join(b.projectDir, binName)

	var params []string
	params = append(params, b.objects...)
	params = append(params, "-o", binPath)
	params = append(params, gtkmmLinkerFlags...)
	cmd := exec.Command("g++", params...)
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer
	if err := cmd.Run(); tr.IsOK(err) {
		return true, outBuffer.String(), errBuffer.String()
	}
	return false, outBuffer.String(), errBuffer.String()
}
func (b *Builder) Compile() bool {
	for _, src := range b.files {
		if _, name := shared.PathComponents(src); name != "" {
			if base, ext := shared.NameComponent(name); validExt(ext) {
				dst := b.workingDir + string(os.PathSeparator) + base + ".o"
				ok := compileFile(src, dst)
				if !ok {
					return false
				}
				b.objects = append(b.objects, dst)
			}
		}
	}
	return true
}

// g++ -c test.cc -o test.o `pkg-config gtkmm-3.0 --cflags`
func compileFile(src, dst string) bool {
	var errBuffer bytes.Buffer
	params := []string{"-c", src, "-o", dst}
	params = append(params, gtkmmCompilerFlags...)
	cmd := exec.Command("g++", params...)
	cmd.Stderr = &errBuffer
	if err := cmd.Run(); tr.IsOK(err) {
		return true
	}
	fmt.Println(errBuffer.String())
	return false
}

func fetchGtkmmCompilerFlags() {
	var outBuffer bytes.Buffer
	cmd := exec.Command("pkg-config", "gtkmm-3.0", "glib-2.0", "--cflags")
	cmd.Stdout = &outBuffer
	err := cmd.Run()
	if err != nil {
		log.Println(err)
		return
	}
	result := strings.TrimSpace(outBuffer.String())
	gtkmmCompilerFlags = strings.Split(result, " ")
}

func fetchGtkmmLinkerFlags() {
	var outBuffer bytes.Buffer
	cmd := exec.Command("pkg-config", "gtkmm-3.0", "--libs")
	cmd.Stdout = &outBuffer
	err := cmd.Run()
	if err != nil {
		log.Println(err)
	}
	result := strings.TrimSpace(outBuffer.String())
	gtkmmLinkerFlags = strings.Split(result, " ")
}

func display(name string, values []string) {
	fmt.Println(name + ":")
	for _, f := range values {
		fmt.Printf("\t%s\n", f)
	}
	fmt.Println()
}

func validExt(ext string) bool {
	return ext == "cpp" || ext == "cc" || ext == "cxx" || ext == "c"
}
