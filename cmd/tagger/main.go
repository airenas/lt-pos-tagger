package main

import (
	"github.com/airenas/go-app/pkg/goapp"
	"github.com/airenas/lt-pos-tagger/internal/pkg/morphology"
	"github.com/airenas/lt-pos-tagger/internal/pkg/segmentation"
	"github.com/airenas/lt-pos-tagger/internal/pkg/service"
	"github.com/labstack/gommon/color"

	"github.com/pkg/errors"
)

func main() {
	goapp.StartWithDefault()

	data := service.Data{}
	data.Port = goapp.Config.GetInt("port")
	var err error
	data.Segmenter, err = segmentation.NewClient(goapp.Config.GetString("segmentation.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init segmenter"))
	}

	data.Tagger, err = morphology.NewClient(goapp.Config.GetString("morphology.url"))
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't init tagger"))
	}

	printBanner()

	err = service.StartWebServer(&data)
	if err != nil {
		goapp.Log.Fatal(errors.Wrap(err, "Can't start the service"))
	}
}

var (
	version string
)

func printBanner() {
	banner := `
        __  ______   ____  ____  _____
       / / /_  __/  / __ \/ __ \/ ___/
      / /   / /    / /_/ / / / /\__ \ 
     / /___/ /    / ____/ /_/ /___/ / 
    /_____/_/    /_/    \____//____/  
   __                             
  / /_____ _____ _____ ____  _____
 / __/ __ ` + "`" + `/ __ ` + "`" + `/ __ ` + "`" + `/ _ \/ ___/
/ /_/ /_/ / /_/ / /_/ /  __/ /    
\__/\__,_/\__, /\__, /\___/_/   v: %s    
         /____//____/             

%s
________________________________________________________                                                 

`
	cl := color.New()
	cl.Printf(banner, cl.Red(version), cl.Green("https://github.com/airenas/lt-pos-tagger"))
}
