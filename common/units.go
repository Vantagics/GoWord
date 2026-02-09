// Package common provides shared utilities, types, and constants for goword.
package common

// OOXML unit conversion constants.
const (
	// TwipsPerInch is the number of twips in one inch.
	TwipsPerInch = 1440
	// TwipsPerCm is the number of twips in one centimeter.
	TwipsPerCm = 567.0
	// TwipsPerPoint is the number of twips in one point.
	TwipsPerPoint = 20
	// EmuPerTwip is the number of EMUs (English Metric Units) in one twip.
	// 1 inch = 914400 EMU, 1 inch = 1440 twip => 1 twip = 635 EMU.
	EmuPerTwip = 635
	// EmuPerPixel is the number of EMUs per pixel at 96 DPI.
	// 1 inch = 914400 EMU, 96 DPI => 914400/96 = 9525 EMU/pixel.
	EmuPerPixel = 9525
)

// InchToTwip converts inches to twips.
func InchToTwip(inches float64) int {
	return int(inches * TwipsPerInch)
}

// CmToTwip converts centimeters to twips.
func CmToTwip(cm float64) int {
	return int(cm * TwipsPerCm)
}

// PointToTwip converts points to twips.
func PointToTwip(pt float64) int {
	return int(pt * TwipsPerPoint)
}

// TwipToInch converts twips to inches.
func TwipToInch(twip int) float64 {
	return float64(twip) / TwipsPerInch
}

// TwipToCm converts twips to centimeters.
func TwipToCm(twip int) float64 {
	return float64(twip) / TwipsPerCm
}

// TwipToPoint converts twips to points.
func TwipToPoint(twip int) float64 {
	return float64(twip) / TwipsPerPoint
}

// EmuToTwip converts EMUs (English Metric Units) to twips.
func EmuToTwip(emu int64) int {
	return int(emu / EmuPerTwip)
}

// TwipToEmu converts twips to EMUs.
func TwipToEmu(twip int) int64 {
	return int64(twip) * EmuPerTwip
}

// PointToHalfPoint converts points to half-points (used for font sizes in OOXML).
func PointToHalfPoint(pt float64) int {
	return int(pt * 2)
}

// HalfPointToPoint converts half-points to points.
func HalfPointToPoint(hp int) float64 {
	return float64(hp) / 2.0
}

// EmuToPixel converts EMUs to pixels (at 96 DPI).
func EmuToPixel(emu int64) int {
	return int(emu / EmuPerPixel)
}

// PixelToEmu converts pixels to EMUs (at 96 DPI).
func PixelToEmu(px int) int64 {
	return int64(px) * EmuPerPixel
}
