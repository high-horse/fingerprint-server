package main

type CompareFingerprintRequest struct {
	ProbeImage string `json:"image1"`
	ReferenceImage string `json:"image2"`
}

type MatchRequest struct {
	ProbeImage     string `json:"image1"`     // base64 encoded image
	CandidateImage string `json:"image2"` // base64 encoded image
}

type CompareFingerprintResponse struct {
	Score float64 `json:"score"`
	Match bool `json:"is_match"`
	Confidence string `json:"confidence"`
	Details interface{} `json:"details"`
	Message string `json:"message"`
}


type MatchResponse struct {
	Score   float64 `json:"score"`
	Elapsed string  `json:"elapsed"`
	Error   string  `json:"error,omitempty"`
}