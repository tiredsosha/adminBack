package web

import (
	"github.com/gin-gonic/gin"
	"github.com/tiredsosha/admin/protocols"
	config "github.com/tiredsosha/admin/tools/configurator"
	"github.com/tiredsosha/admin/tools/formater"
	"github.com/tiredsosha/admin/tools/logger"
)

func restartPc(c *gin.Context) {
	var data JsonID

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("request data -", data)

	if data.Zone != "guards" && data.Zone != "ussr" {
		protocols.SendGet(formater.CustomStr(
			"http://{ip}:3001/restart",
			map[string]any{"ip": config.FindPC(data.Zone, "ip")}), 2,
		)
	}

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

// func restartPark(c *gin.Context) {
// 	var data JsonID

// 	// Bind JSON and validate
// 	if err := c.ShouldBindJSON(&data); err != nil {
// 		logger.Error.Println("Invalid input:", err)
// 		c.JSON(400, gin.H{"error": err.Error()})
// 		return
// 	}

// 	logger.Info.Println(data)

// 	go protocols.SendGet(formater.CustomStr(
// 		"http://{ip}:3001/restart",
// 		map[string]any{"ip": config.FindPC(data.Zone, "ip")}),
// 	)

// 	c.JSON(200, gin.H{
// 		"message": "OK",
// 	})
// }
