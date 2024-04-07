package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/koron/go-ssdp"
)

var (
	a            fyne.App
	roku         ssdp.Service
	allRokus     []ssdp.Service
	allRokuNames []string
	scanWin      *fyne.Container
	rokuSelect   *widget.Select
	inputText    binding.String
	lastText     string
)

func main() {
	a = app.NewWithID("DesktopRokuRemote")
	w := a.NewWindow("Roku Remote")
	inputText = binding.NewString()

	//attempt to load.
	loadScan()
	sel := a.Preferences().Int("SelectedRoku")

	//empty until we scan
	scanWin = scannerWindow()
	rokuSelect.SetSelectedIndex(sel)
	c := container.New(layout.NewHBoxLayout(), scanWin, remoteWindow())
	w.SetContent(c)
	w.ShowAndRun()

}

func scannerWindow() *fyne.Container {
	findRokus := widget.NewButton("ScanNetwork", func() {
		fmt.Println("Scanning....")
		Scan()
		fmt.Println("Scanning Complete")
		elems := len(allRokus)
		fmt.Printf("%i Rokus found", elems)
		saveScan()

		//recreate select window
		scanWin.Objects[1] = RokuSelect()

		//refresh?
		scanWin.Refresh()
	})
	selectedRokuLabel := widget.NewLabel(roku.Location)
	rokuSelect = RokuSelect()

	box := container.New(layout.NewVBoxLayout(), findRokus, rokuSelect, selectedRokuLabel)

	return box
}
func Scan() {
	allRokus = FindRoku()
	allRokuNames = make([]string, len(allRokus))
	for i, r := range allRokus {
		allRokuNames[i] = r.USN
	}
}
func RokuSelect() *widget.Select {
	return widget.NewSelect(allRokuNames, func(value string) {
		for i, r := range allRokus {
			if r.USN == value {
				roku = r
				a.Preferences().SetInt("SelectedRoku", i)
			}
		}
	})
}
func saveScan() {
	j, _ := json.Marshal(allRokus)
	a.Preferences().SetString("rokus", string(j))
}
func loadScan() {
	data := a.Preferences().String("rokus")
	if len(data) == 0 {
		return
	}

	_ = json.Unmarshal([]byte(data), &allRokus)

	allRokuNames = make([]string, len(allRokus))
	for i, r := range allRokus {
		allRokuNames[i] = r.USN
	}
}

func remoteWindow() *fyne.Container {
	back := widget.NewButton("Back", func() {
		SendCommand(roku, "keypress/back")
	})
	home := widget.NewButton("home", func() {
		SendCommand(roku, "keypress/home")
	})
	toprow := container.New(layout.NewHBoxLayout(), back, layout.NewSpacer(), home)
	up := widget.NewButton("Up", func() {
		SendCommand(roku, "keypress/up")
	})
	uprow := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		up,
		layout.NewSpacer())
	left := widget.NewButton("Left", func() {
		SendCommand(roku, "keypress/left")
	})
	right := widget.NewButton("Right", func() {
		SendCommand(roku, "keypress/right")
	})
	sele := widget.NewButton("Select", func() {
		SendCommand(roku, "keypress/select")
	})
	midrow := container.New(layout.NewHBoxLayout(),
		left,
		sele,
		right,
	)
	down := widget.NewButton("Down", func() {
		SendCommand(roku, "keypress/down")
	})
	downrow := container.New(layout.NewHBoxLayout(),
		layout.NewSpacer(),
		down,
		layout.NewSpacer())

	ir := widget.NewButton("<-", func() {
		SendCommand(roku, "keypress/InstantReplay")
	})

	info := widget.NewButton("*", func() {
		SendCommand(roku, "keypress/info")
	})

	backrow := container.New(layout.NewHBoxLayout(),
		ir,
		layout.NewSpacer(),
		info,
	)
	rev := widget.NewButton("<<", func() {
		SendCommand(roku, "keypress/rev")
	})
	fwd := widget.NewButton(">>", func() {
		SendCommand(roku, "keypress/fwd")
	})
	play := widget.NewButton("Play", func() {
		SendCommand(roku, "keypress/play")
	})
	playrow := container.New(layout.NewHBoxLayout(),
		rev,
		layout.NewSpacer(),
		play,
		layout.NewSpacer(),
		fwd,
	)

	inputEntry := widget.NewEntryWithData(inputText)
	inputEntry.OnChanged = func(s string) {
		UpdateTextToRoku(s)
	}

	backspaceButton := widget.NewButton("BS", func() {
		SendCommand(roku, "keypress/Backspace")
	})

	inputrow := container.New(layout.NewHBoxLayout(), backspaceButton, inputEntry)

	box := container.New(layout.NewVBoxLayout(),
		toprow,
		uprow,
		midrow,
		downrow,
		backrow,
		playrow,
		inputrow,
	)

	return box
}

func SendTextData(s string) {
	for i, _ := range s {
		SendCommand(roku, "keypress/Lit_"+string(s[i]))
	}
}
func UpdateTextToRoku(newText string) {
	if newText == lastText {
		return
	}
	if len(newText) > len(lastText) {
		for i, c := range newText {
			if i < len(lastText) {
				if newText[i] == lastText[i] {
					continue
				}
				//we have a mismatch between last and new. Erase existing from end to i point.
			}
			//we need to add newtext here
			//todo: get the substring from here to new.
			SendTextData(string(c))

		}
		lastText = newText
		return
	} else if len(lastText) > len(newText) {
		delCount := len(lastText) - len(newText)
		for _ = range delCount {
			SendCommand(roku, "keypress/Backspace")
		}
		//now make sure we are synced up.
		matching := true
		for i, _ := range newText {
			if i < len(lastText) {
				if newText[i] != lastText[i] {
					matching = false
					break
				}
			}
			if matching {
				lastText = newText
				return
			} //else: continue below. delete and retype everything.
		}
	}
	//and the final give up and go: erase everything and retype
	for _ = range len(lastText) {
		SendCommand(roku, "keypress/Backspace")
	}
	SendTextData(newText)

	lastText = newText
}
