package internal

import (
	"bytes"
	"fmt"
	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
	"image"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

func TakeScreenshot(url, dimensions, outFolder string, port int) error {
	dim := strings.Split(dimensions, "x")
	width, err := strconv.ParseUint(dim[0], 10, 0)
	if err != nil {
		return err
	}
	height, err := strconv.ParseUint(dim[1], 10, 0)
	if err != nil {
		return err
	}

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
		return err
	}
	if err := os.MkdirAll(outFolder, 666); err != nil {
		return err
	}
	if err = ioutil.WriteFile(path.Join(outFolder, dimensions+".png"), sc, 666); err != nil {
		return err
	}
	return nil
}

func TakeFullScreenshot(url, dimensions, outFolder string, port int) error {
	var error error

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
		return err
	}
	defer func() {
		if err := wd.Quit(); err != nil {
			error = err
		}
	}()

	if err := wd.Get(url); err != nil {
		return err
	}

	// get maximum document height, to know when to stop scrolling
	mxHeight, err := wd.ExecuteScript("return document.documentElement.scrollHeight", nil)
	if err != nil {
		return err
	}
	maxHeight := int(mxHeight.(float64))

	bgImg := image.NewRGBA(image.Rect(0, 0, int(width), maxHeight))

	// scroll and add to image
	for i := 0; ; i++ {
		partScreensht, err := wd.Screenshot()
		if err != nil {
			return err
		}
		img, _, err := image.Decode(bytes.NewReader(partScreensht))
		if err != nil {
			return err
		}
		draw.Draw(bgImg, img.Bounds().Add(image.Pt(0, heightOnPage)), img, image.Point{}, draw.Over)
		if err := wd.KeyDown(selenium.PageDownKey); err != nil {
			return err
		}
		if err := wd.KeyUp(selenium.PageDownKey); err != nil {
			return err
		}
		if err := wd.SetImplicitWaitTimeout(5 * time.Second); err != nil {
			return err
		}

		if heightOnPage >= maxHeight {
			break
		}
		heightOnPage += int(height)

		if i == 0 {

			// delete all fixed items
			// this is useful for navbars and similar items
			_, err = wd.ExecuteScript(`

				[].forEach.call(document.querySelectorAll('*'), function(el) {
					if (window.getComputedStyle(el).position === 'fixed') {
						el.remove()
					}
					});
		
			`, nil)
			if err != nil {
				return err
			}
			if err := wd.SetImplicitWaitTimeout(2 * time.Second); err != nil {
				return err
			}

		}
	}
	if err := os.MkdirAll(outFolder, 666); err != nil {
		return err
	}
	f, err := os.Create(path.Join(outFolder, dimensions+".png"))
	if err != nil {
		return err
	}
	defer f.Close()

	if err := png.Encode(f, bgImg); err != nil {
		return err
	}

	return error
}
