package models_test

import (
	"fmt"
	"testing"

	"github.com/ossan-dev/coworkingapp/models"
)

func BenchmarkParsingJsonFile_V_1_24(b *testing.B) {
	roomNoToBench := [3]int{1, 999, 9999}
	for _, v := range roomNoToBench {
		rooms := make([]models.Room, 0, v)

		b.Run(fmt.Sprintf("ParseModelWithUnmarshal-%d-rooms", v), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := models.ParseModelWithUnmarshal(&rooms, fmt.Sprintf("../testdata/rooms_%d.json", v)); err != nil {
					b.Errorf("failed to run bench: %v", err)
				}
			}
		})

		b.Run(fmt.Sprintf("ParseModelWithDecoder-%d-rooms", v), func(b *testing.B) {
			b.ReportAllocs()
			for b.Loop() {
				if err := models.ParseModelWithDecoder(&rooms, fmt.Sprintf("../testdata/rooms_%d.json", v)); err != nil {
					b.Errorf("failed to run bench: %v", err)
				}
			}
		})
	}
}

func BenchmarkParsingJsonFile(b *testing.B) {
	roomNoToBench := [3]int{1, 999, 9999}
	for _, v := range roomNoToBench {
		rooms := make([]models.Room, 0, v)

		b.Run(fmt.Sprintf("ParseModelWithUnmarshal-%d-rooms", v), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := models.ParseModelWithUnmarshal(&rooms, fmt.Sprintf("../testdata/rooms_%d.json", v)); err != nil {
					b.Errorf("failed to run bench: %v", err)
				}
			}
		})

		b.Run(fmt.Sprintf("ParseModelWithDecoder-%d-rooms", v), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := models.ParseModelWithDecoder(&rooms, fmt.Sprintf("../testdata/rooms_%d.json", v)); err != nil {
					b.Errorf("failed to run bench: %v", err)
				}
			}
		})

	}
}
