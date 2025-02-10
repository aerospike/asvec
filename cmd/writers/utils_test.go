package writers

import (
	"reflect"
	"testing"

	"github.com/aerospike/avs-client-go/protos"
)

func Test_formatRole(t *testing.T) {
	type args struct {
		role protos.NodeRole
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "TestRolePrimary",
			args: args{role: protos.NodeRole_INDEX_QUERY},
			want: "INDEX_QUERY",
		},
		{
			name: "TestRoleSecondary",
			//nolint:staticcheck // This value is deprecated but is still returned by the server as a default node role
			args: args{role: protos.NodeRole_INDEX_UPDATE},
			want: "INDEX_UPDATE",
		},
		{
			name: "TestRoleUnknown",
			args: args{role: protos.NodeRole_KV_READ},
			want: "KV_READ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRole(tt.args.role); got != tt.want {
				t.Errorf("formatRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_formatRoles(t *testing.T) {
	type args struct {
		roles []protos.NodeRole
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "TestMultipleRoles",
			args: args{
				roles: []protos.NodeRole{
					protos.NodeRole_INDEX_QUERY,
					//nolint:staticcheck // This value is deprecated but is still returned by the server as a default node role
					protos.NodeRole_INDEX_UPDATE,
					protos.NodeRole_KV_READ,
				},
			},
			want: []string{"INDEX_QUERY", "INDEX_UPDATE", "KV_READ"},
		},
		{
			name: "TestSingleRole",
			args: args{roles: []protos.NodeRole{protos.NodeRole_INDEX_QUERY}},
			want: []string{"INDEX_QUERY"},
		},
		{
			name: "TestNoRoles",
			args: args{roles: []protos.NodeRole{}},
			want: []string{},
		},
		{
			name: "TestNilRoles",
			args: args{roles: nil},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatRoles(tt.args.roles); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("formatRoles() = %v, want %v", got, tt.want)
			}
		})
	}
}
