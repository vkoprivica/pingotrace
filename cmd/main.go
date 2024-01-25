package main

import (
	"context"
	"fmt"
	"image/color"
	"os"
	"pingotrace/pingotrace"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var placeHolderText1 string
var minRowVisible int

func main() {
	// context slise to collect Function calncelation
	var cancelFuncs []context.CancelFunc

	fyneApp := app.NewWithID("net.pingotrace")
	win := fyneApp.NewWindow("PinGoTrace")

	// Disable this when compiling
	defaultFont, err := os.ReadFile("cmd\\assets\\fonts\\AssociateSans-Regular.ttf")
	if err != nil {
		panic(err)
	}
	fontResource := fyne.NewStaticResource("Default Font", defaultFont)
	welcomeText := pingotrace.ReadFile("cmd\\assets\\Welcome.txt")
	licenseText := pingotrace.ReadFile("cmd\\assets\\License.txt")

	// Enable this when compiling
	// fontResource := fyne.Resource(resourceAssociateSansRegularTtf)
	// welcomeText := string(fyne.Resource(resourceWelcomeTxt).Content())
	// licenseText := string(fyne.Resource(resourceLicenseTxt).Content())

	// Dark theme
	darkColors := map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:      color.RGBA{0, 0, 0, 255},
		theme.ColorNameForeground:      color.RGBA{255, 255, 255, 255}, // Text color in dark mode
		theme.ColorNameButton:          color.RGBA{25, 25, 25, 255},
		theme.ColorNameInputBackground: color.RGBA{25, 25, 25, 255},
		theme.ColorNameInputBorder:     color.RGBA{35, 35, 35, 255},
		theme.ColorNameMenuBackground:  color.RGBA{0, 0, 0, 255},
	}

	// Light theme
	lightColors := map[fyne.ThemeColorName]color.Color{
		theme.ColorNameBackground:      color.RGBA{255, 255, 255, 255},
		theme.ColorNameForeground:      color.RGBA{0, 0, 0, 255}, // Text color in light mode
		theme.ColorNameButton:          color.RGBA{240, 240, 240, 255},
		theme.ColorNameInputBackground: color.RGBA{240, 240, 240, 255},
		theme.ColorNameInputBorder:     color.RGBA{240, 240, 240, 255},
		theme.ColorNameMenuBackground:  color.RGBA{255, 255, 255, 255},
	}

	// Overwrite default theme
	customTheme := &myTheme{
		myFont:      fontResource,
		fontSize:    12,
		darkColors:  darkColors,
		lightColors: lightColors,
	}

	// Place Holder Text
	placeHolderText1 = "Please enter up to 30 text lines:"
	placeHolderText2 := "Working ..."
	placeHolderText3 := "No IPv4 Found!"

	// Loading custom theme
	minRowVisible = 31
	fyneApp.Settings().SetTheme(customTheme)
	entryField := newTappableEntry(string(welcomeText))
	entryField.SetPlaceHolder(string(welcomeText))
	entryField.SetMinRowsVisible(minRowVisible)

	// Adding botom container
	vBoxCenter := container.NewVBox()
	vBoxCenter.Resize(fyne.NewSize(980, 537))
	// Adding main container
	mainBox := container.NewBorder(nil, nil, nil, nil, nil)
	hBoxTop := container.NewHBox()

	var userInput string

	// Define the buttons
	var btDNSBack *widget.Button
	btDNSPTRtoIP := widget.NewButton("DNS/PTR to IP", func() {})
	btPing := widget.NewButton("\u221E PING", func() {})
	btTrace := widget.NewButton("TRACE", func() {})
	btPinGoTrace := widget.NewButton("PINGOTRACE", func() {})
	btContinuousTrace := widget.NewButton("\u221E TRACE", func() {})
	btIPConfig := widget.NewButton("IP CONFIG", func() {})
	btStopBack := widget.NewButton("STOP", func() {})
	btMainClear := widget.NewButton("CLEAR", func() {})
	btLicense := widget.NewButton("LICENSE", func() {})
	btDark := widget.NewButton("DARK", func() {})
	btLight := widget.NewButton("LIGHT", func() {})

	btParser := widget.NewButton("DOMAIN/IP PARSER", func() {
		// Get the current text from the entry field
		userInput = entryField.Text
		if len(userInput) == 0 {
			entryField.SetText("")
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
		} else {
			parsedInput := pingotrace.ParseInput(entryField.Text)
			parsedStrings, _ := parsedInput.([]string)
			resultParsedInput := strings.Join(parsedStrings, "\n")
			hBoxTop.RemoveAll()
			hBoxTop = container.NewHBox(btDNSBack, layout.NewSpacer(), btDark, btLight)
			mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
			win.SetContent(mainBox)
			entryField.SetText(resultParsedInput)
		}
	})

	// Define a new button labeled "DNS/PTR" with the associated behavior on click
	btDNSPTRLookup := widget.NewButton("DNS/PTR", func() {
		// Get the current text from the entry field
		userInput = entryField.Text
		// Parse the input using the pingotrace function
		rawParsedInput := pingotrace.ParseInput(entryField.Text)
		// Use type assertion to determine the type of parsed input (either string or slice of strings)
		switch parsedInput := rawParsedInput.(type) {

		case string: // If the parsed input is a string, treat it as an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // Exit the function after displaying the error

		case []string: // If the parsed input is a slice of strings
			// If the slice is empty (i.e., the user pressed the button without any input)
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
			} else { // If there's some input to work with
				// Update the top horizontal box layout
				hBoxTop.RemoveAll()
				hBoxTop = container.NewHBox(btDNSBack, layout.NewSpacer(), btDark, btLight)
				mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
				win.SetContent(mainBox)
				win.Resize(fyne.NewSize(980, 537))

				// Clear the entry field and set a new placeholder text
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)

				// Create a cancellable context
				ctx, cancel := context.WithCancel(context.Background())
				// Add the cancel function to a global slice to allow cancellation later
				cancelFuncs = append(cancelFuncs, cancel)

				// Create a channel for receiving DNS results
				resultsChan := make(chan map[string][]interface{})
				// Create a channel to signal when displaying of results is done
				doneChan := make(chan bool)

				// Goroutine to fetch DNS/PTR results
				go func() {
					results, _ := pingotrace.DNSPTR(ctx, parsedInput)
					// Send results to the channel or return if the operation was cancelled
					select {
					case resultsChan <- results:
					case <-ctx.Done():
						return
					}
				}()

				// Goroutine to process and display the fetched results
				go func() {
					results := <-resultsChan
					resultText := ""
					// Compile results into a readable format
					for key, value := range results {
						resultText += fmt.Sprintf("%s: %s\n", key, value[0])
					}
					entryField.SetText(resultText)
					doneChan <- true
				}()

				// Goroutine to finalize the UI updates once results are displayed
				go func() {
					<-doneChan
					// Reset the UI elements
					hBoxTop.RemoveAll()
					hBoxTop = container.NewHBox(btDNSBack, layout.NewSpacer(), btDark, btLight)
					mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
					win.SetContent(mainBox)
					win.Resize(fyne.NewSize(980, 537))
				}()
			}
		}
	})

	btDNSPTRtoIP = widget.NewButton("DNS/PTR to IP", func() {
		// Get the current text from the entry field
		userInput = entryField.Text

		// Parse the input using the pingotrace function
		rawParsedInput := pingotrace.ParseInput(entryField.Text)

		// Use type assertion to determine the type of parsed input (either string or slice of strings)
		switch parsedInput := rawParsedInput.(type) {

		case string: // If the parsed input is a string, treat it as an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // Exit the function after displaying the error

		case []string: // If the parsed input is a slice of strings
			// If the slice is empty (i.e., the user pressed the button without any input)
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
			} else { // If there's some input to work with
				// Update the top horizontal box layout
				hBoxTop.RemoveAll()
				hBoxTop = container.NewHBox(btDNSBack, layout.NewSpacer(), btDark, btLight)
				mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
				win.SetContent(mainBox)
				win.Resize(fyne.NewSize(980, 537))

				// Clear the entry field and set a new placeholder text
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)

				// Create a cancellable context
				ctx, cancel := context.WithCancel(context.Background())
				// Add the cancel function to a global slice to allow cancellation later
				cancelFuncs = append(cancelFuncs, cancel)

				// Create a channel for receiving IP addresses
				ipAddressChan := make(chan []string)
				// Create a channel to signal when displaying of results is done
				doneChan := make(chan bool)

				// Goroutine to fetch IP addresses using DNSPTRtoIP function
				go func() {
					ipAddresses := pingotrace.DNSPTRtoIP(ctx, parsedInput)
					// Send IP addresses to the channel or return if the operation was cancelled
					select {
					case ipAddressChan <- ipAddresses:
					case <-ctx.Done():
						return
					}
				}()

				// Goroutine to process and display the fetched results
				go func() {
					ipAddresses := <-ipAddressChan
					if len(ipAddresses) == 0 {
						entryField.SetText(placeHolderText3)
					} else {
						ipAddressesText := strings.Join(ipAddresses, "\n")
						entryField.SetText(ipAddressesText)
					}
					doneChan <- true
				}()

				// Goroutine to finalize the UI updates once results are displayed
				go func() {
					<-doneChan
					// Reset the UI elements
					hBoxTop.RemoveAll()
					hBoxTop = container.NewHBox(btDNSBack, layout.NewSpacer(), btDark, btLight)
					mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
					win.SetContent(mainBox)
					win.Resize(fyne.NewSize(980, 537))
				}()
			}
		}
	})

	btDNSBack = widget.NewButton("BACK", func() {
		// 1. Cancel any ongoing operations
		for _, cancel := range cancelFuncs {
			cancel()
		}
		// Reset the cancelFuncs slice for future operations
		cancelFuncs = []context.CancelFunc{}

		// 2. Reset the UI
		hBoxTop = container.NewHBox(btParser, btDNSPTRLookup, btDNSPTRtoIP, btPing, btTrace, btPinGoTrace, btContinuousTrace, btIPConfig, btMainClear, btLicense, layout.NewSpacer(), btDark, btLight)
		vBoxCenter.Objects = nil
		vBoxCenter.Add(entryField)
		mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
		win.SetContent(mainBox)
		win.Resize(fyne.NewSize(980, 537))

		// 3. Reset the text to the previous input
		entryField.SetText(userInput)
		entryField.SetPlaceHolder(placeHolderText1)
	})

	var shouldUpdate bool
	btStopBack = widget.NewButton("STOP/BACK", func() {
		for _, cancel := range cancelFuncs {
			cancel()
		}
		// Stop updating the screen.
		shouldUpdate = false
		cancelFuncs = []context.CancelFunc{}
		// entryField.SetPlaceHolder(placeHolderText1)
		vBoxCenter.RemoveAll()
		entryField.SetText(userInput)
		vBoxCenter.Add(entryField)
		hBoxTop = container.NewHBox(btParser, btDNSPTRLookup, btDNSPTRtoIP, btPing, btTrace, btPinGoTrace, btContinuousTrace, btIPConfig, btMainClear, btLicense, layout.NewSpacer(), btDark, btLight)
		mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
		win.SetContent(mainBox)
		win.Resize(fyne.NewSize(980, 537))
	})

	var prevWidth float32 = 980 // initial width
	// This function updates the hashes based on the window width
	updateHashRows := func(width float32) {
		numOfHashes := (width / 8)
		hashRow := strings.Repeat("#", int(numOfHashes))

		// Iterate through all the widgets in your vBoxCenter and find hash rows to update
		for _, obj := range vBoxCenter.Objects {
			lbl, isLabel := obj.(*widget.Label)
			if isLabel && strings.HasPrefix(lbl.Text, "#") {
				lbl.SetText(hashRow)
			}
		}
	}

	ticker := time.NewTicker(100 * time.Millisecond)
	go func() {
		for range ticker.C {
			currentWidth := win.Canvas().Size().Width
			if currentWidth != prevWidth {
				updateHashRows(currentWidth)
				prevWidth = currentWidth
			}
		}
	}()

	btPing = widget.NewButton("\u221E PING", func() {
		// Stop all previous ping goroutines
		for _, cancel := range cancelFuncs {
			cancel()
		}
		// Clear the cancelFuncs slice
		cancelFuncs = []context.CancelFunc{}

		// Set the update flag to false
		shouldUpdate = false
		// Create a new context for the current ping
		ctx, cancel := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, cancel)
		shouldUpdate = true

		// Save previous user input into variable
		userInput = entryField.Text
		rawParsedInput := pingotrace.ParseInput(entryField.Text)

		switch parsedInput := rawParsedInput.(type) {
		case string: // this is an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // exit the function or handle this error scenario further

		case []string: // this is a slice of strings
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
			} else {
				vScrollBoxCenter := container.NewVScroll(vBoxCenter)
				vScrollBoxCenter.Resize(fyne.NewSize(980, 537))

				hScrollBoxCenter := container.NewHScroll(vScrollBoxCenter)
				hScrollBoxCenter.Resize(fyne.NewSize(980, 537))

				hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
				mainBox = container.NewBorder(hBoxTop, nil, nil, nil, hScrollBoxCenter)
				win.SetContent(mainBox)

				numOfHashes := 121
				numOfColumns := 14
				vBoxCenter.RemoveAll()
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)
				vBoxCenter.Add(entryField)
				tableListPing := []*fyne.Container{}
				dnsPTRResults, dnsPTRKeys := pingotrace.DNSPTR(ctx, parsedInput)
				ipv4Addresses := []string{}

				dnsPTRResults = pingotrace.RemoveDuplicatesMap(dnsPTRResults)
				// Remove field as Ping can display info
				vBoxCenter.RemoveAll()
				for dnsPTRKey := range dnsPTRKeys {
					for key, value := range dnsPTRResults {
						if dnsPTRKeys[dnsPTRKey] == key {
							var ipAddr, host string
							if pingotrace.CheckIPv4(key) {
								ipAddr = key
								host = value[0].(string)
								if value[len(value)-1] == false {
									labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", ipAddr, ipAddr)
									vBoxCenter.Add(widget.NewLabel(labelTextPing))
								} else {
									labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", ipAddr, host)
									vBoxCenter.Add(widget.NewLabel(labelTextPing))
								}
								if !contains(ipv4Addresses, ipAddr) {
									ipv4Addresses = append(ipv4Addresses, ipAddr)
								}
							} else {
								if value[len(value)-1] == false {
									labelTextPing := fmt.Sprintf("%s: %s", key, value[0].(string))
									vBoxCenter.Add(widget.NewLabel(labelTextPing))
									hashRow := strings.Repeat("#", numOfHashes)
									hashLabel := widget.NewLabel(hashRow)
									vBoxCenter.Add(hashLabel)
									continue
								} else {
									ipAddr = value[0].(string)
									host = key
									labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", host, ipAddr)
									vBoxCenter.Add(widget.NewLabel(labelTextPing))
									if !contains(ipv4Addresses, ipAddr) {
										ipv4Addresses = append(ipv4Addresses, ipAddr)
									}
								}
							}
							tablePing := createTable(2, numOfColumns)
							vBoxCenter.Add(tablePing)
							tableListPing = append(tableListPing, tablePing)
							hashRow := strings.Repeat("#", numOfHashes)
							hashLabel := widget.NewLabel(hashRow)
							vBoxCenter.Add(hashLabel)
						}
					}
				}

				// Create an order channel with a buffer size equal to the number of goroutines
				pingOrderChan := make(chan struct{}, len(ipv4Addresses))

				// Preload the channel with an empty struct for each goroutine
				for i := 0; i < len(ipv4Addresses); i++ {
					pingOrderChan <- struct{}{}
				}

				var wgPing sync.WaitGroup
				chPing := make(chan string)
				for index, ipAddress := range ipv4Addresses {
					wgPing.Add(1)
					go func(ipAddress string, table *fyne.Container, ctx context.Context, index int) {
						defer wgPing.Done()
						cellIndex := 0
						pingSleepDuration := 1 * time.Second
						for {
							select {
							case <-ctx.Done():
								return
							default:
								<-pingOrderChan // Wait for our turn to update

								rawDurationTime, _ := pingotrace.Ping(ipAddress)
								pingResult := ""

								if rawDurationTime > 0 && rawDurationTime < 500*time.Microsecond { // Less than 0.5 ms
									pingResult = "< 1 ms"
								} else if rawDurationTime >= 500*time.Microsecond {
									pingResult = fmt.Sprintf("%.0f ms", float64(rawDurationTime)/float64(time.Millisecond))
								} else { // If duration is 0 or error
									pingResult = "TIMEOUT"
								}

								// pingDisplayMutex.Lock()
								statusLabel := table.Objects[cellIndex].(*canvas.Text)
								if rawDurationTime > 0 {
									statusLabel.Text = "   !"
									statusLabel.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
								} else {
									statusLabel.Text = "   ."
									statusLabel.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
								}
								if shouldUpdate {
									win.Canvas().Refresh(statusLabel)
								}

								label := table.Objects[cellIndex+numOfColumns].(*canvas.Text)
								if rawDurationTime > 0 {
									label.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
								} else {
									label.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
								}

								label.Text = "   " + pingResult
								win.Canvas().Refresh(label)
								cellIndex++
								if cellIndex == numOfColumns {
									time.Sleep(pingSleepDuration)
									for i := 0; i < numOfColumns; i++ {
										statusLabel := table.Objects[i].(*canvas.Text)
										statusLabel.Text = ""
										win.Canvas().Refresh(statusLabel)
										label := table.Objects[i+numOfColumns].(*canvas.Text)
										label.Text = ""
										win.Canvas().Refresh(label)
									}
									cellIndex = 0
								}
								time.Sleep(pingSleepDuration)
								// Signal that we're done updating
								pingOrderChan <- struct{}{}
							}
						}
					}(ipAddress, tableListPing[index], ctx, index)
				}
				go func() {
					wgPing.Wait()
					close(chPing)
				}()
			}
		}
	})

	// traceResultChans := make(map[string]chan []string)
	btTrace = widget.NewButton("TRACE", func() {
		// Stop all previous traceroute goroutines
		for _, cancel := range cancelFuncs {
			cancel()
		}
		// traceResultChans = make(map[string]chan []string)
		// Set the update flag to false
		shouldUpdate = false
		// Create a new context for the current ping
		ctx, cancel := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, cancel)
		// Set the update flag to true
		shouldUpdate = true

		// Save previous user input into variable
		userInput = entryField.Text
		rawParsedInput := pingotrace.ParseInput(entryField.Text)

		switch parsedInput := rawParsedInput.(type) {
		case string: // this is an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // exit the function or handle this error scenario further

		case []string: // this is a slice of strings
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
				// if not empty - proceed with the code
			} else {
				vBoxCenter.RemoveAll()
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)
				vBoxCenter.Add(entryField)
				results, keys := pingotrace.DNSPTR(ctx, parsedInput)

				var ipAddr, host string
				for key, value := range results {
					if key == keys[0] {
						if pingotrace.CheckIPv4(key) {
							ipAddr = key
							host = value[0].(string)
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, host))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}

						} else {
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetPlaceHolder(fmt.Sprintf("%s: %s", key, value[0].(string)))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								// log.Println(key, value, "print3")
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								host = key
								ipAddr = value[0].(string)
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", host, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								// log.Println(key, value, "print4")
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}

						}
						break
					}
				}

				if ipAddr != "" {
					maxHops := 30
					timeout := 1 * time.Second

					traceOutputChan := make(chan []string, maxHops)
					// Goroutine
					go func() {
						pingotrace.Trace(ipAddr, maxHops, timeout, ctx, traceOutputChan)
					}()
					go func() {
						for {
							select {
							case line, ok := <-traceOutputChan:
								if !ok {
									return
								}
								if !shouldUpdate {
									cancel() // cancel the context
									return
								}
								result := fmt.Sprintf("%2s\t", line[0])
								for i := 2; i < len(line); i++ {
									if line[i] == "*" {
										result += fmt.Sprintf("%-10s      \t", line[i]) // Add 6 more spaces after "*"
									} else {
										result += fmt.Sprintf("%-10s\t", line[i])
									}
								}
								result += fmt.Sprintf("%-45s", line[1])
								entryField.SetText(entryField.Text + result + "\n")
								entryField.Refresh() // Notify Fyne to repaint the widget
							case <-ctx.Done():
								// context canceled
								return
							}
						}
					}()
				}
			}
		}
	})

	btPinGoTrace = widget.NewButton("PINGOTRACE", func() {
		// Stop all previous traceroute goroutines
		for _, cancel := range cancelFuncs {
			cancel()
		}

		// Set the update flag to false
		shouldUpdate = false

		// Create a new context for the current ping
		ctx, cancel := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, cancel)
		// Set the update flag to true
		shouldUpdate = true

		// Save previous user input into variable
		userInput = entryField.Text
		rawParsedInput := pingotrace.ParseInput(entryField.Text)

		switch parsedInput := rawParsedInput.(type) {
		case string: // this is an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // exit the function or handle this error scenario further

		case []string: // this is a slice of strings
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
				// if not empty - proceed with the code
			} else {
				vBoxCenter.RemoveAll()
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)
				vBoxCenter.Add(entryField)
				results, keys := pingotrace.DNSPTR(ctx, parsedInput)

				var ipAddr, host string
				for key, value := range results {
					if key == keys[0] {
						if pingotrace.CheckIPv4(key) {
							ipAddr = key
							host = value[0].(string)
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, host))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}
						} else {
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetPlaceHolder(fmt.Sprintf("%s: %s", key, value[0].(string)))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								host = key
								ipAddr = value[0].(string)
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", host, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}
						}
						break
					}
				}

				if ipAddr != "" {
					maxHops := 30
					timeout := 1 * time.Second

					// Wrap the traceroute function to receive the output
					traceOutputChan := make(chan []string, maxHops)

					var wgPinGoPath sync.WaitGroup
					// Goroutine
					wgPinGoPath.Add(1)
					go func() {
						defer wgPinGoPath.Done()
						pingotrace.PinGoTrace(ipAddr, maxHops, timeout, ctx, traceOutputChan)
						close(traceOutputChan)
					}()

					var dstHops []string
					// dstHopsCh := make(chan string)

					wgPinGoPath.Add(1)
					go func() {
						defer wgPinGoPath.Done()
						for {
							select {
							case traceLine, ok := <-traceOutputChan:
								if !ok {
									return
								}
								if !shouldUpdate {
									cancel() // cancel the context
									return
								}

								result := fmt.Sprintf("%2s\t", traceLine[0])
								for i := 2; i < len(traceLine); i++ {
									if traceLine[i] == "*" {
										result += fmt.Sprintf("%-10s      \t", traceLine[i]) // Add 6 more spaces after "*"
									} else {
										result += fmt.Sprintf("%-10s\t", traceLine[i])
									}
								}
								result += fmt.Sprintf("%-45s", traceLine[1])

								// Assuming you're using Fyne, updates to GUI components should be done in the main thread.
								// If Fyne provides a mechanism to run on the main thread (like RunOnMain), use it.
								entryField.SetText(entryField.Text + result + "\n")
								entryField.Refresh() // Notify Fyne to repaint the widget

								// Combine all elements in the line slice into a single string
								// Check if the whole line is a valid IPv4
								wholeLine := strings.Join(traceLine, " ")
								ipAddr := pingotrace.ParseIPv4(wholeLine)
								if pingotrace.CheckIPv4(ipAddr) {
									// mu.Lock()
									ipAddr = pingotrace.ParseIPv4(ipAddr)
									// dstHopsCh <- ipAddr
									dstHops = append(dstHops, ipAddr)
									// dstHopsCh <- dstHops

								}
								// fmt.Printf("first %v", dstHops)
							case <-ctx.Done():
								// fmt.Println("Context cancellation detected!")
								dstHops = []string{}
								return
							}
						}
					}()

					go func() {
						wgPinGoPath.Wait()
						// fmt.Printf("second %v", dstHops)
						time.Sleep(3 * time.Second)
						parsedInput := dstHops
						if len(parsedInput) == 0 {
							// fmt.Println("len0")
							entryField.SetText("")
							vBoxCenter.RemoveAll()
							entryField.SetText(userInput)
							vBoxCenter.Add(entryField)
						} else {
							vScrollBoxCenter := container.NewVScroll(vBoxCenter)
							vScrollBoxCenter.Resize(fyne.NewSize(980, 537))

							hScrollBoxCenter := container.NewHScroll(vScrollBoxCenter)
							hScrollBoxCenter.Resize(fyne.NewSize(980, 537))

							hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
							mainBox = container.NewBorder(hBoxTop, nil, nil, nil, hScrollBoxCenter)
							win.SetContent(mainBox)

							numOfHashes := 121
							numOfColumns := 14
							vBoxCenter.RemoveAll()
							entryField.SetText("")
							entryField.SetPlaceHolder(placeHolderText2)
							entryField.SetMinRowsVisible(minRowVisible)
							vBoxCenter.Add(entryField)
							tableListPing := []*fyne.Container{}
							dnsPTRResults, dnsPTRKeys := pingotrace.DNSPTR(ctx, parsedInput)
							ipv4Addresses := []string{}

							dnsPTRResults = pingotrace.RemoveDuplicatesMap(dnsPTRResults)
							// Remove field as Ping can display info
							vBoxCenter.RemoveAll()
							for dnsPTRKey := range dnsPTRKeys {
								for key, value := range dnsPTRResults {
									if dnsPTRKeys[dnsPTRKey] == key {
										var ipAddr, host string
										if pingotrace.CheckIPv4(key) {
											ipAddr = key
											host = value[0].(string)
											if value[len(value)-1] == false {
												labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", ipAddr, ipAddr)
												vBoxCenter.Add(widget.NewLabel(labelTextPing))
											} else {
												labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", ipAddr, host)
												vBoxCenter.Add(widget.NewLabel(labelTextPing))
											}
											if !contains(ipv4Addresses, ipAddr) {
												ipv4Addresses = append(ipv4Addresses, ipAddr)
											}
										} else {
											if value[len(value)-1] == false {
												labelTextPing := fmt.Sprintf("%s: %s", key, value[0].(string))
												vBoxCenter.Add(widget.NewLabel(labelTextPing))
												hashRow := strings.Repeat("#", numOfHashes)
												hashLabel := widget.NewLabel(hashRow)
												vBoxCenter.Add(hashLabel)
												continue
											} else {
												ipAddr = value[0].(string)
												host = key
												labelTextPing := fmt.Sprintf("Pinging %s [%s] with 32 bytes of data:", host, ipAddr)
												vBoxCenter.Add(widget.NewLabel(labelTextPing))
												if !contains(ipv4Addresses, ipAddr) {
													ipv4Addresses = append(ipv4Addresses, ipAddr)
												}
											}
										}
										tablePing := createTable(2, numOfColumns)
										vBoxCenter.Add(tablePing)
										tableListPing = append(tableListPing, tablePing)
										hashRow := strings.Repeat("#", numOfHashes)
										hashLabel := widget.NewLabel(hashRow)
										vBoxCenter.Add(hashLabel)
									}
								}
							}

							// Create an order channel with a buffer size equal to the number of goroutines
							orderChan := make(chan struct{}, len(ipv4Addresses))

							// Preload the channel with an empty struct for each goroutine
							for i := 0; i < len(ipv4Addresses); i++ {
								orderChan <- struct{}{}
							}

							var wgPinGoPath sync.WaitGroup
							chPinGoPath := make(chan string)
							for index, ipAddress := range ipv4Addresses {
								wgPinGoPath.Add(1)
								go func(ipAddress string, table *fyne.Container, ctx context.Context, index int) {
									defer wgPinGoPath.Done()
									cellIndex := 0
									pingSleepDuration := 1 * time.Second
									for {
										select {
										case <-ctx.Done():
											return
										default:
											<-orderChan // Wait for our turn to update

											rawDurationTime, _ := pingotrace.Ping(ipAddress)
											pingResult := ""

											if rawDurationTime > 0 && rawDurationTime < 500*time.Microsecond { // Less than 0.5 ms
												pingResult = "< 1 ms"
											} else if rawDurationTime >= 500*time.Microsecond {
												pingResult = fmt.Sprintf("%.0f ms", float64(rawDurationTime)/float64(time.Millisecond))
											} else { // If duration is 0 or error
												pingResult = "TIMEOUT"
											}

											statusLabel := table.Objects[cellIndex].(*canvas.Text)
											if rawDurationTime > 0 {
												statusLabel.Text = "   !"
												statusLabel.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
											} else {
												statusLabel.Text = "   ."
												statusLabel.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
											}
											if shouldUpdate {
												win.Canvas().Refresh(statusLabel)
											}

											label := table.Objects[cellIndex+numOfColumns].(*canvas.Text)
											if rawDurationTime > 0 {
												label.Color = color.RGBA{R: 0, G: 255, B: 0, A: 255}
											} else {
												label.Color = color.RGBA{R: 255, G: 0, B: 0, A: 255}
											}

											label.Text = "   " + pingResult
											win.Canvas().Refresh(label)
											cellIndex++
											if cellIndex == numOfColumns {
												time.Sleep(pingSleepDuration)
												for i := 0; i < numOfColumns; i++ {
													statusLabel := table.Objects[i].(*canvas.Text)
													statusLabel.Text = ""
													win.Canvas().Refresh(statusLabel)
													label := table.Objects[i+numOfColumns].(*canvas.Text)
													label.Text = ""
													win.Canvas().Refresh(label)
												}
												cellIndex = 0
											}
											time.Sleep(pingSleepDuration)
											orderChan <- struct{}{}
										}
									}
								}(ipAddress, tableListPing[index], ctx, index)
							}
							go func() {
								wgPinGoPath.Wait()
								close(chPinGoPath)
							}()
						}
					}()
				}
			}
		}
	})

	btContinuousTrace = widget.NewButton("\u221E TRACE", func() {
		// Stop all previous traceroute goroutines
		for _, cancel := range cancelFuncs {
			cancel()
		}
		// Set the update flag to false
		shouldUpdate = false

		// Create a new context for the current ping
		ctx, cancel := context.WithCancel(context.Background())
		cancelFuncs = append(cancelFuncs, cancel)

		// Set the update flag to true
		shouldUpdate = true

		// Save previous user input into variable
		userInput = entryField.Text
		rawParsedInput := pingotrace.ParseInput(entryField.Text)

		switch parsedInput := rawParsedInput.(type) {
		case string: // this is an error message
			entryField.SetText(parsedInput)
			vBoxCenter.RemoveAll()
			vBoxCenter.Add(entryField)
			return // exit the function or handle this error scenario further

		case []string: // this is a slice of strings
			if len(parsedInput) == 0 {
				entryField.SetText("")
				vBoxCenter.RemoveAll()
				vBoxCenter.Add(entryField)
				// if not empty - proceed with the code
			} else {
				vBoxCenter.RemoveAll()
				entryField.SetText("")
				entryField.SetPlaceHolder(placeHolderText2)
				vBoxCenter.Add(entryField)
				results, keys := pingotrace.DNSPTR(ctx, parsedInput)

				var ipAddr, host string
				for key, value := range results {
					if key == keys[0] {
						if pingotrace.CheckIPv4(key) {
							ipAddr = key
							host = value[0].(string)
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", ipAddr, host))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}

						} else {
							if value[len(value)-1] == false {
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetPlaceHolder(fmt.Sprintf("%s: %s", key, value[0].(string)))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)

								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							} else {
								host = key
								ipAddr = value[0].(string)
								vBoxCenter.RemoveAll()
								win.Resize(fyne.NewSize(980, 537))
								entryField = newTappableEntry("")
								entryField.SetText(fmt.Sprintf("Traceroute to %s [%s]:\n\n", host, ipAddr))
								entryField.SetMinRowsVisible(minRowVisible)
								vBoxCenter.Add(entryField)
								hBoxTop = container.NewHBox(btStopBack, layout.NewSpacer(), btDark, btLight)
								mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
								win.SetContent(mainBox)
								win.Resize(fyne.NewSize(980, 537))
							}

						}
						break
					}
				}
				tracerouteDst := entryField.Text
				if ipAddr != "" {
					maxHops := 30
					timeout := 1 * time.Second
					// Wrap the traceroute function to receive the output
					traceOutputChan1 := make(chan []string, maxHops)
					// Goroutine

					go func() {
						for {
							// Check if context has been cancelled before starting a new trace.
							select {
							case <-ctx.Done():
								return
							default:
								// If context not cancelled, continue to trace.
								pingotrace.Trace(ipAddr, maxHops, timeout, ctx, traceOutputChan1)
								time.Sleep(5 * time.Second)

								// Before updating the UI, check again if the context has been cancelled.
								select {
								case <-ctx.Done():
									return
								default:
									entryField.SetText("")
									entryField.Refresh()
									entryField.SetText(tracerouteDst)
								}
							}
						}
					}()

					// Start another goroutine to update the display as soon as data becomes available in the outputCh
					go func() {
						for {
							select {
							case line, ok := <-traceOutputChan1:
								if !ok {
									return
								}
								if !shouldUpdate {
									cancel() // cancel the context
									return
								}
								result := fmt.Sprintf("%2s\t", line[0])
								for i := 2; i < len(line); i++ {
									if line[i] == "*" {
										result += fmt.Sprintf("%-10s      \t", line[i]) // Add 6 more spaces after "*"
									} else {
										result += fmt.Sprintf("%-10s\t", line[i])
									}
								}
								result += fmt.Sprintf("%-45s", line[1])
								entryField.SetText(entryField.Text + result + "\n")
								entryField.Refresh() // Notify Fyne to repaint the widget
							case <-ctx.Done():
								// context canceled
								return
							}
						}
					}()
				}
			}
		}
	})

	btIPConfig = widget.NewButton("IP CONFIG", func() {
		resultText := pingotrace.IPConfig()
		entryField.SetText(resultText)
	})

	btMainClear = widget.NewButton("CLEAR", func() {
		entryField.SetText("")
		vBoxCenter.RemoveAll()
		entryField.SetPlaceHolder(placeHolderText1)
		entryField.SetMinRowsVisible(minRowVisible)
		vBoxCenter.Add(entryField)
		mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)
		win.SetContent(mainBox)
		win.Resize(fyne.NewSize(980, 537))
	})

	btLicense = widget.NewButton("LICENSE", func() {
		// Display license information or do something else
		licenseText := licenseText
		entryField.SetText("")
		entryField.SetPlaceHolder(string(licenseText))
	})

	setDarkMode := func() {
		customTheme.SetDark(true)
		fyneApp.Settings().SetTheme(customTheme)
	}

	btLight = widget.NewButton("LIGHT", func() {
		customTheme.SetDark(false)
		fyneApp.Settings().SetTheme(customTheme)
	})

	vBoxCenter.Add(entryField)
	btDark = widget.NewButton("DARK", setDarkMode)

	hBoxTop = container.NewHBox(btParser, btDNSPTRLookup, btDNSPTRtoIP, btPing, btTrace, btPinGoTrace, btContinuousTrace, btIPConfig, btMainClear, btLicense, layout.NewSpacer(), btDark, btLight)
	mainBox = container.NewBorder(hBoxTop, nil, nil, nil, vBoxCenter)

	win.SetContent(mainBox)
	win.Resize(fyne.NewSize(980, 537))
	setDarkMode()
	win.ShowAndRun()
}

type myTheme struct {
	myFont      fyne.Resource
	fontSize    float32
	darkColors  map[fyne.ThemeColorName]color.Color
	lightColors map[fyne.ThemeColorName]color.Color
	dark        bool
}

type fixedLayout struct {
	size fyne.Size
}

func (fl *fixedLayout) Layout(objects []fyne.CanvasObject, containerSize fyne.Size) {
	for _, object := range objects {
		object.Resize(fl.size)
	}
}

func (fl *fixedLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	return fl.size
}

func (m *myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if m.dark {
		if color, ok := m.darkColors[name]; ok {
			return color
		}
	} else {
		if color, ok := m.lightColors[name]; ok {
			return color
		}
	}
	return theme.DefaultTheme().Color(name, variant)
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return m.myFont
}

func (m *myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m *myTheme) Size(name fyne.ThemeSizeName) float32 {
	if name == theme.SizeNameText {
		return m.fontSize
	}
	return theme.DefaultTheme().Size(name)
}
func (m myTheme) BackgroundColor() color.Color {
	return m.darkColors[theme.ColorNameBackground]
}
func (m *myTheme) SetDark(dark bool) {
	m.dark = dark
}

func createTable(rows, cols int) *fyne.Container {
	table := container.NewGridWithColumns(cols)

	for i := 0; i < rows*cols; i++ {
		label := canvas.NewText("", color.White)
		label.Alignment = fyne.TextAlignLeading
		label.TextStyle.Bold = true
		table.Add(label)
	}

	return table
}

// contains checks if a string is present in a slice of strings.
func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

type tappableEntry struct {
	widget.Entry
	originalText string
}

func newTappableEntry(placeHolder string) *tappableEntry {
	newEntry := &tappableEntry{originalText: placeHolder}
	newEntry.ExtendBaseWidget(newEntry)
	newEntry.SetPlaceHolder(placeHolder)
	newEntry.SetMinRowsVisible(minRowVisible)
	newEntry.MultiLine = true
	return newEntry
}

func (e *tappableEntry) Tapped(ev *fyne.PointEvent) {
	if e.PlaceHolder == e.originalText {
		e.SetPlaceHolder(placeHolderText1)
	}
	e.Entry.Tapped(ev)
}
