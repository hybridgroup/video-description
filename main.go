package main

import (
	"fmt"
	"os"

	"github.com/hybridgroup/mjpeg"
)

func main() {
	if err := handleFlags(); err != nil {
		showUsage()
		os.Exit(0)
	}

	stream := mjpeg.NewStream()

	go startVideoCapture(*deviceID, stream)
	go startCaptions(*modelPath, *projectorPath, *promptText)

	fmt.Println("Capturing. Point your browser to", host)

	startWebServer(*host, stream, *promptText)
}
