package main

import (
	"log"
	"time"
	"os"
	"io"
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

	return rotateLogs, nil

    // Optionally, log to both stdout and the file.
    // mw := io.MultiWriter(os.Stdout, rotateLogs)
    // return mw, nil
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
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
    }))
	app.Use(cors.New())

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		log.Println("Health check called") 
		return c.JSON(fiber.Map{
			"status": "ok",
			"time":   time.Now(),
		})
	})

	// Fingerprint matching endpoint
	app.Post("/compare-fingerprints", matchFingerprints)

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

	var match bool
	var confidence string

	switch {
	case score > 300:
		match = true
		confidence = "High"
	case score > 199 && score <= 300:
		match = true
		confidence = "Medium"
	case score <= 199:
		match = false
		confidence = "Low"
	default:
		match = false
		confidence = "None"
	}

	// prepare response 
	response := CompareFingerprintResponse{
		StdOutput: CompareFingerprintResponseV2{
			Score:      score,
			Match:      match, 
			Confidence: confidence,
			Details:    map[string]interface{}{
				"elapsed_tile": time.Since(start).String(),
			},
			Message:    "Fingerprint comparison completed successfully",
		},
	}
	return c.JSON(response)
}

