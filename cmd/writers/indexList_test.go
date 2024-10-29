package writers

import (
	"testing"

	"github.com/aerospike/avs-client-go/protos"
)

func Test_calculateIndexSize(t *testing.T) {
	type args struct {
		index  *protos.IndexDefinition
		status *protos.IndexStatusResponse
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "positive simple",
			args: args{
				index: &protos.IndexDefinition{
					Dimensions: 100,
				},
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 10,
				},
			},
			want: 9000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateIndexSize(tt.args.index, tt.args.status); got != tt.want {
				t.Errorf("calculateIndexSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatBytes(t *testing.T) {
	type args struct {
		bytes int64
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "petabytes",
			args: args{
				bytes: 1024 * 1024 * 1024 * 1024 * 1024,
			},
			want: "1.00 PB",
		},
		{
			name: "terabytes",
			args: args{
				bytes: 1024 * 1024 * 1024 * 1024,
			},
			want: "1.00 TB",
		},
		{
			name: "gigabytes",
			args: args{
				bytes: 1024 * 1024 * 1024,
			},
			want: "1.00 GB",
		},
		{
			name: "megabytes",
			args: args{
				bytes: 1024 * 1024,
			},
			want: "1.00 MB",
		},
		{
			name: "kilobytes",
			args: args{
				bytes: 1024,
			},
			want: "1.00 KB",
		},
		{
			name: "bytes",
			args: args{
				bytes: 512,
			},
			want: "512 B",
		},
		{
			name: "zero bytes",
			args: args{
				bytes: 0,
			},
			want: "0 B",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.args.bytes); got != tt.want {
				t.Errorf("formatBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_getPercentUnmerged(t *testing.T) {
	type args struct {
		status *protos.IndexStatusResponse
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zero vertices",
			args: args{
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 0,
					UnmergedRecordCount:      10,
				},
			},
			want: "0%",
		},
		{
			name: "zero unmerged records",
			args: args{
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 100,
					UnmergedRecordCount:      0,
				},
			},
			want: "0.00%",
		},
		{
			name: "50 percent unmerged",
			args: args{
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 100,
					UnmergedRecordCount:      50,
				},
			},
			want: "50.00%",
		},
		{
			name: "100 percent unmerged",
			args: args{
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 100,
					UnmergedRecordCount:      100,
				},
			},
			want: "100.00%",
		},
		{
			name: "33.33 percent unmerged",
			args: args{
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 300,
					UnmergedRecordCount:      100,
				},
			},
			want: "33.33%",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getPercentUnmerged(tt.args.status); got != tt.want {
				t.Errorf("getPercentUnmerged() = %v, want %v", got, tt.want)
			}
		})
	}
}
