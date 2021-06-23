//go:generate go run . -o actions -m github.com/aligator/goplug/example/host ./example/host
//go:generate go build -o ./example/plugin-bin  ./example/plugin
//go:generate go build -o ./example/plugin-bin  ./example/plugin2

package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/aligator/goplug/generate"
	"github.com/spf13/afero"
)

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

	g := generate.Generator{
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
