package main

import (
	"encoding/base64"
	"log"
	"strings"
	"time"
	"os"
	"os/exec"
	"fmt"

	"github.com/gofiber/fiber/v2"

)

func deleteFile(path string) {
	if err := os.Remove(path); err != nil {
		log.Printf("Failed to delete file %s: %v", path, err)
	} else {
		log.Printf("Deleted file: %s", path)
	}
}

func sanitizeImage(imagePath string) {
	// Build sanitized output path (overwrite same file)
	// cmd := exec.Command("convert", imagePath, "-colorspace", "Gray", "-density", "500", imagePath)
	cmd := exec.Command("convert", imagePath, imagePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("Failed to sanitize image %s: %v\nOutput: %s", imagePath, err, string(output))
	} else {
		log.Printf("Sanitized image: %s", imagePath)
	}
}

func storeImage(base64img string) (string, error) {
	ext := "png" // Default extension, can be changed based on image type
	if strings.HasPrefix(base64img, "data:") {
		parts := strings.SplitN(base64img,  "," , 2)
		if len(parts) != 2 {
			return "", fiber.NewError(fiber.StatusBadRequest, "Invalid base64 image format")
		}
		meta := parts[0]
		base64img = parts[1]

		if strings.Contains(meta, "image/jpeg") {
			ext = "jpg"
		} else if strings.Contains(meta, "image/png") {
			ext = "png"
		} else if strings.Contains(meta, "image/gif") {
			ext = "gif"
		} else {
			return "", fiber.NewError(fiber.StatusUnsupportedMediaType, "Unsupported image type")
		}
	}

	// decode Baase64 image
	decoded, err := base64.StdEncoding.DecodeString(base64img)
	if err != nil {
		return "", fiber.NewError(fiber.StatusBadRequest, "Failed to decode base64: "+err.Error())
	}

	// ensure dir exists
	if err := os.MkdirAll("temp", os.ModePerm); err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Failed to create temp directory: "+err.Error())
	}


	// create filename
	filename := fmt.Sprintf("image_%d.%s", time.Now().UnixNano(), ext)
	path := fmt.Sprintf("./temp/%s", filename)


		// Write file
	err = os.WriteFile(path, decoded, 0644)
	if err != nil {
		return "", fiber.NewError(fiber.StatusInternalServerError, "Failed to write image: "+err.Error())
		// return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to write image"})
	}

	return path, nil

}