package services

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/color"
	"io"

	"github.com/skip2/go-qrcode"
)

// QRCodeFormat represents the format of a QR code
type QRCodeFormat string

const (
	// QRCodeFormatPNG represents the PNG format
	QRCodeFormatPNG QRCodeFormat = "png"
	// QRCodeFormatSVG represents the SVG format
	QRCodeFormatSVG QRCodeFormat = "svg"
)

// QRCodeOptions represents the options for a QR code
type QRCodeOptions struct {
	Size          int
	Foreground    color.Color
	Background    color.Color
	DisableBorder bool
}

// DefaultQRCodeOptions returns the default QR code options
func DefaultQRCodeOptions() *QRCodeOptions {
	return &QRCodeOptions{
		Size:          256,
		Foreground:    color.Black,
		Background:    color.White,
		DisableBorder: false,
	}
}

// QRCodeService handles QR code generation
type QRCodeService struct{}

// NewQRCodeService creates a new QR code service
func NewQRCodeService() *QRCodeService {
	return &QRCodeService{}
}

// Generate generates a QR code for the given URL
func (s *QRCodeService) Generate(url string, format QRCodeFormat, options *QRCodeOptions) ([]byte, error) {
	if options == nil {
		options = DefaultQRCodeOptions()
	}

	qr, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	qr.DisableBorder = options.DisableBorder
	qr.ForegroundColor = options.Foreground
	qr.BackgroundColor = options.Background

	switch format {
	case QRCodeFormatPNG:
		return qr.PNG(options.Size)
	case QRCodeFormatSVG:
		var buf bytes.Buffer
		if err := qr.Write(options.Size, &buf); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// GenerateBase64 generates a base64-encoded QR code for the given URL
func (s *QRCodeService) GenerateBase64(url string, format QRCodeFormat, options *QRCodeOptions) (string, error) {
	data, err := s.Generate(url, format, options)
	if err != nil {
		return "", err
	}

	var prefix string
	switch format {
	case QRCodeFormatPNG:
		prefix = "data:image/png;base64,"
	case QRCodeFormatSVG:
		prefix = "data:image/svg+xml;base64,"
	}

	return prefix + base64.StdEncoding.EncodeToString(data), nil
}

// WriteQRCode writes a QR code to the given writer
func (s *QRCodeService) WriteQRCode(w io.Writer, url string, format QRCodeFormat, options *QRCodeOptions) error {
	data, err := s.Generate(url, format, options)
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}
