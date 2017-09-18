package main

import (
	"os"
	"sort"
	"testing"

	"github.com/posener/complete"
)

func TestPredictions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		predictor complete.Predictor
		last      string
		want      []string
	}{
		{
			name:      "predict tests ok",
			predictor: predictTest,
			want:      []string{"TestPredictions", "Example"},
		},
		{
			name:      "predict tests not found",
			predictor: predictTest,
			last:      "X",
		},
		{
			name:      "predict benchmark ok",
			predictor: predictBenchmark,
			want:      []string{"BenchmarkFake"},
		},
		{
			name:      "predict benchmarks not found",
			predictor: predictBenchmark,
			last:      "X",
		},
		{
			name:      "predict local ok",
			predictor: complete.PredictFunc(predictPackages),
			last:      ".",
			want:      []string{"./"},
		},
		{
			name:      "predict system ok",
			predictor: complete.PredictFunc(predictPackages),
			last:      "github.com/posener/complete/goc",
			want:      []string{"github.com/posener/complete/gocomplete/"},
		},
		{
			name:      "predict packages not found",
			predictor: complete.PredictFunc(predictPackages),
			last:      "X",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := complete.Args{Last: tt.last}
			got := tt.predictor.Predict(a)
			if !equal(got, tt.want) {
				t.Errorf("Failed %s: got: %q, want: %q", t.Name(), got, tt.want)
			}
		})
	}
}

func BenchmarkFake(b *testing.B) {}

func Example() {
	os.Setenv("COMP_LINE", "go ru")
	main()
	// output: run

}

func equal(s1, s2 []string) bool {
	sort.Strings(s1)
	sort.Strings(s2)
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
