# gostarline
This is very simple experiment to see how GitHub Copilot and https://developer.starline.ru works.

This is very simple experiment to see how GitHub Copilot and https://developer.starline.ru works.

Results:
- GitHub Copilot is a quite good tool to bootstrap projects and extend them.
- https://developer.starline.ru API is very unstable and has a very savvier issue, but it was a fan to investigate and play around!

## How to build and run
- Go to https://my.starline.ru register and request api credentials
- Use these credentials to run https://gitlab.com/starline/openapi/-/blob/master/auth.py and generate slnet cookie token to use with the app
- `go build`
- `gostarline  --token=4962......26D7 --device_id=99.....99`
- or `go run . --token=4962......26D7 --device_id=99.....99`
