package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"Gtkmmcargo/shared"
	"Gtkmmcargo/tr"
)

const (
	DefaultConfigFileName   = "gtkmmcargo.cfg"
	DefaultWorkingDirectory = ".gtkmmcargo"
)

var (
	gtkmmCompilerFlags []string
	gtkmmLinkerFlags   []string
)

type Builder struct {
	ProjectDirectory   string   `json:"project_directory"`
	WorkingDirectory   string   `json:"working_directory"`
	ExecutableName     string   `json:"executable_name"`
	SourceFiles        []string `json:"source_files"`
	CustomCompileFlags []string `json:"custom_compile_flags"`
	CustomLinkFlags    []string `json:"custom_link_flags"`
	ExternalObjects    []string `json:"external_object_files"`
	objects            []string
	cfgFilePath        string
	objMutex           sync.Mutex
}

func init() {
	fetchGtkmmCompilerFlags()
	fetchGtkmmLinkerFlags()
}

func New(cfgFilePath string) *Builder {
	if cfgFilePath == "" {
		cfgFilePath = DefaultConfigFileName
	}
	if data := readConfig(cfgFilePath); len(data) > 0 {
		if b := parseConfigData(data); b != nil {
			if b.WorkingDirectory == "" {
				b.WorkingDirectory = filepath.Join(b.ProjectDirectory, DefaultWorkingDirectory)
			} else {
				// absolute path
				if b.WorkingDirectory[0] == '/' {
					// it is OK
				} else {
					b.WorkingDirectory = filepath.Join(b.ProjectDirectory, b.WorkingDirectory)
				}
			}
			shared.CreateDirIfNeeded(b.WorkingDirectory)

			return b
		}
	}
	return nil
}

func NewEmpty() *Builder {
	return &Builder{
		ProjectDirectory:   "",
		WorkingDirectory:   "",
		ExecutableName:     "",
		SourceFiles:        []string{},
		CustomCompileFlags: []string{"-Wall", "-std=c++17", "-O3"},
		CustomLinkFlags:    []string{},
		ExternalObjects:    []string{},
	}
}

func readConfig(fpath string) []byte {
	if handle := shared.OpenFile(fpath); handle != nil {
		defer handle.Close()

		if data := shared.ReadFileContent(handle); data != nil {
			return data
		}
	}
	return nil
}

func parseConfigData(data []byte) *Builder {
	var b Builder

	if err := json.Unmarshal(data, &b); tr.IsOK(err) {
		return &b
	}
	return nil
}

func (b *Builder) Save() bool {
	if data, err := json.MarshalIndent(b, "", "   "); tr.IsOK(err) {
		data = append(data, '\n')
		if shared.OverwriteFileContent(b.cfgFilePath, data) {
			return true
		}
	}
	return false
}

func (b *Builder) PrintFilesToCompile() {
	print("files to compile", b.SourceFiles)
}

func PrintGtkmmFlags() {
	display("gtkmm builder flags", gtkmmCompilerFlags)
	display("gtkmm linker flags", gtkmmLinkerFlags)
}

func (b *Builder) Build() bool {
	t := time.Now()
	if b.compile() {
		binPath := filepath.Join(b.ProjectDirectory, b.ExecutableName)
		if b.link(binPath) {
			elapsed := time.Since(t).Seconds()
			fmt.Printf("OK. Duration: %v sec.\n", elapsed)
			return true
		} else {
			fmt.Println("Linking failure")
		}
	} else {
		fmt.Println("Compilation failure")
	}
	return false
}

func (b *Builder) link(binPath string) bool {
	var (
		outBuffer, errBuffer bytes.Buffer
		params               []string
	)
	params = append(params, b.objects...)
	params = append(params, b.ExternalObjects...)
	params = append(params, "-o", binPath)
	params = append(params, gtkmmLinkerFlags...)
	params = append(params, b.CustomLinkFlags...)
	cmd := exec.Command("g++", params...)
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return false
	}

	stdOutput := strings.TrimSpace(outBuffer.String())
	if len(stdOutput) > 0 {
		fmt.Println(stdOutput)
	}

	errOutput := strings.TrimSpace(errBuffer.String())
	if len(errOutput) > 0 {
		fmt.Println(errOutput)
		return false
	}
	return true
}

func (b *Builder) compile() bool {
	var wg sync.WaitGroup

	for _, src := range b.SourceFiles {
		if _, name := shared.PathComponents(src); name != "" {
			if base, ext := shared.NameComponent(name); validExt(ext) {
				dst := b.WorkingDirectory + string(os.PathSeparator) + base + ".o"
				wg.Add(1)
				go b.compileFile(&wg, src, dst)
			}
		}
	}

	wg.Wait()
	return len(b.SourceFiles) == len(b.objects)
}

func (b *Builder) compileFile(wg *sync.WaitGroup, src, dst string) {
	defer wg.Done()

	var outBuffer, errBuffer bytes.Buffer

	src = filepath.Join(b.ProjectDirectory, src)
	if !shared.ExistsFile(src) {
		fmt.Printf("File not exists (%s)\n", src)
		return
	}

	params := []string{"-c", src, "-o", dst}
	params = append(params, b.CustomCompileFlags...)
	params = append(params, gtkmmCompilerFlags...)
	//fmt.Println(params)
	cmd := exec.Command("g++", params...)
	cmd.Stdout = &outBuffer
	cmd.Stderr = &errBuffer

	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}

	stdOutput := strings.TrimSpace(outBuffer.String())
	if len(stdOutput) > 0 {
		fmt.Println(stdOutput)
	}

	errOutput := strings.TrimSpace(errBuffer.String())
	if len(errOutput) > 0 {
		fmt.Println(errOutput)
		return
	}

	b.objMutex.Lock()
	defer b.objMutex.Unlock()
	b.objects = append(b.objects, dst)
	return
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
