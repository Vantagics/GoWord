// Package common provides shared utilities, types, and constants for goword.
package common

// Twip is the base unit in OOXML: 1 twip = 1/1440 inch.

// InchToTwip converts inches to twips.
func InchToTwip(inches float64) int {
	return int(inches * 1440)
}

// CmToTwip converts centimeters to twips.
func CmToTwip(cm float64) int {
	return int(cm * 567.0)
}

// PointToTwip converts points to twips.
func PointToTwip(pt float64) int {
	return int(pt * 20)
}

// TwipToInch converts twips to inches.
func TwipToInch(twip int) float64 {
	return float64(twip) / 1440.0
}

// TwipToCm converts twips to centimeters.
func TwipToCm(twip int) float64 {
	return float64(twip) / 567.0
}

// TwipToPoint converts twips to points.
func TwipToPoint(twip int) float64 {
	return float64(twip) / 20.0
}

// EmuToTwip converts EMUs (English Metric Units) to twips.
// 1 inch = 914400 EMU, 1 inch = 1440 twip => 1 twip = 635 EMU.
func EmuToTwip(emu int64) int {
	return int(emu / 635)
}

// TwipToEmu converts twips to EMUs.
func TwipToEmu(twip int) int64 {
	return int64(twip) * 635
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
	return int(emu / 9525)
}

// PixelToEmu converts pixels to EMUs (at 96 DPI).
func PixelToEmu(px int) int64 {
	return int64(px) * 9525
}
