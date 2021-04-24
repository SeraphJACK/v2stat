package templates

import (
	"github.com/SeraphJACK/v2stat/config"
	"github.com/SeraphJACK/v2stat/db"
	"os"
	"testing"
)

func Test(t *testing.T) {
	config.Config.DbDir = ".."
	err := db.InitDb(true)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("test.html")
	if err != nil {
		panic(err)
	}
	err = Generate(file)
	if err != nil {
		panic(err)
	}
	file.Close()
}
