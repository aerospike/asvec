package writers

import (
	"asvec/utils"
	"testing"

	"github.com/aerospike/avs-client-go/protos"
)

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

func Test_calculateIndexSizeDetailed(t *testing.T) {
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
					Params: &protos.IndexDefinition_HnswParams{
						HnswParams: &protos.HnswParams{
							M: utils.Ptr(uint32(10)),
						},
					},
				},
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 100,
				},
			},
			want: 84360,
		},
		{
			name: "zero vertices",
			args: args{
				index: &protos.IndexDefinition{
					Dimensions: 100,
					Params: &protos.IndexDefinition_HnswParams{
						HnswParams: &protos.HnswParams{
							M: utils.Ptr(uint32(10)),
						},
					},
				},
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 0,
				},
			},
			want: 0,
		},
		{
			name: "different dimension and vertices",
			args: args{
				index: &protos.IndexDefinition{
					Dimensions: 200,
					Params: &protos.IndexDefinition_HnswParams{
						HnswParams: &protos.HnswParams{
							M: utils.Ptr(uint32(15)),
						},
					},
				},
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 800_000,
				},
			},
			want: 1122854710,
		},
		{
			name: "different m value",
			args: args{
				index: &protos.IndexDefinition{
					Dimensions: 100,
					Params: &protos.IndexDefinition_HnswParams{
						HnswParams: &protos.HnswParams{
							M: utils.Ptr(uint32(20)),
						},
					},
				},
				status: &protos.IndexStatusResponse{
					IndexHealerVerticesValid: 10,
				},
			},
			want: 10600,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateIndexSize(tt.args.index, tt.args.status); got != tt.want {
				t.Errorf("calculateIndexSizeDetailed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_calculateTotalGraphNodes(t *testing.T) {
	type args struct {
		m                int64
		numValidVertices int64
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "simple case",
			args: args{
				m:                2,
				numValidVertices: 8,
			},
			want: 15,
		},
		{
			name: "m is 1",
			args: args{
				m:                1,
				numValidVertices: 10,
			},
			want: 10,
		},
		{
			name: "numValidVertices is 0",
			args: args{
				m:                2,
				numValidVertices: 0,
			},
			want: 0,
		},
		{
			name: "m is greater than numValidVertices",
			args: args{
				m:                10,
				numValidVertices: 5,
			},
			want: 5,
		},
		{
			name: "large values",
			args: args{
				m:                3,
				numValidVertices: 1000,
			},
			want: 1498,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTotalGraphNodes(tt.args.m, tt.args.numValidVertices); got != tt.want {
				t.Errorf("calculateTotalGraphNodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
