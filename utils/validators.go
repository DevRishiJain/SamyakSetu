// All rights reserved Samyak-Setu

package utils

import (
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"
)

// ValidatePhone checks that the phone number is numeric and 10-13 digits.
func ValidatePhone(phone string) error {
	matched, _ := regexp.MatchString(`^\d{10,13}$`, phone)
	if !matched {
		return fmt.Errorf("phone must be numeric and 10-13 digits, got: %s", phone)
	}
	return nil
}

// ValidateLatitude checks that latitude is within valid range.
func ValidateLatitude(lat float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90, got: %f", lat)
	}
	return nil
}

// ValidateLongitude checks that longitude is within valid range.
func ValidateLongitude(lng float64) error {
	if lng < -180 || lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180, got: %f", lng)
	}
	return nil
}

// ValidateCoordinates checks both latitude and longitude.
func ValidateCoordinates(lat, lng float64) error {
	if err := ValidateLatitude(lat); err != nil {
		return err
	}
	return ValidateLongitude(lng)
}

// AllowedImageTypes contains the set of accepted image MIME types.
var AllowedImageTypes = map[string]bool{
	"image/jpeg": true,
	"image/jpg":  true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// MaxFileSize is the maximum allowed upload size (5 MB).
const MaxFileSize = 5 * 1024 * 1024

// ValidateImageFile checks that the uploaded file is an allowed image type and within size limits.
func ValidateImageFile(file *multipart.FileHeader) error {
	if file.Size > MaxFileSize {
		return fmt.Errorf("file size %d bytes exceeds maximum of 5MB", file.Size)
	}

	contentType := file.Header.Get("Content-Type")
	// Normalize content type
	contentType = strings.ToLower(strings.Split(contentType, ";")[0])

	if !AllowedImageTypes[contentType] {
		return fmt.Errorf("unsupported image type: %s (allowed: jpeg, png, webp, gif)", contentType)
	}

	return nil
}

// GetMimeType extracts and normalizes the MIME type from a file header.
func GetMimeType(file *multipart.FileHeader) string {
	ct := file.Header.Get("Content-Type")
	ct = strings.ToLower(strings.Split(ct, ";")[0])
	if ct == "image/jpg" {
		ct = "image/jpeg"
	}
	return ct
}

// ReadFileBytes reads the full contents of a multipart file into memory.
func ReadFileBytes(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}
