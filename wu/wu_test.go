package wu

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
	"log"
	"reflect"
	"testing"
)

var p [][]int

func init() {
	img, err := rgbUtil.ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatal(err)
	}
	p = helper.SubsamplingPixels(img)
}

func TestQuantWu(t *testing.T) {
	expected := [][]int{
		{108, 206, 225},
		{54, 36, 27},
		{174, 188, 111},
		{107, 112, 99},
		{206, 222, 223},
		{214, 120, 24},
	}
	palette, _ := QuantWu(p, 6)
	for i := 0; i < 6; i++ {
		if !reflect.DeepEqual(palette[i], expected[i]) {
			t.Errorf("unequaled palette found, expected: %v, got %v", expected[i], palette[i])
		}
	}
}

func BenchmarkQuantWu(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = QuantWu(p, 6)
	}
}
