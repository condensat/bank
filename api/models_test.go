package api

import (
	"reflect"
	"testing"
)

func TestModels(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"default", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Models(); !reflect.DeepEqual(len(got), tt.want) {
				t.Errorf("Models() = %v, want %v", len(got), tt.want)
			}
		})
	}
}