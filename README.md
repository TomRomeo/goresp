# goresp
A program to tests the responsiveness of websites :)

By default it will take screenshots of your specified website in the following formats:

- 1920x1080
- 3440x1440
- 390x844
- 1024x600

But you can also specify custom resolutions

## How to install
In your terminal, type:
```bash
go install github.com/TomRomeo/goresp/cmd/goresp
```

You can now use the `goresp` command, try `goresp -h`.

Before you can start with the action though, you will have to install the correct version of [selenium standalone server](https://selenium-release.storage.googleapis.com/index.html?path=4.0/)
and the correct version of the [chrome-driver](https://chromedriver.chromium.org/downloads).

After downloading both, point to them by setting the environment variables `SELENIUM_PATH` and `CHROME_DRIVER_PATH`

Now, you are ready for the action!

## How to use
### If you want to take screenshots of \<url>:
```bash
goresp -o ./output <url>
```

### If you want to take scrolling screenshots of \<url>:
```bash
goresp -f -o ./output <url>
```

### If you want to take screenshots in custom resolutions
```bash
goresp -r "1x1" -r "1200x1200" -o ./output <url>
```

### If you want to take screenshots after a delay (in seconds)
```bash
goresp -d 2 -o ./output <url>
```
