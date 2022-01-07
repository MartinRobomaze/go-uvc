package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	gouvc "github.com/MartinRobomaze/go-uvc"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	uvc := &gouvc.UVC{}
	if err := uvc.Init(); err != nil {
		log.Fatal("init: ", err)
	}
	defer uvc.Exit()
	log.Println("UVC initialized")

	devs, err := uvc.GetDevices()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("devices:", len(devs))

	dev, err := uvc.FindDevice(0, 0, "")
	if err != nil {
		log.Fatal("find device: ", err)
	}
	defer dev.Unref()

	desc, _ := dev.Descriptor()
	log.Println("device found:\n", desc)

	if err := dev.Open(); err != nil {
		log.Fatal(err)
	}
	defer dev.Close()
	log.Println("device opened")

	// dev.Diag()

	if err := dev.SetAEMode(gouvc.AEModeManual); err != nil {
		log.Fatal("set ae mode:", err)
	}

	ci := dev.ControlInterface()
	log.Println("device control interface:\n", ci)
	for i, si := range dev.StreamInterfaces() {
		log.Printf("stream interface: %d\n%s", i, si.String())
		for j, formatDesc := range si.FormatDescriptors() {
			log.Printf("format descriptor: %d\n%s", j, formatDesc.String())
			for k, frameDesc := range formatDesc.FrameDescriptors() {
				log.Printf("frame descriptor: %d\n%s", k, frameDesc.String())
			}
		}
	}

	formatDesc := dev.GetFormatDesc()

	frameFormat := gouvc.FRAME_FORMAT_Y16

	frameDesc := formatDesc.FrameDescriptors()

	fmt.Println(frameDesc)

	stream, err := dev.GetStream(
		frameFormat,
		int(frameDesc[0].Width),
		int(frameDesc[0].Height),
		int(10000000/frameDesc[0].DefaultFrameInterval),
	)
	if err != nil {
		log.Fatal("get stream:", err)
	}

	ctrl := stream.Ctrl()
	log.Println("stream ctrl:\n", ctrl.String())

	if err := stream.Open(); err != nil {
		log.Fatal("open stream:", err)
	}
	defer stream.Close()

	cf, err := stream.Start()
	if err != nil {
		log.Fatal("start streaming:", err)
	}
	defer stream.Stop()

	// stream for 10 seconds
	tc := time.After(10 * time.Second)
	for {
		select {
		case frame := <-cf:
			log.Printf("got image: %d, %dx%d, %s",
				frame.Sequence, frame.Width, frame.Height, frame.CaptureTime)
			if err := writeFrameFile(frame, fmt.Sprintf("frame%d.jpg", frame.Sequence)); err != nil {
				log.Fatal("write frame:", err)
			}
		case <-tc:
			return
		}
	}
}

func writeFrameFile(r io.Reader, name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, r)
	return err
}
