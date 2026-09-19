package protocols

import (
	"context"
	"strings"
	"time"

	"github.com/tiredsosha/admin/tools/logger"
	"github.com/tiredsosha/gopjlink"
	"github.com/tiredsosha/pjlink"
)

// func SendPjlink(ip, command string) {
// 	defer func() {
// 		if r := recover(); r != nil {
// 			// Log the panic or handle it as needed
// 			logger.Error.Printf("Recovered from panic in GetPjlink: %v", r)
// 		}
// 	}()

// 	proj := pjlink.NewProjector(ip, "")
// 	// fmt.Println(ip, command)

// 	switch command {
// 	case "on":
// 		if err := proj.TurnOn(); err != nil {
// 			logger.Error.Printf("couldn't send execute pjlink %v", err)
// 		}
// 	case "off":
// 		if err := proj.TurnOff(); err != nil {
// 			logger.Error.Printf("couldn't send execute pjlink %v", err)
// 		}
// 	}

// }

func SendPjlink(ip, command string) {
	var boolCommand bool

	if command == "on" {
		boolCommand = true
	} else {
		boolCommand = false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// Initialize your projector object with connection details
	proj := gopjlink.NewProjector(ip) // or however the constructor is defined

	// Send power off command
	err := proj.SetPower(ctx, boolCommand)
	if err != nil {
		logger.Error.Printf("couldn't send execute pjlink %v", err)
	}
}

// func GetPjlink(ip string) int {
// 	response := 520
// 	proj := pjlink.NewProjector(ip, "")
// 	status, err := proj.GetPowerStatus()
// 	if err != nil {
// 		logger.Error.Printf("couldn't send execute pjlink %v", err)
// 	} else {
// 		if len(status.Response) > 0 {
// 			//logger.Info.Println("pjlink status -", status.Response)
// 			boolStatus, _ := strconv.ParseBool(status.Response[0])
// 			if boolStatus {
// 				response = 200
// 			} else {
// 				response = 521
// 			}
// 		}
// 	}
// 	logger.Info.Println("pjlink status -", response)
// 	return response
// }

func GetPjlink(ip string) (response int) {
	response = 520

	defer func() {
		if r := recover(); r != nil {
			logger.Error.Printf("Recovered from panic in GetPjlink: %v", r)
			response = 520
		}
	}()

	proj := pjlink.NewProjector(ip, "")

	status, err := proj.GetPowerStatus()
	if err != nil {
		logger.Error.Printf("couldn't execute pjlink for %s: %v", ip, err)
		return
	}

	if status == nil || len(status.Response) == 0 {
		logger.Error.Printf("Empty PJLink response from %s", ip)
		return
	}

	response = pjlinkPowerStatus(status.Response)
	if response == 520 {
		logger.Error.Printf(
			"Unknown PJLink power status from %s: %q",
			ip,
			status.Response[0],
		)
	}

	return
}

func pjlinkPowerStatus(values []string) int {
	if len(values) == 0 {
		return 520
	}

	status := strings.Trim(values[0], "\x00 \t\r\n")

	switch status {
	case "1":
		// ON
		return 200

	case "3":
		// Warming up
		return 200

	case "0":
		// Standby / OFF
		return 521

	case "2":
		// Cooling
		return 521
	}

	return 520
}
