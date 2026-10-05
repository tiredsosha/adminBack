package web

import (
	"github.com/gin-gonic/gin"
	"github.com/tiredsosha/admin/protocols"
	config "github.com/tiredsosha/admin/tools/configurator"
	"github.com/tiredsosha/admin/tools/logger"
)

func resolumeNext(c *gin.Context) {
	var data JsonID

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("resolume next -", data)

	protocols.SendOscInt(config.FindPC(data.Zone, "ip"), 8010, "/composition/connectnextcolumn", 1)

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func resolumePrev(c *gin.Context) {
	var data JsonID

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("resolume prev -", data)

	protocols.SendOscInt(config.FindPC(data.Zone, "ip"), 8010, "/composition/connectprevcolumn", 1)

	c.JSON(200, gin.H{
		"message": "OK",
	})
}

func resolumeBlackout(c *gin.Context) {
	var data JsonID

	// Bind JSON and validate
	if err := c.ShouldBindJSON(&data); err != nil {
		logger.Error.Println("Invalid input:", err)
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	logger.Info.Println("resolume blackout -", data)

	protocols.SendOscInt(config.FindPC(data.Zone, "ip"), 8010, "/composition/bypassed", 1)

	c.JSON(200, gin.H{
		"message": "OK",
	})
}
