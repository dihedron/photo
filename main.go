package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dihedron/photo/log"
	"github.com/h2non/filetype"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

type options struct {
	Input  string `short:"i" long:"input" description:"The name of the input file" value-name:"INPUT"`
	Output string `short:"o" long:"output" description:"The name of the output file" value-name:"OUTPUT"`
}

func main() {
	defer log.L.Sync()

	opts := &options{}
	args, err := flags.Parse(opts)
	if err != nil {
		panic(err)
	}

	if len(args) == 0 {
		args = append(args, ".")
		log.L.Debug("no paths specified, assuming current directory")
	}

	for _, arg := range args {
		err := filepath.Walk(arg, walk)
		if err != nil {
			log.L.Error("error walking path", zap.String("path", arg), zap.Error(err))
		}
	}
}

func makeWalk() func(path string, info os.FileInfo, err error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.L.Error("error walking path", zap.Error(err))
			return err
		}
		if info.Name() != "." && info.Name() != ".." {
			fmt.Println(path, " --> ", info.Size())
		}
		return nil
	}
}

func walk(path string, info os.FileInfo, err error) error {
	if err != nil {
		log.L.Error("error walking path", zap.Error(err))
		return err
	}
	if info.Name() != "." && info.Name() != ".." {
		fmt.Println(path, " --> ", info.Size())

		file, err := os.Open(path)
		if err != nil {
			log.L.Error("error opening file", zap.String("file", path), zap.Error(err))
			return nil
		}
		defer file.Close()

		// We only have to pass the file header = first 261 bytes
		head := make([]byte, 261)
		_, err = file.Read(head)
		if err != nil {
			log.L.Error("error reading file", zap.String("file", path), zap.Error(err))
			return nil
		}

		// kind, _ := filetype.Match(head)
		// switch kind {
		// case filetype.Unknown:
		// 	fmt.Println("Unknown file type")
		// case filetype.JPEG:
		// 	fmt.Println("File is a JPEG")
		// }

		if filetype.IsImage(head) {
			fmt.Println("File is an image")
		} else {
			fmt.Println("Not an image")
		}
	}
	return nil
}
