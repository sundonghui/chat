package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/sundonghui/chat/config"
	"github.com/sundonghui/chat/database"
	"github.com/sundonghui/chat/mode"
	"github.com/sundonghui/chat/model"
)

var (
	// Version of the service
	Version = "0.0.1"
	// Commit the git commit hash of this version.
	Commit = "xxx"
	// BuildDate the date when this binary was built.
	BuildDate = "xxx"
	// Mode the mode of the service.
	Mode = mode.Debug
)

func main() {
	vInfo := &model.VersionInfo{Version: Version, Commit: Commit, BuildDate: BuildDate}
	mode.Set(Mode)
	log.Println(fmt.Println("Starting version", vInfo.Version+"@"+BuildDate))

	rand.Seed(time.Now().UnixNano()) // initialize random number generator
	conf := config.Get()
	if conf.PluginsDir != "" {
		if err := os.MkdirAll(conf.PluginsDir, 0o755); err != nil {
			panic(err)
		}
	}
	if err := os.MkdirAll(conf.UploadedImagesDir, 0o755); err != nil {
		panic(err)
	}

	db, err := database.New(database.DatabaseOptions{
		Dialect:    conf.Database.Dialect,
		Connection: conf.Database.Connection,
		DefaultUserList: []database.DefaultUser{
			{
				Username: conf.DefaultUser.Name,
				Password: conf.DefaultUser.Pass,
			},
		},
		PasswordStrength: conf.PassStrength,
	})
	if err != nil {
		panic(err)
	}
	defer db.Close()

}
