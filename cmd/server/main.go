package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/khaundadi/license-manager/pkg/auth"
)

func main() {
	// Initialize LicenseManager
	lm := auth.NewLicenseManager()

	// Create a sample license
	secret := []byte("a very secret key for AES256!!!!")
	encryptedKey, err := lm.CreateLicense("Alice", "MyProduct", "alice@example.com", time.Now().Add(30*24*time.Hour), secret) // expires in 30 days
	if err != nil {
		fmt.Println("Error creating license:", err)
	} else {
		fmt.Println("License created with encrypted key:", encryptedKey)
	}
	// license, err := lm.ValidateLicense(encryptedKey, secret)
	// if err != nil {
	// 	fmt.Println("Error validating license:", err)
	// } else {
	// 	fmt.Println("License is valid for:", license.User)
	// }
	r := gin.Default()

	// Endpoint to validate a license
	r.GET("/validate/:licenseKey", func(c *gin.Context) {
		licenseKey := c.Param("licenseKey")
		license, err := lm.ValidateLicense(licenseKey, secret)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"license": license})
	})

	// Start the server on port 8080
	r.Run(":8080")
}
