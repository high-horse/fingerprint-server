package main

import (
	"encoding/base64"
	"log"
	"strings"
	"time"
	"os"
	"io"
	"os/exec"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"

	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
)

// setupLogger configures the file-rotatelogs output so that the log file follows
// the pattern "logs/finger_yyyy-mm-dd.log" and rotates every 24 hours
func setupLogger() (io.Writer, error) {
    // Ensure the logs directory exists.
    if err := os.MkdirAll("logs", 0755); err != nil {
        return nil, fmt.Errorf("failed to create logs directory: %v", err)
    }

    // The log file name pattern. %Y, %m, and %d will be replaced with the current year, month, and day.
    rotateLogs, err := rotatelogs.New(
        "logs/finger_%Y-%m-%d.log",
        rotatelogs.WithRotationTime(24*time.Hour), // rotate every 24 hours
        rotatelogs.WithMaxAge(7*24*time.Hour),       // keep logs for 7 days (optional)
    )
    if err != nil {
        return nil, fmt.Errorf("failed to initialize file rotatelogs: %v", err)
    }

    // Optionally, log to both stdout and the file.
    mw := io.MultiWriter(os.Stdout, rotateLogs)
    return mw, nil
}


func main() {
	logOutput, err := setupLogger()
	if err != nil {
        log.Fatalf("Error setting up logger: %v", err)
    }

    // Set the log flags to include date, time, and short file information.
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// Direct Go's default logger output to our log writer.
    log.SetOutput(logOutput)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(MatchResponse{
				Error: err.Error(),
			})
		},
	})

	// Middleware
	// app.Use(logger.New())
	// Use middleware logging for Fiber, directing logs to our log output.
    app.Use(fiberLogger.New(fiberLogger.Config{
        Output: logOutput,
    }))
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Fingerprint matching endpoint
	app.Post("/match", matchFingerprints)

	// Start server
	log.Println("Server starting on :9090")
	log.Fatal(app.Listen(":9090"))

}


func matchFingerprints(c *fiber.Ctx)error {
	start := time.Now()

	var req MatchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MatchResponse{
			Error: "Invalid request body: " + err.Error(),
		})
	}

	// req.ProbeImage = c.FormValue("probe_image")
	// req.CandidateImage = c.FormValue("candidate_image")

	// Validate input
	if req.ProbeImage == "" || req.CandidateImage == "" {
		return c.Status(fiber.StatusBadRequest).JSON(MatchResponse{
			Error: "Both probe_image and candidate_image are required",
		})
	}

	probefile, err := storeImage(req.ProbeImage)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MatchResponse{
			Error: "Failed to store probe image: " + err.Error(),
		})
	}
	candidatefile, err := storeImage(req.CandidateImage)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(MatchResponse{
			Error: "Failed to store candidate image: " + err.Error(),
		})
	}

	sanitizeImage(probefile)
	sanitizeImage(candidatefile)

	defer deleteFile(probefile)
	defer deleteFile(candidatefile)


	log.Println("Probe image stored at:", probefile)
	log.Println("Candidate image stored at:", candidatefile)

	score, err := compareFingerprint(probefile, candidatefile)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(MatchResponse{
			Error: "Failed to compare fingerprints: " + err.Error(),
		})
	}
	log.Println("Fingerprint comparison score:", score)
	// Prepare response
	response := MatchResponse{
		Score:   score,
		Elapsed: time.Since(start).String(),
	}
	if score > 0.5 {
		response.Error = "Match found with score: " + fmt.Sprintf("%.2f", score)
	} else {
		response.Error = "No match found, score: " + fmt.Sprintf("%.2f", score)
	}
	log.Println("Response:", response)
	
	return c.JSON(response)
}


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