package generate

import (
	"bytes"
	"embed"
	"errors"
	"github.com/aligator/checkpoint"
	"github.com/spf13/afero"
	"go/ast"
	"go/format"
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

var (
	ErrTypeNotSupported = errors.New("type not supported (some types can be activated by GoPlug options)")
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

type atomicInt32 struct {
	internal *int32
}

func (a atomicInt32) Add(delta int32) {
	if a.internal == nil {
		a.internal = new(int32)
	}
	atomic.AddInt32(a.internal, delta)
}

func (a atomicInt32) Load() int32 {
	if a.internal == nil {
		a.internal = new(int32)
	}
	return atomic.LoadInt32(a.internal)
}

type Generator struct {
	Out    string
	In     string
	Module string
	Pack   string
	FS     afero.Fs

	// AllowStructs enables the plugin-methods to use other
	// structs as param or return type.
	// Be aware that the plugins will need to import these structs and therefore
	// have the host as direct dependency.
	// So be careful what methods these structs include. All plugins will have
	// access to them. However only public fields will be sent.
	//
	// To have a strict decoupling of plugins from the host-types and
	// to avoid possible confusion for plugin-developers this
	// option should be false.
	AllowStructs bool

	// AllowPointers enables the plugin-methods to use pointers.
	// Be aware that these pointers are in fact get still copied as
	// they get serialized and de-serialized fully.
	// To avoid confusion for plugin-developers this option should
	// be false.
	AllowPointers bool

	// AllowSlices enables the plugin-methods to use slices.
	// Be aware that these slices are always copied.
	AllowSlices bool

	found        []match
	generated    *PluginData
	finalImports map[string]Import

	// importCounter just counts the imports added in order to generate
	// unique imports in any case by just appending this number.
	importCounter atomicInt32
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
			if g.Module == "" {
				g.Module = pkg.Module.Path
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

func (g *Generator) mapParamType(expr ast.Expr, actionMatch match, packageName string) (string, error) {
	// Use an internal mapper function to add the isExternal flag.
	var mapper func(expr ast.Expr, actionMatch match, packageName string, isExternal bool) (string, error)
	mapper = func(expr ast.Expr, actionMatch match, packageName string, isExternal bool) (string, error) {
		var resultType string

		switch v := expr.(type) {
		case *ast.Ident:
			resultType = v.Name
			if v.Obj == nil && !isExternal {
				return resultType, nil
			}

			resultType = packageName + "." + resultType

			// It is a object, e.g. a struct
			if !g.AllowStructs {
				return "", checkpoint.Wrap(errors.New("structs are not allowed"), ErrTypeNotSupported)
			}
		case *ast.StarExpr:
			if !g.AllowPointers {
				return "", checkpoint.Wrap(errors.New("pointers are not allowed"), ErrTypeNotSupported)
			}

			targetType, err := mapper(v.X, actionMatch, packageName, isExternal)
			if err != nil {
				return "", err
			}

			resultType = "*" + targetType
		case *ast.SelectorExpr:
			// It is a reference to another type from another package.
			fakeName, err := g.addImport(v.X.(*ast.Ident).Name, actionMatch.imports)
			if err != nil {
				return "", checkpoint.From(err)
			}

			refType, err := mapper(v.Sel, actionMatch, fakeName, true)
			if err != nil {
				return "", checkpoint.From(err)
			}
			resultType = refType
		case *ast.ArrayType:
			if !g.AllowSlices {
				return "", checkpoint.Wrap(errors.New("slices are not allowed"), ErrTypeNotSupported)
			}

			targetType, err := mapper(v.Elt, actionMatch, packageName, isExternal)
			if err != nil {
				return "", err
			}

			resultType = "[]" + targetType
		default:
			return "", checkpoint.From(ErrTypeNotSupported)
		}

		return resultType, nil
	}

	return mapper(expr, actionMatch, packageName, false)
}

// addImport adds the given name or path to the imports.
// If fileImports is given it is used to resolve it.
// If nameOrPath is a name the fileImports are mandatory, to resolve it.
func (g *Generator) addImport(nameOrPath string, fileImports []Import) (string, error) {
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
	if i, ok := g.finalImports[match.Path]; !ok {
		fakeName := match.Name
		// Make sure it is unique.
		count := g.importCounter.Load()
		fakeName += strconv.Itoa(int(count))
		g.importCounter.Add(1)

		// If not, create it.
		g.finalImports[match.Path] = Import{
			FakeName: fakeName,
			Name:     match.Name,
			Path:     match.Path,
		}

		return g.finalImports[match.Path].FakeName, nil
	} else {
		return i.FakeName, nil
	}
}

// Generate the plugin actions file
func (g *Generator) Generate() error {
	g.generated = &PluginData{
		Package: g.Pack,
	}

	g.finalImports = make(map[string]Import)

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

		importPath := filepath.ToSlash(filepath.Join(g.Module, action.path))
		fakeName, err := g.addImport(importPath, []Import{
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
			paramType, err := g.mapParamType(param.Type, action, fakeName)
			if err != nil {
				return err
			}

			actionData.Request = append(actionData.Request, Param{
				Name:       param.Names[0].Name,
				NamePublic: strings.ToUpper(string(param.Names[0].Name[0])) + param.Names[0].Name[1:],
				Type:       paramType,
			})

		}

		// Add response.
		for i, res := range action.fn.Type.Results.List {
			if i == len(action.fn.Type.Results.List)-1 {
				// As the last return type has to be an error and that
				// is handled separately, ignore it.
				// TODO: Add check if the last is really an error
				break
			}

			resType, err := g.mapParamType(res.Type, action, fakeName)
			if err != nil {
				return err
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

	for _, imp := range g.finalImports {
		g.generated.Imports = append(g.generated.Imports, imp)
	}

	return nil
}

//go:embed template
var templateFS embed.FS

func (g Generator) Write() error {
	// Get the template.
	t, err := template.ParseFS(templateFS, "template/*")
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
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return checkpoint.From(err)
	}

	// And save it.
	_, err = f.Write(formatted)
	if err != nil {
		return checkpoint.From(err)
	}

	return nil
}
