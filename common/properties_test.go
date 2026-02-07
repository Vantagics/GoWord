package common

import (
	"testing"
	"time"
)

func TestNewDocProperties(t *testing.T) {
	p := NewDocProperties()
	if p.Creator != "GoWord" {
		t.Errorf("creator = %q", p.Creator)
	}
	if p.LastModifiedBy != "GoWord" {
		t.Errorf("lastModifiedBy = %q", p.LastModifiedBy)
	}
	if p.Revision != 1 {
		t.Errorf("revision = %d", p.Revision)
	}
	if time.Since(p.Created) > time.Second {
		t.Error("created time too old")
	}
	if time.Since(p.Modified) > time.Second {
		t.Error("modified time too old")
	}
}

func TestZeroConversions(t *testing.T) {
	if InchToTwip(0) != 0 {
		t.Error("InchToTwip(0)")
	}
	if CmToTwip(0) != 0 {
		t.Error("CmToTwip(0)")
	}
	if PointToTwip(0) != 0 {
		t.Error("PointToTwip(0)")
	}
	if TwipToInch(0) != 0 {
		t.Error("TwipToInch(0)")
	}
	if TwipToCm(0) != 0 {
		t.Error("TwipToCm(0)")
	}
	if TwipToPoint(0) != 0 {
		t.Error("TwipToPoint(0)")
	}
	if EmuToTwip(0) != 0 {
		t.Error("EmuToTwip(0)")
	}
	if TwipToEmu(0) != 0 {
		t.Error("TwipToEmu(0)")
	}
	if PointToHalfPoint(0) != 0 {
		t.Error("PointToHalfPoint(0)")
	}
	if HalfPointToPoint(0) != 0 {
		t.Error("HalfPointToPoint(0)")
	}
	if EmuToPixel(0) != 0 {
		t.Error("EmuToPixel(0)")
	}
	if PixelToEmu(0) != 0 {
		t.Error("PixelToEmu(0)")
	}
}

func TestLargeConversions(t *testing.T) {
	// 10 inches
	if InchToTwip(10) != 14400 {
		t.Errorf("InchToTwip(10) = %d", InchToTwip(10))
	}
	// 72 points = 1 inch = 1440 twips
	if PointToTwip(72) != 1440 {
		t.Errorf("PointToTwip(72) = %d", PointToTwip(72))
	}
}
