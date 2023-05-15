package tracker

import (
	"reflect"
	"testing"
)

func Test_blockRange_split(t *testing.T) {
	tests := []struct {
		name  string
		r     *blockRange
		want  *blockRange
		want1 *blockRange
	}{
		{
			r:     newBlockRange(1, 100),
			want:  newBlockRange(1, 50),
			want1: newBlockRange(51, 100),
		},
		{
			r:     newBlockRange(1, 101),
			want:  newBlockRange(1, 51),
			want1: newBlockRange(52, 101),
		},
		{
			r:     newBlockRange(3, 4),
			want:  newBlockRange(3, 3),
			want1: newBlockRange(4, 4),
		},
		{
			r:     newBlockRange(2, 3),
			want:  newBlockRange(2, 2),
			want1: newBlockRange(3, 3),
		},
	}
	for _, tt := range tests {
		name := tt.r.String()
		t.Run(name, func(t *testing.T) {
			got, got1 := tt.r.split()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("blockRange.split() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("blockRange.split() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
