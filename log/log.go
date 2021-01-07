package log

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
)

// L is the global logger.
var L *zap.Logger

func init() {
	app := strings.Replace(filepath.Base(os.Args[0]), ".exe", "", 1)
	var err error
	if content, err := ioutil.ReadFile(app + ".json"); err == nil {
		var configuration zap.Config
		if err := json.Unmarshal(content, &configuration); err == nil {
			L, err = configuration.Build()
			if err != nil {
				panic(fmt.Sprintf("error initialising logger: %v", err))
			}
			L.Info("application starting with custom configuration")
			return
		}
	}

	configuration := zap.NewProductionConfig()
	configuration.Encoding = "json" // or "console"
	configuration.OutputPaths = []string{app + ".log"}
	configuration.ErrorOutputPaths = []string{app + ".log"}
	configuration.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	L, err = configuration.Build()
	if err != nil {
		panic(fmt.Sprintf("error initialising logger: %v", err))
	}
	L.Info("application starting with default configuration")
	return
}
