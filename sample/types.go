package main

type CompareFingerprintRequest struct {
	ProbeImage string `json:"image1"`
	ReferenceImage string `json:"image2"`
}

type CompareFingerprintResponse struct {
	Score float64 `json:"score"`
	Match bool `json:"is_match"`
	Confidence string `json:"confidence"`
	Details interface{} `json:"details"`
	Message string `json:"message"`
}