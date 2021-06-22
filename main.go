package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"github.com/aligator/checkpoint"
	"github.com/spf13/afero"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/packages"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"text/template"
)

type match struct {
	fn       *ast.FuncDecl
	comment  string
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
	Pack   string
	FS     afero.Fs

	found     []match
	generated *PluginData
}

// CleanDestination removes the whole destination folder (if it exists)
// and re-creates a new empty folder.
func (g *Generator) CleanDestination() error {
	// Check if folder exists. If it already exists, delete it.
	if _, err := g.FS.Stat(g.Out); err == nil {
		err := g.FS.RemoveAll(g.Out)
		if err != nil {
			return checkpoint.From(err)
		}
	} else if !errors.Is(err, afero.ErrFileNotFound) {
		return checkpoint.From(err)
	}

	// Re-create it.
	err := g.FS.Mkdir(g.Out, 0777)
	if err != nil {
		return checkpoint.From(err)
	}

	return nil
}

// Search for the '//goplug:generate annotation' in the code and try to get all
// needed information about the methods annotated by them.
func (g *Generator) Search() error {
	fset := token.NewFileSet()

	absIn, err := filepath.Abs(g.In)
	if err != nil {
		return checkpoint.From(err)
	}

	// Walk all dirs of the input project
	err = afero.Walk(g.FS, absIn, func(fullPath string, info fs.FileInfo, err error) error {
		if err != nil {
			return checkpoint.From(err)
		}

		if !info.IsDir() {
			return nil
		}

		packagePath := filepath.Base(fullPath)

		// Parse the dir as package
		pkgs, err := parser.ParseDir(fset, fullPath, nil, parser.ParseComments)
		if err != nil {
			return checkpoint.From(err)
		}

		// Run through everything and find the prefix.
		for _, p := range pkgs {
			for _, f := range p.Files {
				for _, d := range f.Decls {
					// Only functions are interesting.
					funcDecl, ok := d.(*ast.FuncDecl)
					if !ok || funcDecl.Doc == nil {
						continue
					}

					for _, c := range funcDecl.Doc.List {
						if strings.HasPrefix(c.Text, "//goplug:generate") {
							// Re-generate the comment
							comment := ""
							for _, c := range funcDecl.Doc.List {
								// Ignore all lines which do not start with "// "
								// as these are normally comment directives like
								// "//line" and "//go:noinline".
								// This also prevents adding "//goplug:generate".
								if !strings.HasPrefix(c.Text, "// ") {
									continue
								}

								if comment != "" {
									comment += "\n"
								}

								comment += c.Text
							}

							g.found = append(g.found, match{
								fn:       funcDecl,
								comment:  comment,
								pack:     p,
								file:     f,
								path:     packagePath,
								fullPath: fullPath,
							})
							break
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

	// Run type checker because otherwise I could not resolve imports needed
	// by params / result values used by the annotated functions.
	// Maybe there is a way to combine the logic above with the type checker,
	// but for now it works.
	cfg := &packages.Config{
		Mode: packages.NeedImports | packages.NeedTypes | packages.NeedSyntax | packages.NeedName | packages.NeedModule,
		Dir:  absIn,
		Fset: fset,
	}

	for foundI, found := range g.found {
		pkgs, err := packages.Load(cfg, found.fullPath)
		if err != nil {
			return checkpoint.From(err)
		}
		if packages.PrintErrors(pkgs) > 0 {
			return nil
		}

		for _, pkg := range pkgs {
			// Read the module path (if it is not set explicitly).
			// That is needed to compose the import paths correctly.
			// This is only done once.
			if g.Import == "" {
				g.Import = pkg.Module.Path
			}

			for _, imp := range pkg.Imports {
				// Find the import in found to get the fake names.
				// (e.g. named imports)
				// They are somehow not loaded with packages.Load...
				fakeName := ""
				for _, foundImp := range found.file.Imports {
					if foundImp.Path.Value != "\""+imp.PkgPath+"\"" {
						continue
					}

					if foundImp.Name != nil {
						fakeName = foundImp.Name.Name
					}
				}

				// Add the information learned.
				g.found[foundI].imports = append(g.found[foundI].imports, Import{
					FakeName: fakeName,
					Path:     imp.PkgPath,
					Name:     imp.Name,
				})
			}
		}
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
	Comment  string
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

// Generate the plugin actions file
func (g *Generator) Generate() error {
	g.generated = &PluginData{
		Package: g.Pack,
	}

	// importCounter just counts the imports added to generate in any case
	// unique imports by just appending this number.
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
	for _, action := range g.found {
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

		refName := strings.ToUpper(string(fakeName[0])) + fakeName[1:] + rcvTargetName + "Ref"
		refType := fakeName + "." + rcvTargetName

		found := false
		for _, ref := range g.generated.References {
			if ref.Name == refName {
				found = true
				break
			}
		}

		if !found {
			g.generated.References = append(g.generated.References, Reference{
				Name: refName,
				Type: refType,
			})
		}

		// Generate actions.
		actionData := Action{
			Name:    action.fn.Name.Name,
			Comment: action.comment,
			Ref:     refName,
		}

		// Add parameters.
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
				// Todo: this part may need rework...
				switch target := v.X.(type) {
				case *ast.Ident:
					paramType = target.Name
					if target.Obj != nil {
						// It is a object, not a standard type (like int, string, ...)
						paramType = fakeName + "." + paramType
					}
					// It is a reference to another type.
					paramType = "*" + paramType
				case *ast.SelectorExpr:
					// It is a reference to another type from another package.
					fakeName, err := addImport(target.X.(*ast.Ident).Name, action.imports)
					if err != nil {
						return checkpoint.From(err)
					}
					paramType = "*" + fakeName + "." + target.Sel.Name
				}

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

		// Add response.
		// TODO: remove code duplication with for above
		for i, res := range action.fn.Type.Results.List {
			if i == len(action.fn.Type.Results.List)-1 {
				// As the last return type has to be an error and that
				// is handled separately, ignore it.
				// TODO: Add check if the last is really an error
				break
			}

			var resType string

			switch v := res.Type.(type) {
			case *ast.Ident:
				resType = v.Name
				if v.Obj != nil {
					// It is a object, not a standard type (like int, string, ...)
					resType = fakeName + "." + resType
				}
			case *ast.StarExpr:
				switch target := v.X.(type) {
				case *ast.Ident:
					resType = target.Name
					if target.Obj != nil {
						// It is a object, not a standard type (like int, string, ...)
						resType = fakeName + "." + resType
					}
					// It is a reference to another type.
					resType = "*" + resType
				case *ast.SelectorExpr:
					// It is a reference to another type from another package.
					fakeName, err := addImport(target.X.(*ast.Ident).Name, action.imports)
					if err != nil {
						return checkpoint.From(err)
					}
					resType = "*" + fakeName + "." + target.Sel.Name
				}

			case *ast.SelectorExpr:
				// It is a reference to another type from another package.
				fakeName, err := addImport(v.X.(*ast.Ident).Name, action.imports)
				if err != nil {
					return checkpoint.From(err)
				}
				resType = fakeName + "." + v.Sel.Name

			default:
				return checkpoint.From(errors.New("res type not supported"))
			}

			name := ""
			if len(res.Names) >= 1 {
				name = res.Names[0].Name
			} else {
				name = "res" + strconv.Itoa(i)
			}

			actionData.Response = append(actionData.Response, Param{
				Name:       name,
				NamePublic: strings.ToUpper(string(name[0])) + name[1:],
				Type:       resType,
			})

		}

		g.generated.Actions = append(g.generated.Actions, actionData)
	}

	for _, imp := range imports {
		g.generated.Imports = append(g.generated.Imports, imp)
	}

	return nil
}

//go:embed generate/template
var templateFS embed.FS

func (g Generator) Write() error {
	// Get the template.
	t, err := template.ParseFS(templateFS, "generate/template/*")
	if err != nil {
		return checkpoint.From(err)
	}

	f, err := g.FS.Create(filepath.Join(g.Out, "actions.go"))
	if err != nil {
		return checkpoint.From(err)
	}

	// Generate the file to a buffer using the template.
	var b []byte
	buf := bytes.NewBuffer(b)
	err = t.ExecuteTemplate(buf, "actions.got", g.generated)
	if err != nil {
		return checkpoint.From(err)
	}

	// Format it.
	//formatted, err := format.Source(buf.Bytes())
	//if err != nil {
	//	return checkpoint.From(err)
	//}

	// And save it.
	_, err = f.Write(buf.Bytes())
	if err != nil {
		return checkpoint.From(err)
	}

	return nil
}

func main() {
	out := flag.String("o", "plug", "sub-folder of the project to generate the new code into")
	importPath := flag.String("m", "", "module path to use if it should not be the module path from the go.mod file")
	pack := flag.String("p", "", "package name if not the same as the given output folder")

	flag.Parse()
	args := flag.Args()

	// If the package is not given, just use the folder name of out
	if *pack == "" {
		*pack = filepath.Base(*out)
	}

	g := Generator{
		In:     args[0],
		Out:    filepath.Join(args[0], *out),
		Import: *importPath,
		FS:     afero.NewOsFs(),
		Pack:   *pack,
	}

	fmt.Printf("Clean target directory %v\n", g.Out)
	err := g.CleanDestination()
	if err != nil {
		panic(err)
	}

	fmt.Println("Search for the annotation")
	err = g.Search()
	if err != nil {
		panic(err)
	}

	fmt.Println("Generate")
	err = g.Generate()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Write to target directory %v\n", g.Out)
	err = g.Write()
	if err != nil {
		panic(err)
	}
}
