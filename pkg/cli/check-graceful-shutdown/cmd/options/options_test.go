package options

import (
	"net/url"
	"reflect"
	"testing"
)

func TestNewConfigWithDefaults(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewConfigWithDefaults(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewConfigWithDefaults() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURI_String(t *testing.T) {
	type fields struct {
		Val url.URL
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URI{
				Val: tt.fields.Val,
			}
			if got := u.String(); got != tt.want {
				t.Errorf("URI.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURI_Set(t *testing.T) {
	type fields struct {
		Val url.URL
	}
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URI{
				Val: tt.fields.Val,
			}
			if err := u.Set(tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("URI.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestURI_Type(t *testing.T) {
	type fields struct {
		Val url.URL
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &URI{
				Val: tt.fields.Val,
			}
			if got := u.Type(); got != tt.want {
				t.Errorf("URI.Type() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCutProcessConfigFromArgs(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name  string
		args  args
		want  []string
		want1 ProcessConfig
	}{
		{
			name:  "ok_simple",
			args:  args{args: []string{"foo", "-b", "bar", "-c", "az", "--", "./second", "--foobar", "-bar", "-bar", "config.yaml"}},
			want:  []string{"foo", "-b", "bar", "-c", "az"},
			want1: ProcessConfig{Command: "./second", Arguments: []string{"--foobar", "-bar", "-bar", "config.yaml"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := CutProcessConfigFromArgs(tt.args.args...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CutProcessConfigFromArgs() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("CutProcessConfigFromArgs() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
