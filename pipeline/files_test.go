package pipeline

import "testing"

func Test_checkFileAccessibility(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "check permission denied",
			args: args{
				filename: "../test_configs/testFile",
			},
			wantErr: false,
		},
		{
			name: "check derivet string is dir",
			args: args{
				filename: "../test_configs",
			},
			wantErr: true,
		},
		{
			name: "check completed",
			args: args{
				filename: "../test_configs/testmap.hocon",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := checkFileAccessibility(tt.args.filename); (err != nil) != tt.wantErr {
				t.Errorf("checkFileAccessibility() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
