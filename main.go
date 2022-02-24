package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	// These paths will be different on your system.
	seleniumPath = ""
	chromeDriver = ""
	port         = 5555
	url          = ""
)

func main() {

	resolutions := []string{
		"1920x1080",
		"3440x1440",
		"390x844",
		"1024x600",
	}

	errch := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
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
				if err := takeFullScreenshot(url, res, res+".png"); err != nil {
					errch <- err
					cancel()
				}
			}
		}(errch, res)
	}
	wg.Wait()
	if ctx.Err() != nil {
		err := <-errch
		log.Fatal(err)
	}

}

func takeScreenshot(url, dimensions, outfile string) error {
	dim := strings.Split(dimensions, "x")
	width, err := strconv.ParseUint(dim[0], 10, 0)
	if err != nil {
		return err
	}
	height, err := strconv.ParseUint(dim[1], 10, 0)
	if err != nil {
		return err
	}

	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	opts := []selenium.ServiceOption{
		selenium.ChromeDriver(chromeDriver),
	}
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	// change screen resolution
	caps.AddChrome(chrome.Capabilities{MobileEmulation: &chrome.MobileEmulation{
		DeviceMetrics: &chrome.DeviceMetrics{
			Width:      uint(width),
			Height:     uint(height),
			PixelRatio: 1,
		},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
	}})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	if err := wd.Get(url); err != nil {
		panic(err)
	}

	sc, err := wd.Screenshot()
	if err != nil {
		log.Fatalln(err)
	}
	if err = ioutil.WriteFile(outfile, sc, 666); err != nil {
		log.Fatal(err)
	}
	return nil
}
func takeFullScreenshot(url, dimensions, outfile string) error {

	heightOnPage := 0
	dim := strings.Split(dimensions, "x")
	width, err := strconv.ParseUint(dim[0], 10, 0)
	if err != nil {
		return err
	}
	height, err := strconv.ParseUint(dim[1], 10, 0)
	if err != nil {
		return err
	}

	// Start a Selenium WebDriver server instance (if one is not already
	// running).
	opts := []selenium.ServiceOption{
		selenium.ChromeDriver(chromeDriver),
	}
	service, err := selenium.NewSeleniumService(seleniumPath, port, opts...)
	if err != nil {
		panic(err) // panic is used only as an example and is not otherwise recommended.
	}
	defer service.Stop()

	// Connect to the WebDriver instance running locally.
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}
	caps.AddChrome(chrome.Capabilities{MobileEmulation: &chrome.MobileEmulation{
		DeviceMetrics: &chrome.DeviceMetrics{
			Width:      uint(width),
			Height:     uint(height),
			PixelRatio: 1,
		},
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
	}})
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		panic(err)
	}
	defer wd.Quit()

	if err := wd.Get(url); err != nil {
		panic(err)
	}

	// get maximum document height, to know when to stop scrolling
	mxHeight, err := wd.ExecuteScript("return document.documentElement.scrollHeight", nil)
	if err != nil {
		return err
	}
	maxHeight := int(mxHeight.(float64))

	bgImg := image.NewRGBA(image.Rect(0, 0, int(width), maxHeight))

	// scroll and add to image
	for {
		partScreensht, err := wd.Screenshot()
		if err != nil {
			return err
		}
		img, _, err := image.Decode(bytes.NewReader(partScreensht))
		if err != nil {
			return err
		}
		draw.Draw(bgImg, img.Bounds().Add(image.Pt(0, heightOnPage)), img, image.Point{}, draw.Over)
		wd.KeyDown(selenium.PageDownKey)
		time.Sleep(5 * time.Second)

		heightOnPage += int(height)
		if heightOnPage >= maxHeight {
			break
		}

	}
	f, err := os.Create(outfile)
	if err != nil {
		return err
	}
	defer f.Close()

	png.Encode(f, bgImg)

	return nil
}
