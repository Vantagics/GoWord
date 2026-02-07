package common

import (
	"math"
	"testing"
)

func TestInchToTwip(t *testing.T) {
	tests := []struct {
		inches float64
		want   int
	}{
		{1.0, 1440},
		{0.5, 720},
		{2.0, 2880},
		{0, 0},
	}
	for _, tt := range tests {
		got := InchToTwip(tt.inches)
		if got != tt.want {
			t.Errorf("InchToTwip(%v) = %d, want %d", tt.inches, got, tt.want)
		}
	}
}

func TestCmToTwip(t *testing.T) {
	got := CmToTwip(2.54) // ~1 inch
	// 2.54 cm * 567 = 1440.18 -> 1440
	if got != 1440 {
		t.Errorf("CmToTwip(2.54) = %d, want 1440", got)
	}
}

func TestPointToTwip(t *testing.T) {
	got := PointToTwip(12)
	if got != 240 {
		t.Errorf("PointToTwip(12) = %d, want 240", got)
	}
}

func TestTwipToInch(t *testing.T) {
	got := TwipToInch(1440)
	if got != 1.0 {
		t.Errorf("TwipToInch(1440) = %f, want 1.0", got)
	}
}

func TestTwipToCm(t *testing.T) {
	got := TwipToCm(567)
	if math.Abs(got-1.0) > 0.01 {
		t.Errorf("TwipToCm(567) = %f, want ~1.0", got)
	}
}

func TestTwipToPoint(t *testing.T) {
	got := TwipToPoint(240)
	if got != 12.0 {
		t.Errorf("TwipToPoint(240) = %f, want 12.0", got)
	}
}

func TestEmuConversions(t *testing.T) {
	twip := 1440
	emu := TwipToEmu(twip)
	back := EmuToTwip(emu)
	if back != twip {
		t.Errorf("TwipToEmu->EmuToTwip roundtrip: got %d, want %d", back, twip)
	}
}

func TestPointToHalfPoint(t *testing.T) {
	got := PointToHalfPoint(12)
	if got != 24 {
		t.Errorf("PointToHalfPoint(12) = %d, want 24", got)
	}
}

func TestHalfPointToPoint(t *testing.T) {
	got := HalfPointToPoint(24)
	if got != 12.0 {
		t.Errorf("HalfPointToPoint(24) = %f, want 12.0", got)
	}
}

func TestPixelEmuConversions(t *testing.T) {
	px := 100
	emu := PixelToEmu(px)
	back := EmuToPixel(emu)
	if back != px {
		t.Errorf("PixelToEmu->EmuToPixel roundtrip: got %d, want %d", back, px)
	}
}
