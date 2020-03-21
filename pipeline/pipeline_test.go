package pipeline

import (
	"reflect"
	"testing"

	"github.com/goreflect/gostructor/infra"
)

func Test_getFunctionChain(t *testing.T) {
	type args struct {
		fileName       string
		pipelineChanes []infra.FuncType
	}
	tests := []struct {
		name string
		args args
		want *Pipeline
	}{
		{
			name: "error getting chain functinos by using not implemented source",
			args: args{
				fileName: "",
				pipelineChanes: []infra.FuncType{
					infra.FunctionNotExist,
				},
			},
			want: &Pipeline{chains: &Chain{stageFunction: nil, next: nil, notAValues: []*infra.NotAValue{&infra.NewNotAValue(reflect.ValueOf())}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getFunctionChain(tt.args.fileName, tt.args.pipelineChanes); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getFunctionChain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getChainByIdentifier(t *testing.T) {
	type args struct {
		idFunc   infra.FuncType
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		want    IConfigure
		want1   int
		wantErr bool
	}{
		{
			name: "check function setup default",
			args: args{
				idFunc: infra.FunctionSetupDefault,
			},
			want:    &DefaultConfig{},
			want1:   sourceFileNotUsed,
			wantErr: false,
		},
		{
			name: "check function setup environment",
			args: args{
				idFunc: infra.FunctionSetupEnvironment,
			},
			want:    &EnvironmentConfig{},
			want1:   sourceFileNotUsed,
			wantErr: false,
		},
		{
			name: "check function setup hocon",
			args: args{
				idFunc:   infra.FunctionSetupHocon,
				fileName: "test",
			},
			want:    &HoconConfig{fileName: "test"},
			want1:   sourceFileInDisk,
			wantErr: false,
		},
		{
			name: "check function setup json",
			args: args{
				idFunc: infra.FunctionSetupJson,
			},
			want:    &JsonConfig{},
			want1:   sourceFileInDisk,
			wantErr: true,
		},
		{
			name: "check function setup yaml",
			args: args{
				idFunc: infra.FunctionSetupYaml,
			},
			want:    &YamlConfig{},
			want1:   sourceFileInDisk,
			wantErr: true,
		},
		{
			name: "check function setup config server",
			args: args{
				idFunc: infra.FunctionSetupConfigServer,
			},
			want:    nil,
			want1:   sourceFielInServer,
			wantErr: true,
		},
		{
			name: "check function setup vault",
			args: args{
				idFunc: infra.FunctionSetupVault,
			},
			want:    nil,
			want1:   sourceFielInServer,
			wantErr: true,
		},
		{
			name: "check unknown setup function",
			args: args{
				idFunc: infra.FunctionNotExist,
			},
			want:    nil,
			want1:   sourceFileNotUsed,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getChainByIdentifier(tt.args.idFunc, tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("getChainByIdentifier() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getChainByIdentifier() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getChainByIdentifier() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
