//go:generate go run . generate actions -m github.com/aligator/goplug/example/host --allow-structs --allow-pointers --allow-slices ./example/host
//go:generate go build -o ./example/plugin-bin ./example/plugin
//go:generate go build -o ./example/plugin-bin ./example/plugin2

package main

import (
	"fmt"
	"github.com/aligator/goplug/generate"
	"github.com/spf13/afero"
	"github.com/spf13/pflag"
	"os"
	"path/filepath"
)

func main() {
	out := pflag.StringP("out", "o", "actions", "folder, relative to the project root, to generate the action code into")
	module := pflag.StringP("module", "m", "", "module path to use if it should not be the module path from the go.mod file")
	pack := pflag.StringP("package", "p", "", "package name if not the same as the name of the given output folder")
	allowStructs := pflag.Bool("allow-structs", false, `AllowStructs enables the plugin-methods to use other structs as param or return type.
Be aware that the plugins will need to import these structs and therefore have the host as direct dependency.
So be careful what methods these structs include.

If "allow-structs" is enabled, these structs may include pointers and slices even if the respective option is disabled!
All plugins will have access to them. However only public fields will be sent.

To have a strict decoupling of plugins from the host-types and to avoid possible confusion for plugin-developers this option should be false.`)
	allowPointers := pflag.Bool("allow-pointers", false, `AllowPointers enables the plugin-methods to use pointers.
Be aware that these pointers are in fact get still copied as they get serialized and de-serialized fully. 

If "allow-structs" is enabled, these structs may include pointers and slices even if the respective option is disabled!
To avoid confusion for plugin-developers this option should be false.`)
	allowSlices := pflag.Bool("allow-slices", false, `AllowSlices enables the plugin-methods to use slices.

If "allow-structs" is enabled, these structs may include pointers and slices even if the respective option is disabled!
Be aware that these slices are always copied.
`)
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage of goplug:")
		fmt.Fprintln(os.Stderr, "goplug generate actions [ OPTION ]... { PROJECT_ROOT }")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "OPTIONS: ")
		pflag.PrintDefaults()
	}
	pflag.Parse()
	args := pflag.Args()

	// For now hardcode the usage.
	// Later, when more generation commands exist, just use cobra.
	if len(args) < 3 || os.Args[1] != "generate" || os.Args[2] != "actions" {
		pflag.Usage()
		return
	}
	// remove the "generate actions"
	args = args[2:]

	// If the package is not given, just use the folder name of out
	if *pack == "" {
		*pack = filepath.Base(*out)
	}

	g := generate.Generator{
		In:     args[0],
		Out:    filepath.Join(args[0], *out),
		Module: *module,
		FS:     afero.NewOsFs(),
		Pack:   *pack,

		AllowStructs:  *allowStructs,
		AllowPointers: *allowPointers,
		AllowSlices:   *allowSlices,
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
