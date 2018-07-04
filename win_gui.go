// Copyright 2017 The Walk Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/jacobsa/go-serial/serial"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
	"golang.org/x/sys/windows/registry"
)

/* ReadSerialPortTimerID */
var (
	ReadSerialPortTimerID uintptr
	port                  io.ReadWriteCloser
	covalue               = "100"
	tempvalue             = "100"
	coLB, tempLB          *walk.Label
)

func tableViewHdrWndProc(hwnd win.HWND, msg uint32, wp, lp uintptr) uintptr {

	switch msg {
	case win.WM_TIMER:
		switch wp {
		case ReadSerialPortTimerID:
			//fmt.Println("Timer1 expierd")

			lpData := make([]byte, 300)
			n, err := port.Read(lpData)
			if err != nil {
				log.Fatalf("port.Read: %v", err)
				time.Sleep(5000 * time.Millisecond)
			}
			if n != 0 {
				fmt.Println("Read", n, "bytes.")
				//fmt.Println(fmt.Sprintf("%s", lpData))
				//canSplit := func(c rune) bool { return c == '\r' }
				//x := strings.FieldsFunc(fmt.Sprintf("%s", lpData), canSplit)
				var i int
				x := fmt.Sprintf("%s", lpData)
				z := strings.Fields(x)
				for i = 0; i < len(z); i++ {
					fmt.Println(z[i])

				}

				tempLB.SetText(z[1])
				coLB.SetText(z[3])

				time.Sleep(5 * time.Millisecond)
			}
		}
	}
	return wp
}

func main() {

	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\\DEVICEMAP\\SERIALCOMM`, registry.QUERY_VALUE)
	if err != nil {
		log.Fatal(err)
	}
	defer k.Close()

	ki, err := k.Stat()

	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Subkey %d ValueCount %d\n", ki.SubKeyCount, ki.ValueCount)

	s, err := k.ReadValueNames(int(ki.ValueCount))
	if err != nil {
		log.Fatal(err)
	}
	kvalue := make([]string, ki.ValueCount)

	for i, test := range s {
		q, _, err := k.GetStringValue(test)
		if err != nil {
			log.Fatal(err)
		}
		kvalue[i] = q
	}

	fmt.Printf("%s \n", kvalue)

	var connPB, exitPB *walk.PushButton
	var comCB *walk.ComboBox

	if _, err := (MainWindow{
		Title:  "Sensor Monitor",
		Layout: VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					Label{
						Text: "COM Port:",
					},
					ComboBox{
						AssignTo: &comCB,
						Editable: false,
						Model:    kvalue,
						//ContextMenuItems: kvalue,
					},
					PushButton{
						AssignTo: &connPB,
						Text:     "Connect",
						OnClicked: func() {

							fmt.Println("Connect  " + kvalue[comCB.CurrentIndex()])
							options := serial.OpenOptions{
								PortName:        kvalue[comCB.CurrentIndex()],
								BaudRate:        115200,
								DataBits:        8,
								StopBits:        1,
								MinimumReadSize: 4,
							}
							// Open the port.
							port, err = serial.Open(options)
							if err != nil {
								log.Fatalf("serial.Open: %v", err)
							}

							//
							var tableViewHdrWndProcPtr = syscall.NewCallback(tableViewHdrWndProc)
							ReadSerialPortTimerID = win.SetTimer(win.HWND_TOP, 0, 1000, tableViewHdrWndProcPtr)
							fmt.Println("Start Serial Port Read Timer")

						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					Label{
						Text: "CO 濃度:",
					},
					Label{
						AssignTo: &coLB,
						Text:     covalue,
					},
					Label{
						Text: "ppm",
					},
				},
			},
			HSplitter{
				Children: []Widget{
					Label{
						Text: "溫度:",
					},
					Label{
						AssignTo: &tempLB,
						Text:     tempvalue,
					},
					Label{
						Text: " ℃",
					},
				},
			},
			PushButton{
				AssignTo: &exitPB,
				Text:     "Exit",
				OnClicked: func() {
					fmt.Println("Exit")
					os.Exit(0)
				},
			},
		},
	}).Run(); err != nil {
		log.Fatal(err)
	}

}
