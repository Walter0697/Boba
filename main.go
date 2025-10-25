package main

import (
	"fmt"
	"os"
	
	"boba/internal/ui"
)

func main() {
	uiManager := ui.NewUIManager()
	if err := uiManager.Start(); err != nil {
		fmt.Printf("Error running application: %v\n", err)
		os.Exit(1)
	}
}