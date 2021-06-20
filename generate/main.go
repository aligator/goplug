package main

import (
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
)

type match struct {
	fn   *ast.FuncDecl
	pack *ast.Package
	path string
}

type Generator struct {
	Out    string
	In     string
	Import string

	found []match
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

type PluginData struct {
	Package    string
	Imports    []string
	References []Reference
	Actions    []Action
}

//go:embed template
var templateFS embed.FS

func (g *Generator) Generate() error {
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

	data := PluginData{
		Package: "plugin",
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
		refType := action.pack.Name + "." + rcvTargetName
		data.References = append(data.References, Reference{
			Name: refName,
			Type: refType,
		})

		data.Imports = append(data.Imports, filepath.Join(g.Import, action.path))

		data.Actions = append(data.Actions, Action{
			Name: action.fn.Name.Name,
			Ref:  refName,
			Request: []Param{{
				Name:       "n",
				NamePublic: "N",
				Type:       "int",
			}},
			Response: []Param{{
				Name:       "rand",
				NamePublic: "Rand",
				Type:       "int",
			}},
		})
	}

	t, err := template.ParseFS(templateFS, "template/*")
	if err != nil {
		return checkpoint.From(err)
	}

	err = t.ExecuteTemplate(os.Stdout, "actions.got", data)
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

	err := g.Search()
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
