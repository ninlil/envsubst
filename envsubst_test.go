package envsubst

import "testing"

func TestConvertString(t *testing.T) {

	x := make(map[string]string)
	x["X"] = "+"

	type args struct {
		str    string
		fields map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"empty", args{"", nil}, "", false},
		{"no-change", args{"test", nil}, "test", false},

		{"simple-var", args{"1$(X)2", x}, "1+2", false},
		{"double-var", args{"1$(X)$(X)2", x}, "1++2", false},

		{"single-$", args{"1$2", x}, "1$2", false},
		{"double-$", args{"1$$2", x}, "1$$2", false},
		{"triple-$", args{"1$$$2", x}, "1$$$2", false},

		{"double-$-var", args{"1$$(X)2", x}, "1$+2", false},

		{"not-ended", args{"1$(X", x}, "1$(X", false},
		{"$", args{"$", x}, "$", false},
		{"$(", args{"$(", x}, "$(", false},
		{"$(xyzzy", args{"$(xyzzy", x}, "$(xyzzy", false},
		{"$()", args{"$()", x}, "$()", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ConvertString(tt.args.str, Map(tt.args.fields))
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvertString_Options(t *testing.T) {

	x := make(map[string]string)
	x["X"] = "+"

	type args struct {
		prefix  rune
		wrapper rune
		str     string
		fields  map[string]string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"$ (", args{'$', '(', "1$(X)2", x}, "1+2", false},
		{"$ )", args{'$', ')', "1$(X)2", x}, "1+2", false},
		{"$ {", args{'$', '{', "1${X}2", x}, "1+2", false},
		{"$ }", args{'$', '}', "1${X}2", x}, "1+2", false},
		{"$ [", args{'$', '[', "1$[X]2", x}, "1+2", false},
		{"$ ]", args{'$', ']', "1$[X]2", x}, "1+2", false},
		{"$ <", args{'$', '<', "1$<X>2", x}, "1+2", false},
		{"$ >", args{'$', '>', "1$<X>2", x}, "1+2", false},

		{"$ ) {}", args{'$', ')', "1${X}2", x}, "1${X}2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetPrefix(tt.args.prefix)
			SetWrapper(tt.args.wrapper)
			got, err := ConvertString(tt.args.str, Map(tt.args.fields))
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ConvertString() = %v, want %v", got, tt.want)
			}
		})
	}
}
