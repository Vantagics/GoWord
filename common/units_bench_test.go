package common

import "testing"

// Benchmark unit conversions to verify performance.

func BenchmarkInchToTwip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InchToTwip(8.5)
	}
}

func BenchmarkCmToTwip(b *testing.B) {
	for i := 0; i < b.N; i++ {
		CmToTwip(21.0)
	}
}

func BenchmarkPixelToEmu(b *testing.B) {
	for i := 0; i < b.N; i++ {
		PixelToEmu(1920)
	}
}

// Test that exported constants have expected values.
func TestConstants(t *testing.T) {
	if TwipsPerInch != 1440 {
		t.Errorf("TwipsPerInch = %d", TwipsPerInch)
	}
	if TwipsPerPoint != 20 {
		t.Errorf("TwipsPerPoint = %d", TwipsPerPoint)
	}
	if EmuPerTwip != 635 {
		t.Errorf("EmuPerTwip = %d", EmuPerTwip)
	}
	if EmuPerPixel != 9525 {
		t.Errorf("EmuPerPixel = %d", EmuPerPixel)
	}
}

// Test negative value conversions.
func TestNegativeConversions(t *testing.T) {
	if InchToTwip(-1.0) != -1440 {
		t.Errorf("InchToTwip(-1) = %d", InchToTwip(-1.0))
	}
	if TwipToInch(-1440) != -1.0 {
		t.Errorf("TwipToInch(-1440) = %f", TwipToInch(-1440))
	}
	if CmToTwip(-2.54) != -1440 {
		t.Errorf("CmToTwip(-2.54) = %d", CmToTwip(-2.54))
	}
}

// Test fractional point conversions.
func TestFractionalConversions(t *testing.T) {
	// 0.5 points = 10 twips
	if PointToTwip(0.5) != 10 {
		t.Errorf("PointToTwip(0.5) = %d", PointToTwip(0.5))
	}
	// 10.5 points = 21 half-points
	if PointToHalfPoint(10.5) != 21 {
		t.Errorf("PointToHalfPoint(10.5) = %d", PointToHalfPoint(10.5))
	}
	if HalfPointToPoint(21) != 10.5 {
		t.Errorf("HalfPointToPoint(21) = %f", HalfPointToPoint(21))
	}
}
