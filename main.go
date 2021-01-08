package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/dihedron/photo/log"
	"github.com/fatih/color"
	"github.com/jessevdk/go-flags"
	"go.uber.org/zap"
)

// Options is the set of application options.
type Options struct {
	From     string `short:"f" long:"from" description:"The pattern of the directory where to scan." value-name:"SOURCE"`
	To       string `short:"t" long:"to" description:"The path of the directory where matching files go." value-name:"DESTINATION"`
	Pattern  string `short:"p" long:"pattern" description:"The pattern for the directory names." value-name:"PATTERN"`
	Exif     bool   `short:"e" long:"exif" description:"Whether files should be moved to directories based on their EXIF data." value-name:"ORGANISE"`
	Year     int    `short:"y" long:"year" description:"Index of the year in the regular expression matching groups." value-name:"YEAR"`
	Month    int    `short:"m" long:"month" description:"Index of the month in the regular expression matching groups." value-name:"MONTH"`
	Day      int    `short:"d" long:"day" description:"Index of the day in the regular expression matching groups." value-name:"DAY"`
	Limit    int    `short:"l" long:"limit" description:"Number of objects to process." value-name:"LIMIT"`
	Simulate bool   `short:"s" long:"simulate" description:"Whether to execute in simulation mode." value-name:"SIMULATE"`
}

func main() {
	defer log.L.Sync()

	options := &Options{}
	_, err := flags.Parse(options)
	if err != nil {
		panic(err)
	}
	// use the current directory if no source specified
	if options.From == "" {
		options.From = "."
		log.L.Warn("no source path specified, assuming current directory")
	}
	// create the destination directory; if none specified, it is a dry-run
	if options.To != "" {
		err := os.MkdirAll(options.To, 0o755)
		if err != nil {
			log.L.Error("error creating directory", zap.Error(err))
			options.To = ""
		}
	}
	if options.Limit == 0 {
		options.Limit = -1
	}
	err = filepath.Walk(options.From, makeWalk(options))
	if err != nil {
		log.L.Error("error walking path", zap.String("path", options.From), zap.Error(err))
	}
}

func makeWalk(options *Options) func(path string, info os.FileInfo, err error) error {
	count := 0
	return func(path string, info os.FileInfo, err error) error {
		white := color.New(color.FgWhite).PrintfFunc()
		red := color.New(color.FgRed).PrintfFunc()
		// yellow := color.New(color.FgYellow).PrintfFunc()
		green := color.New(color.FgGreen).PrintfFunc()
		// if a pattern is specified
		var pattern *regexp.Regexp
		if options.Pattern != "" {
			pattern, err = regexp.Compile(options.Pattern)
			if err != nil {
				log.L.Error("error compiling pattern", zap.String("pattern", options.Pattern), zap.Error(err))
				return err
			}
		}

		if err != nil {
			log.L.Error("error walking path", zap.Error(err))
			return err
		}
		if info.Name() != "." && info.Name() != ".." {
			if pattern != nil {
				match := pattern.FindStringSubmatch(info.Name())

				if (0 < options.Year && options.Year < len(match)) &&
					(0 < options.Month && options.Month < len(match)) &&
					(0 < options.Day && options.Day < len(match)) {

					year, err := strconv.Atoi(match[options.Year])
					if err != nil {
						log.L.Error("error parsing year: invalid regexp?", zap.String("regexp", options.Pattern), zap.Error(err))
					}
					month, err := strconv.Atoi(match[options.Month])
					if err != nil {
						log.L.Error("error parsing month: invalid regexp?", zap.String("regexp", options.Pattern), zap.Error(err))
					}
					day, err := strconv.Atoi(match[options.Day])
					if err != nil {
						log.L.Error("error parsing day: invalid regexp?", zap.String("regexp", options.Pattern), zap.Error(err))
					}

					if day > 0 && month > 0 && year > 0 {
						newDir := fmt.Sprintf("%04d_%02d_%02d", year, month, day)
						newPath := filepath.Join(options.To, newDir, info.Name())
						white("%s: moving %s to %s...", info.Name(), path, newPath)
						count++
						if !options.Simulate {
							if err := os.MkdirAll(filepath.Join(options.To, newDir), 0o755); err != nil {
								red(" KO! (%v)\n", err)
							} else {
								if err := os.Rename(path, newPath); err != nil {
									red(" KO! (%v)\n", err)
								} else {
									green(" OK!\n")
								}
							}
						} else {
							white(" SKIP!\n")
						}
					} else {
						red("%s: no match\n", info.Name())
					}
				}
			}

			// ext := filepath.Ext(strings.ToLower(info.Name()))
			// switch ext {
			// case ".jpg", ".jpeg":
			// 	white("%s: size %d, type: JPEG\n", info.Name(), info.Size())
			// case ".png":
			// 	yellow("%s: size %d, type: PNG\n", info.Name(), info.Size())
			// case ".mp4":
			// 	yellow("%s: size %d, type: MP4\n", info.Name(), info.Size())
			// case ".webp":
			// 	yellow("%s: size %d, type: WEBP\n", info.Name(), info.Size())
			// 	if options.Destination != "" {
			// 		newPath := filepath.Join(options.Destination, info.Name())
			// 		log.L.Info("moving file to junk", zap.String("oldPath", path), zap.String("newPath", newPath))
			// 		err := os.Rename(path, newPath)
			// 		if err == nil {
			// 			log.L.Info("file moved", zap.String("oldPath", path), zap.String("newPath", newPath))
			// 		} else {
			// 			log.L.Error("error moving file", zap.String("oldPath", path), zap.String("newPath", newPath), zap.Error(err))
			// 		}
			// 	}
			// default:
			// 	red("%s: size %d, type: %s\n", info.Name(), info.Size(), ext)
			// }
		}

		if options.Limit > 0 && count >= options.Limit {
			os.Exit(0)
		}
		return nil
	}
}

/*
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
*/
