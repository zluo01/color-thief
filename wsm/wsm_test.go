package wsm

import (
	"color-thief/helper"
	"log"
	"reflect"
	"testing"
)

var (
	p1 [][3]int
)

func init() {
	var err error
	img1, err := helper.ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatal(err)
	}
	p1 = helper.SubsamplingPixelsFromImage(img1)

	if len(p1) != 300*225 {
		log.Fatal("Unexpected sample size found for photo1: ", len(p1))
	}
}

func BenchmarkGetHistogram(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = getHistogram(p1)
	}
}

func TestWSM(t *testing.T) {
	expected := [][3]int{
		{109, 210, 229},
		{56, 39, 29},
		{174, 185, 124},
		{89, 123, 119},
		{209, 226, 222},
		{202, 126, 31},
	}
	palette := WSM(p1, 6)
	for i := 0; i < 6; i++ {
		if !reflect.DeepEqual(palette[i], expected[i]) {
			t.Errorf("unequaled palette found, expected: %v, got %v", expected[i], palette[i])
		}
	}
}

func BenchmarkWSM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = WSM(p1, 6)
	}
}
