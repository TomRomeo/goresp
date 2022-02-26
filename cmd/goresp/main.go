package main

import (
	"context"
	"errors"
	"github.com/TomRomeo/goresp/internal"
	"github.com/tebeka/selenium"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"sync"
	"time"
)

var (
	full         bool
	seleniumPath string
	chromePath   string
	outPath      string
	res          cli.StringSlice
	port         = 5555
	delay        int
)

func main() {
	app := &cli.App{
		Name:        "goresp",
		HelpName:    "goresp",
		Usage:       "A tool to test your website for multiple resolutions",
		Description: "",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:        "full",
				Aliases:     []string{"f"},
				Usage:       "",
				EnvVars:     nil,
				FilePath:    "",
				DefaultText: "",
				Destination: &full,
			},
			&cli.StringFlag{
				Name:        "seleniumPath",
				Aliases:     []string{"s"},
				Usage:       "",
				EnvVars:     []string{"SELENIUM_PATH"},
				DefaultText: "",
				Destination: &seleniumPath,
			},
			&cli.StringFlag{
				Name:        "chromedriverPath",
				Aliases:     []string{"c"},
				Usage:       "",
				EnvVars:     []string{"CHROME_DRIVER_PATH"},
				DefaultText: "",
				Destination: &chromePath,
			},
			&cli.PathFlag{
				Name:        "outPath",
				Aliases:     []string{"o"},
				Usage:       "",
				EnvVars:     nil,
				DefaultText: ".",
				Destination: &outPath,
			},
			&cli.StringSliceFlag{
				Name:        "resolutions",
				Aliases:     []string{"r"},
				Usage:       "",
				DefaultText: "",
				Destination: &res,
			},
			&cli.IntFlag{
				Name:        "delay",
				Aliases:     []string{"d"},
				Usage:       "",
				Value:       0,
				Destination: &delay,
				HasBeenSet:  false,
			},
		},
		Action:          mainCommand,
		CommandNotFound: nil,
		OnUsageError:    nil,
		Compiled:        time.Time{},
		Authors: []*cli.Author{
			{
				Name:  "Tom Doil",
				Email: "Tom.Romeo.Doil@gmail.com",
			},
		},
		Copyright:              "",
		UseShortOptionHandling: true,
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var mainCommand = func(c *cli.Context) error {

	// TODO: validate paths

	url := c.Args().Get(0)
	if url == "" {
		return errors.New("Missing url parameter")
	}

	resolutions := res.Value()

	if len(resolutions) == 0 {
		resolutions = []string{
			"1920x1080",
			"3440x1440",
			"390x844",
			"1024x600",
		}
	}

	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	opts := []selenium.ServiceOption{
		selenium.ChromeDriver(chromePath),
	}
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer service.Stop()

	errch := make(chan error, len(resolutions))
	ctx, cancel := context.WithCancel(c.Context)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(len(resolutions))

	for _, res := range resolutions {
		go func(errch chan<- error, res string) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			default:
				if full {
					if err := internal.TakeFullScreenshot(url, res, outPath, delay, port); err != nil {
						errch <- err
						cancel()
					}
				} else {
					if err := internal.TakeScreenshot(url, res, outPath, delay, port); err != nil {
						errch <- err
						cancel()
					}
				}

			}
		}(errch, res)
	}
	wg.Wait()
	if ctx.Err() != nil {
		err := <-errch
		log.Fatal(err)
	}

	return nil
}
