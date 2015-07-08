package string

import "testing"

func Test(t *testing.T){
	var tests = []struct{
		s, want string
	}{
		{"Backward","drawkcaB"},
		{"Hello World", "dlroW olleH"},
		{"",""},
	}
	for _, c := range tests{
		got := Reverse(c.s)
		if got != c.want{
			t.Errorf("Reverse(%s) == %s, want %s",c.s,got,c.want)
		}
	}
}
