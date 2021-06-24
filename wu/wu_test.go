package wu

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
	"log"
	"testing"
)

var p [][]int

func init() {
	img1, err := rgbUtil.ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	p = helper.SubsamplingPixels(img1)
}

func TestQuantWu(t *testing.T) {
	expected := []string{"36241b", "aebc6f", "6ccee1", "6b7063", "cededf", "d67818"}
	palette := QuantWu(p, 6)
	for i := 0; i < 6; i++ {
		if palette[i] != expected[i] {
			t.Errorf("unequaled palette found, expected: %s, got %s", expected[i], palette[i])
		}
	}
}

func BenchmarkQuantWu(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = QuantWu(p, 6)
	}
}
