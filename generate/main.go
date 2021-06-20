package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"text/template"

	"github.com/aligator/checkpoint"
	"github.com/spf13/afero"
	"golang.org/x/tools/go/packages"
)

type match struct {
	fn       *ast.FuncDecl
	pack     *ast.Package
	file     *ast.File
	path     string
	fullPath string
	imports  []Import
}

type Generator struct {
	Out    string
	In     string
	Import string

	found []match
}

func (g *Generator) Clean() error {
	fs := afero.NewOsFs()
	// Generate the plugin file.

	// Check if folder exists. If it already exists, delete it.
	if _, err := fs.Stat(g.Out); err == nil {
		err := fs.RemoveAll(g.Out)
		if err != nil {
			return checkpoint.From(err)
		}
	} else if err != nil && !errors.Is(err, afero.ErrFileNotFound) {
		return checkpoint.From(err)
	}

	// Re-create it.
	err := fs.Mkdir(g.Out, 0777)
	if err != nil {
		return checkpoint.From(err)
	}

	return nil
}

func (g *Generator) Search() error {
	fset := token.NewFileSet()

	err := fs.WalkDir(os.DirFS(g.In), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		fullPath := filepath.Join(g.In, path)

		if d.IsDir() {
			pkgs, err := parser.ParseDir(fset, fullPath, nil, parser.ParseComments)

			if err != nil {
				return err
			}

			for _, p := range pkgs {
				for _, f := range p.Files {
					for _, d := range f.Decls {
						funcDecl, ok := d.(*ast.FuncDecl)
						if !ok || funcDecl.Doc == nil {
							continue
						}

						for _, c := range funcDecl.Doc.List {
							if strings.HasPrefix(c.Text, "//goplug:generate") {
								g.found = append(g.found, match{
									fn:       funcDecl,
									pack:     p,
									file:     f,
									path:     path,
									fullPath: fullPath,
								})
								break
							}
						}
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// Run type checker
	cfg := &packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedModule,
		Dir:  ".",
		Fset: fset,
	}

	for foundI, found := range g.found {
		pkgs, err := packages.Load(cfg, "./"+found.fullPath)

		if err != nil {
			return checkpoint.From(err)
		}
		if packages.PrintErrors(pkgs) > 0 {
			return nil
		}

		// Print the names of the source files
		// for each package listed on the command line.
		for _, pkg := range pkgs {
			// Read the module path (if it is not set explicitly).
			fmt.Println(pkg.Module.Path)
			if g.Import == "" {
				g.Import = filepath.Join(pkg.Module.Path, g.In)
			}

			for _, imp := range pkg.Imports {
				// Find the import in found to get the fake names. They are somehow not loaded with packages.Load...
				fakeName := ""
				for _, foundImp := range found.file.Imports {
					if foundImp.Path.Value != "\""+imp.PkgPath+"\"" {
						continue
					}

					if foundImp.Name != nil {
						fakeName = foundImp.Name.Name
					}
				}

				g.found[foundI].imports = append(g.found[foundI].imports, Import{
					FakeName: fakeName,
					Path:     imp.PkgPath,
					Name:     imp.Name,
				})
			}
		}
	}

	if err != nil {
		log.Fatal(err)
	}

	return nil
}

type Param struct {
	Name       string
	NamePublic string
	Type       string
}

type Action struct {
	Name     string
	Ref      string
	Request  []Param
	Response []Param
}

type Reference struct {
	Name string
	Type string
}

type Import struct {
	FakeName string
	Name     string
	Path     string
}

type PluginData struct {
	Package    string
	Imports    []Import
	References []Reference
	Actions    []Action
}

//go:embed template
var templateFS embed.FS

func (g *Generator) Generate() error {
	data := PluginData{
		Package: "plug",
	}

	var importCounter int32 = 0

	imports := make(map[string]Import)
	// addImport adds the given name or path to the imports.
	// If fileImports is given it is used to resolve it.
	// If nameOrPath is a name the fileImports are mandatory, to resolve it.
	addImport := func(nameOrPath string, fileImports []Import) (string, error) {
		// First get the path behind it if possible.
		var match Import
		if fileImports != nil {
			for _, i := range fileImports {
				if i.FakeName == nameOrPath || i.Name == nameOrPath || strings.HasSuffix(i.Path, nameOrPath) || i.Path == nameOrPath {
					match = i
					// If the file is valid go, there can only be one exact match.
					break
				}
			}
		} else {
			// In this case it must be a path.
			match = Import{
				FakeName: "",
				Name:     filepath.Base(nameOrPath),
				Path:     nameOrPath,
			}
		}

		if match == (Import{}) {
			return "", errors.New("could not find the import")
		}

		// Check if it is already set.
		if i, ok := imports[match.Path]; !ok {
			fakeName := match.Name
			// Make sure it is unique.
			count := atomic.LoadInt32(&importCounter)
			fakeName += strconv.Itoa(int(count))
			atomic.AddInt32(&importCounter, 1)

			// If not, create it.
			imports[match.Path] = Import{
				FakeName: fakeName,
				Name:     match.Name,
				Path:     match.Path,
			}

			return imports[match.Path].FakeName, nil
		} else {
			return i.FakeName, nil
		}
	}

	// Add references
	for i, action := range g.found {
		// Get the receiver.
		if action.fn.Recv == nil || action.fn.Recv.NumFields() < 1 {
			continue
		}

		var rcvTargetName string

		switch v := action.fn.Recv.List[0].Type.(type) {
		case *ast.Ident:
			rcvTargetName = v.Name
		case *ast.StarExpr:
			rcvTargetName = v.X.(*ast.Ident).Name
		default:
			return checkpoint.From(errors.New("receiver type not supported"))
		}

		fmt.Println(rcvTargetName)
		importPath := filepath.Join(g.Import, action.path)
		fakeName, err := addImport(importPath, []Import{
			{
				FakeName: "",
				Name:     action.pack.Name,
				Path:     importPath,
			},
		})
		if err != nil {
			return checkpoint.From(err)
		}

		refName := "ref" + strconv.Itoa(i)
		refType := fakeName + "." + rcvTargetName
		data.References = append(data.References, Reference{
			Name: refName,
			Type: refType,
		})

		// Generate actions.
		actionData := Action{
			Name:    action.fn.Name.Name,
			Ref:     refName,
			Request: []Param{},
			Response: []Param{{
				Name:       "rand",
				NamePublic: "Rand",
				Type:       "int",
			}},
		}

		for _, param := range action.fn.Type.Params.List {
			var paramType string

			switch v := param.Type.(type) {
			case *ast.Ident:
				paramType = v.Name
				if v.Obj != nil {
					// It is a object, not a standard type (like int, string, ...)
					paramType = fakeName + "." + paramType
				}
			case *ast.StarExpr:
				paramType = v.X.(*ast.Ident).Name
			case *ast.SelectorExpr:
				// It is a reference to another type.
				fakeName, err := addImport(v.X.(*ast.Ident).Name, action.imports)
				if err != nil {
					return checkpoint.From(err)
				}
				paramType = fakeName + "." + v.Sel.Name

			default:
				return checkpoint.From(errors.New("param type not supported"))
			}

			actionData.Request = append(actionData.Request, Param{
				Name:       param.Names[0].Name,
				NamePublic: strings.ToUpper(string(param.Names[0].Name[0])) + param.Names[0].Name[1:],
				Type:       paramType,
			})

		}

		data.Actions = append(data.Actions, actionData)
	}

	for _, imp := range imports {
		data.Imports = append(data.Imports, imp)
	}

	t, err := template.ParseFS(templateFS, "template/*")
	if err != nil {
		return checkpoint.From(err)
	}

	fs := afero.NewOsFs()

	f, err := fs.Create(filepath.Join(g.Out, "actions.go"))
	if err != nil {
		return checkpoint.From(err)
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	err = t.ExecuteTemplate(buf, "actions.got", data)
	if err != nil {
		return checkpoint.From(err)
	}
	/*
		fmt.Println(string(buf.Bytes()))

		formatted, err := imports.Process(filepath.Join(g.Out, "actions.go"), buf.Bytes(), &imports.Options{
			Fragment:   true,
			AllErrors:  true,
			Comments:   true,
			TabIndent:  true,
			TabWidth:   8,
			FormatOnly: false,
		})
		if err != nil {
			return checkpoint.From(err)
		}*/

	_, err = f.Write(buf.Bytes())
	if err != nil {
		return checkpoint.From(err)
	}

	return nil
}

func main() {
	out := flag.String("o", "", "")
	importPath := flag.String("i", "", "")

	flag.Parse()
	args := flag.Args()

	g := Generator{
		Out:    *out,
		In:     args[0],
		Import: *importPath,
	}

	err := g.Clean()
	if err != nil {
		panic(err)
	}

	err = g.Search()
	if err != nil {
		panic(err)
	}

	for _, match := range g.found {
		fmt.Println(match.path, match.pack.Name, match.fn.Name)
	}

	err = g.Generate()
	if err != nil {
		panic(err)
	}
}
