package main

import (
	"fmt"
	"os"

	"github.com/skabbio1976/osupgrader-gui/internal/gui"
	"github.com/spf13/pflag"
)

var (
	version = "dev"
	commit  = "unknown"
)

func main() {
	debugFlag := pflag.BoolP("debug", "d", false, "Aktivera debug-loggning (skriver debuglogg.txt)")
	mockFlag := pflag.Bool("mock", false, "Kör i mock-läge med simulerade VMs")
	versionFlag := pflag.BoolP("version", "v", false, "Visa versionsinformation och avsluta")

	pflag.Parse()

	if *versionFlag {
		fmt.Printf("OSUpgrader GUI %s (%s)\n", version, commit)
		return
	}

	app := gui.NewApp(*debugFlag, *mockFlag)
	if app == nil {
		fmt.Fprintln(os.Stderr, "kunde inte skapa GUI-applikationen")
		os.Exit(1)
	}

	app.Run()
}
