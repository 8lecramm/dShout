package main

import (
	"fmt"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func CreateWindow() fyne.Window {

	myApp := app.New()
	myApp.Settings().SetTheme(theme.DarkTheme())

	myWindow := myApp.NewWindow("dShout")
	myWindow.Resize(fyne.NewSize(800, 300))
	myWindow.SetFixedSize(true)

	// input fields
	in_wallets := widget.NewMultiLineEntry()
	in_message := widget.NewMultiLineEntry()
	in_message.SetMinRowsVisible(5)

	// output fields
	output := widget.NewEntry()

	// ringsize dropdown
	rs_options := []string{"2", "4", "8", "16", "32", "64"}
	ringsize := widget.NewSelect(rs_options, nil)
	ringsize.SetSelectedIndex(3)

	// buttons
	button := widget.NewButton("Generate output", func() {

		in_wallets.FocusLost()
		in_message.FocusLost()

		addrs := strings.Split(in_wallets.Text, "\n")
		addrs = ValidateReceivers(addrs)

		if len(addrs) == 0 {
			output.SetText("no (valid) receivers")
			output.FocusGained()
			return
		}
		if len(in_message.Text) < MSG_INPUT {
			output.SetText("no message")
			output.FocusGained()
			return
		}

		p, keys, key, err := GenerateSharedSecrets(addrs)
		if err != nil {
			output.SetText(err.Error())
		} else {
			var key_string string

			for i := range keys {
				key_string += keys[i]
			}
			enc, msg, err := EncryptMessage(in_message.Text, key)
			if err != nil {
				output.SetText(err.Error())
			} else {
				in_message.Text = msg
				in_message.Refresh()
				output.SetText(fmt.Sprintf("%s%sx%s", p, key_string, enc))
			}
		}
		output.FocusGained()
	})
	button2 := widget.NewButton("Send to SC", func() {
		output.FocusLost()
		if len(output.Text) >= MSG_MIN_LENGTH {
			if _, err := SC_SendMessage(output.Text, ringsize.Selected); err == nil {
				// TODO: Transfer response is empty
				//output.Text = fmt.Sprintf("TXID: %s", txid)
				output.Text = "Transaction sent!"
			} else {
				output.Text = fmt.Sprintf("Error: %s", err.Error())
			}
			output.Refresh()
			output.FocusGained()
		}
	})

	button3 := widget.NewButton("Check for messages", func() {
		count, _ := SC_SyncLoop()

		if count > 0 {
			newMessagesContent := widget.NewLabel(fmt.Sprintf("Found %d message(s)!", count))
			dialog.ShowCustom("New Message", "Got it!", newMessagesContent, myWindow)
		} else {
			dialog.ShowInformation("Message", "No new message", myWindow)
		}
	})
	button4 := widget.NewButton("Show messages", func() {
		if len(decrypted_messages) > 0 {
			MessageWindow(myApp)
		}
	})

	// container
	content := container.NewVBox(
		widget.NewLabel("Receiver:"),
		in_wallets,
		widget.NewLabel("Message"),
		in_message,
		widget.NewLabel("Output"),
		output,
		container.NewHBox(
			button,
			button2,
			ringsize,
			layout.NewSpacer(),
			button3,
			button4,
		),
	)

	myWindow.SetContent(content)

	return myWindow
}

// new window wo view messages
func MessageWindow(app fyne.App) {

	myMessageWindow := app.NewWindow("dShout - Messages")
	myMessageWindow.Resize(fyne.NewSize(600, 200))
	myMessageWindow.SetFixedSize(true)

	message := widget.NewMultiLineEntry()
	message.SetMinRowsVisible(6)
	block := widget.NewEntry()

	sort.Slice(decrypted_messages, func(i, j int) bool { return decrypted_messages[i].Block < decrypted_messages[j].Block })

	message.Text = decrypted_messages[0].Message
	block.Text = fmt.Sprintf("%d (%v)", decrypted_messages[0].Block, decrypted_messages[0].Time)
	message.Refresh()

	var pos int
	btn_prev := widget.NewButton("Prev", func() {
		if pos > 0 {
			pos--
			message.Text = decrypted_messages[pos].Message
			block.Text = fmt.Sprintf("%d (%v)", decrypted_messages[pos].Block, decrypted_messages[pos].Time)
			message.Refresh()
			block.Refresh()
		}
	})
	btn_next := widget.NewButton("Next", func() {
		if pos < len(decrypted_messages)-1 {
			pos++
			message.Text = decrypted_messages[pos].Message
			block.Text = fmt.Sprintf("%d (%v)", decrypted_messages[pos].Block, decrypted_messages[pos].Time)
			message.Refresh()
			block.Refresh()
		}
	})
	btn_close := widget.NewButton("Close", func() {
		myMessageWindow.Close()
	})

	content := container.NewVBox(
		widget.NewLabel("Block"),
		block,
		widget.NewLabel("Message"),
		message,
		container.NewHBox(
			btn_prev,
			btn_next,
			layout.NewSpacer(),
			btn_close,
		),
	)

	myMessageWindow.SetContent(content)
	myMessageWindow.Show()
}
