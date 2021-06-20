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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/aligator/checkpoint"
	"github.com/spf13/afero"
	"golang.org/x/tools/imports"
)

type match struct {
	fn   *ast.FuncDecl
	pack *ast.Package
	file *ast.File
	path string
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
	return fs.WalkDir(os.DirFS(g.In), ".", func(path string, d fs.DirEntry, err error) error {
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
									fn:   funcDecl,
									pack: p,
									file: f,
									path: path,
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
	Name string
	Path string
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

		refName := "ref" + strconv.Itoa(i)
		// TODO: resolve package name collision (maybe just use a name for its import) (see also bellow)
		refType := action.pack.Name + "." + rcvTargetName
		data.References = append(data.References, Reference{
			Name: refName,
			Type: refType,
		})

		// TODO: filter double imports
		data.Imports = append(data.Imports, Import{
			Name: "", // Todo: name it with a definitely unique name to be sure.
			Path: filepath.Join(g.Import, action.path),
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
					// TODO: resolve package name collision (maybe just use a name for its import) (see also above)
					paramType = action.pack.Name + "." + paramType
				}
			case *ast.StarExpr:
				paramType = v.X.(*ast.Ident).Name
			case *ast.SelectorExpr:
				// It is a reference to another type.
				paramType = v.X.(*ast.Ident).Name + "." + v.Sel.Name
				// ToDo: Try to find out the import used for that if it is not in the same package.
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
	}

	_, err = f.Write(formatted)
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
