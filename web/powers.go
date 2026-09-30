package web

import (
	"github.com/gin-gonic/gin"
	"github.com/tiredsosha/admin/protocols"
	"github.com/tiredsosha/admin/tools/formater"
	"github.com/tiredsosha/admin/tools/logger"

	config "github.com/tiredsosha/admin/tools/configurator"
)

func powerPc(c *gin.Context) {
	var data JsonCommand

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	if data.Command == "on" {
		protocols.SendWOL(config.FindPC(data.Zone, "mac"))
	} else {
		protocols.SendGet(formater.CustomStr(
			"http://{ip}:3001/off",
			map[string]any{"ip": config.FindPC(data.Zone, "ip")}), 2,
		)
	}

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func powerProjector(c *gin.Context) {
	var data JsonCommand

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	protocols.SendPjlink(config.FindPJ(data.Zone, data.ID), data.Command)

	// fmt.Println(data.Zone)

	// вот это тока для рязани исключение, в других проектах надо его убирать
	if data.Zone == "vynil" {
		command := "0"
		if data.Command == "on" {
			command = "n"
		} else {
			command = "f"
		}
		zoneRelay := config.FindRelay(data.Zone)
		for _, ip := range zoneRelay {
			protocols.SendGet(formater.CustomStr(
				"http://admin:admin@{ip}/protect/rb0{command}.cgi",
				map[string]any{"ip": ip, "command": command},
			), 2,
			)
		}

	}

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func powerRelay(c *gin.Context) {
	var data JsonCommand
	command := "0"

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	if data.Command == "on" {
		command = "n"
	} else {
		command = "f"
	}
	zoneRelay := config.FindRelay(data.Zone)
	for _, ip := range zoneRelay {
		protocols.SendGet(formater.CustomStr(
			"http://admin:admin@{ip}/protect/rb0{command}.cgi",
			map[string]any{"ip": ip, "command": command},
		), 2,
		)
	}

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func powerZone(c *gin.Context) {
	var data JsonCommand
	// command := "0"

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	if data.Zone != "guards" && data.Zone != "ussr" {
		if data.Command == "on" {
			// command = "n"
			for range 4 {
				protocols.SendWOL(config.FindPC(data.Zone, "mac"))
			}
		} else {
			// command = "f"
			protocols.SendGet(formater.CustomStr(
				"http://{ip}:3001/off",
				map[string]any{"ip": config.FindPC(data.Zone, "ip")}), 2,
			)
		}
	}

	switch data.Command {
	case "on":
		switch data.Zone {
		case "guards":
			protocols.SendUdp("10.1.5.24", 8010, "start")
		case "ussr":
			protocols.SendUdp("10.1.6.21", 8010, "start")
		}
	case "off":
		switch data.Zone {
		case "guards":
			protocols.SendUdp("10.1.5.24", 8010, "stop")
		case "ussr":
			protocols.SendUdp("10.1.6.21", 8010, "stop")
		}
	}

	zonePJ := config.FindZonePJ(data.Zone)
	for _, ip := range zonePJ {
		protocols.SendPjlink(ip, data.Command)
	}

	// zoneRelay := config.FindRelay(data.Zone)
	// for _, ip := range zoneRelay {
	// 	protocols.SendGet(formater.CustomStr(
	// 		"http://admin:admin@{ip}/protect/rb0{command}.cgi",
	// 		map[string]any{"ip": ip, "command": command},
	// 	), 2,
	// 	)
	// }

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func powerPark(c *gin.Context) {
	var data JsonNoID

	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	switch data.Command {
	case "on":

		// command := "n"

		// Wake MACs
		go func() {
			for _, mac := range config.ALLMAC {
				for i := 0; i < 4; i++ {
					protocols.SendWOL(mac)
				}
			}
		}()

		// Turn on projectors
		go func() {
			for _, pj := range config.ALLPJ {
				protocols.SendPjlink(pj, "on")
			}
		}()

		// Turn on relay
		// go func() {
		// 	protocols.SendGet(formater.CustomStr(
		// 		"http://admin:admin@{ip}/protect/rb0{command}.cgi",
		// 		map[string]any{"ip": "172.16.3.76", "command": command},
		// 	), 2,
		// 	)
		// }()

		// // Turn defaults on lights
		// go func() {
		// 	for _, dali := range config.ALLDALI {
		// 		out, lamps, defaults := config.FindDali(dali)
		// 		protocols.ArlightDefault(out, lamps, defaults)
		// 	}
		// }()

		go func() {
			for i := 0; i < 2; i++ {
				protocols.SendUdp("10.1.5.24", 8010, "start")
				protocols.SendUdp("10.1.6.21", 8010, "start")
			}
		}()

	case "off":

		// command := "f"

		// Power off PCs by IP
		go func() {
			for _, ip := range config.ALLPC {
				url := formater.CustomStr("http://{ip}:3001/off", map[string]any{"ip": ip})
				protocols.SendGet(url, 2)
			}
		}()

		// Turn off projectors
		go func() {
			for _, pj := range config.ALLPJ {
				protocols.SendPjlink(pj, "off")
			}
		}()

		// Turn off relay
		// go func() {
		// 	protocols.SendGet(formater.CustomStr(
		// 		"http://admin:admin@{ip}/protect/rb0{command}.cgi",
		// 		map[string]any{"ip": "172.16.3.76", "command": command},
		// 	), 2,
		// 	)
		// }()

		go func() {
			for i := 0; i < 2; i++ {
				protocols.SendUdp("10.1.5.24", 8010, "stop")
				protocols.SendUdp("10.1.6.21", 8010, "stop")
			}
		}()

	case "restart":
		// Restart PCs by IP
		go func() {
			for _, ip := range config.ALLPC {
				url := formater.CustomStr("http://{ip}:3001/restart", map[string]any{"ip": ip})
				protocols.SendGet(url, 2)
			}
		}()

	default:
		logger.Warn.Printf("unknown command for park power: %s", data.Command)
	}

	c.JSON(200, gin.H{"message": "OK"})
}
