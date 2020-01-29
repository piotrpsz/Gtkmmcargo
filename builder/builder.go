package builder

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"Gtkmmcargo/shared"
)

var (
	gtkmmCompilerFlags []string
	gtkmmLinkerFlags   []string
)

type Builder struct {
	projectDir string // Path of root project directory (must be finished with /)
	workingDir string

	files []string

	customCompileFlags []string // custom compile flags
	customLinkFlags    []string // custom link flags (to link additional libraries, e.g. SQLite)
	extObjects         []string // external object files for linking
	objects            []string // objects generated during compilation
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

func (b *Builder) Build(binName string) bool {
	t := time.Now()
	if b.compile() {
		binPath := filepath.Join(b.projectDir, binName)
		if b.link(binPath) {
			elapsed := time.Since(t).Seconds()
			fmt.Printf("OK. Duration: %v sec.\n", elapsed)
			return true
		}
	}
	return false
}

func (b *Builder) link(binPath string) bool {
	var (
		errBuffer bytes.Buffer
		params    []string
	)
	params = append(params, b.objects...)
	params = append(params, b.extObjects...)
	params = append(params, "-o", binPath)
	params = append(params, gtkmmLinkerFlags...)
	params = append(params, b.customLinkFlags...)
	cmd := exec.Command("g++", params...)
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return false
	}

	output := strings.TrimSpace(errBuffer.String())
	if len(output) > 0 {
		fmt.Println(output)
		return false
	}
	return true
}

func (b *Builder) compile() bool {
	for _, src := range b.files {
		if _, name := shared.PathComponents(src); name != "" {
			if base, ext := shared.NameComponent(name); validExt(ext) {
				dst := b.workingDir + string(os.PathSeparator) + base + ".o"
				ok := b.compileFile(src, dst)
				if !ok {
					return false
				}
				b.objects = append(b.objects, dst)
			}
		}
	}
	return true
}

func (b *Builder) compileFile(src, dst string) bool {
	var errBuffer bytes.Buffer

	params := []string{"-c", src, "-o", dst}
	params = append(params, gtkmmCompilerFlags...)
	params = append(params, b.customCompileFlags...)
	cmd := exec.Command("g++", params...)
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return false
	}

	output := strings.TrimSpace(errBuffer.String())
	if len(output) > 0 {
		fmt.Println(output)
		return false
	}
	return true
}

func fetchGtkmmCompilerFlags() {
	var outBuffer bytes.Buffer
	cmd := exec.Command("pkg-config", "gtkmm-3.0", "glib-2.0", "--cflags")
	cmd.Stdout = &outBuffer

	if err := cmd.Run(); err != nil {
		log.Println(err)
		return
	}

	if result := strings.TrimSpace(outBuffer.String()); result != "" {
		gtkmmCompilerFlags = strings.Split(result, " ")
	}
}

func fetchGtkmmLinkerFlags() {
	var outBuffer bytes.Buffer
	cmd := exec.Command("pkg-config", "gtkmm-3.0", "--libs")
	cmd.Stdout = &outBuffer

	if err := cmd.Run(); err != nil {
		log.Println(err)
		return
	}

	if result := strings.TrimSpace(outBuffer.String()); result != "" {
		gtkmmLinkerFlags = strings.Split(result, " ")
	}
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
