package complete

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestCompleter_Complete(t *testing.T) {
	initTests()

	c := Command{
		Sub: Commands{
			"sub1": {
				Flags: Flags{
					"-flag1": PredictAnything,
					"-flag2": PredictNothing,
				},
			},
			"sub2": {
				Flags: Flags{
					"-flag2": PredictNothing,
					"-flag3": PredictSet("opt1", "opt2", "opt12"),
				},
				Args: PredictFiles("*.md"),
			},
		},
		Flags: Flags{
			"-o": PredictFiles("*.txt"),
		},
		GlobalFlags: Flags{
			"-h":       PredictNothing,
			"-global1": PredictAnything,
		},
	}
	cmp := New("cmd", c)

	tests := []struct {
		line  string
		point int // -1 indicates len(line)
		want  []string
	}{
		{
			line:  "cmd ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -",
			point: -1,
			want:  []string{"-h", "-global1", "-o"},
		},
		{
			line:  "cmd -h ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -global1 ", // global1 is known follow flag
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd sub",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd sub1",
			point: -1,
			want:  []string{"sub1"},
		},
		{
			line:  "cmd sub2",
			point: -1,
			want:  []string{"sub2"},
		},
		{
			line:  "cmd sub1 ",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd sub1 -",
			point: -1,
			want:  []string{"-flag1", "-flag2", "-h", "-global1"},
		},
		{
			line:  "cmd sub2 ",
			point: -1,
			want:  []string{"./", "dir/", "outer/", "readme.md"},
		},
		{
			line:  "cmd sub2 ./",
			point: -1,
			want:  []string{"./", "./readme.md", "./dir/", "./outer/"},
		},
		{
			line:  "cmd sub2 re",
			point: -1,
			want:  []string{"readme.md"},
		},
		{
			line:  "cmd sub2 ./re",
			point: -1,
			want:  []string{"./readme.md"},
		},
		{
			line:  "cmd sub2 -flag2 ",
			point: -1,
			want:  []string{"./", "dir/", "outer/", "readme.md"},
		},
		{
			line:  "cmd sub1 -fl",
			point: -1,
			want:  []string{"-flag1", "-flag2"},
		},
		{
			line:  "cmd sub1 -flag1",
			point: -1,
			want:  []string{"-flag1"},
		},
		{
			line:  "cmd sub1 -flag1 ",
			point: -1,
			want:  []string{}, // flag1 is unknown follow flag
		},
		{
			line:  "cmd sub1 -flag2 -",
			point: -1,
			want:  []string{"-flag1", "-flag2", "-h", "-global1"},
		},
		{
			line:  "cmd -no-such-flag",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd -no-such-flag ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -no-such-flag -",
			point: -1,
			want:  []string{"-h", "-global1", "-o"},
		},
		{
			line:  "cmd no-such-command",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd no-such-command ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -o ",
			point: -1,
			want:  []string{"a.txt", "b.txt", "c.txt", ".dot.txt", "./", "dir/", "outer/"},
		},
		{
			line:  "cmd -o ./no-su",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd -o ./",
			point: -1,
			want:  []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			line:  "cmd -o=./",
			point: -1,
			want:  []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			line:  "cmd -o .",
			point: -1,
			want:  []string{"./a.txt", "./b.txt", "./c.txt", "./.dot.txt", "./", "./dir/", "./outer/"},
		},
		{
			line:  "cmd -o ./b",
			point: -1,
			want:  []string{"./b.txt"},
		},
		{
			line:  "cmd -o=./b",
			point: -1,
			want:  []string{"./b.txt"},
		},
		{
			line:  "cmd -o ./read",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd -o=./read",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd -o ./readme.md",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd -o ./readme.md ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -o=./readme.md ",
			point: -1,
			want:  []string{"sub1", "sub2"},
		},
		{
			line:  "cmd -o sub2 -flag3 ",
			point: -1,
			want:  []string{"opt1", "opt2", "opt12"},
		},
		{
			line:  "cmd -o sub2 -flag3 opt1",
			point: -1,
			want:  []string{"opt1", "opt12"},
		},
		{
			line:  "cmd -o sub2 -flag3 opt",
			point: -1,
			want:  []string{"opt1", "opt2", "opt12"},
		},
		{
			line: "cmd -o ./b foo",
			//               ^
			point: 10,
			want:  []string{"./b.txt"},
		},
		{
			line: "cmd -o=./b foo",
			//               ^
			point: 10,
			want:  []string{"./b.txt"},
		},
		{
			line: "cmd -o sub2 -flag3 optfoo",
			//                           ^
			point: 22,
			want:  []string{"opt1", "opt2", "opt12"},
		},
		{
			line: "cmd -o ",
			//         ^
			point: 4,
			want:  []string{"sub1", "sub2"},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s@%d", tt.line, tt.point), func(t *testing.T) {
			got := runComplete(cmp, tt.line, tt.point)

			sort.Strings(tt.want)
			sort.Strings(got)

			if !equalSlices(got, tt.want) {
				t.Errorf("failed '%s'\ngot: %s\nwant: %s", t.Name(), got, tt.want)
			}
		})
	}
}

func TestCompleter_Complete_SharedPrefix(t *testing.T) {
	initTests()

	c := Command{
		Sub: Commands{
			"status": {
				Flags: Flags{
					"-f3": PredictNothing,
				},
			},
			"job": {
				Sub: Commands{
					"status": {
						Flags: Flags{
							"-f4": PredictNothing,
						},
					},
				},
			},
		},
		Flags: Flags{
			"-o": PredictFiles("*.txt"),
		},
		GlobalFlags: Flags{
			"-h":       PredictNothing,
			"-global1": PredictAnything,
		},
	}

	cmp := New("cmd", c)

	tests := []struct {
		line  string
		point int // -1 indicates len(line)
		want  []string
	}{
		{
			line:  "cmd ",
			point: -1,
			want:  []string{"status", "job"},
		},
		{
			line:  "cmd -",
			point: -1,
			want:  []string{"-h", "-global1", "-o"},
		},
		{
			line:  "cmd j",
			point: -1,
			want:  []string{"job"},
		},
		{
			line:  "cmd job ",
			point: -1,
			want:  []string{"status"},
		},
		{
			line:  "cmd job -",
			point: -1,
			want:  []string{"-h", "-global1"},
		},
		{
			line:  "cmd job status ",
			point: -1,
			want:  []string{},
		},
		{
			line:  "cmd job status -",
			point: -1,
			want:  []string{"-f4", "-h", "-global1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			got := runComplete(cmp, tt.line, tt.point)

			sort.Strings(tt.want)
			sort.Strings(got)

			if !equalSlices(got, tt.want) {
				t.Errorf("failed '%s'\ngot = %s\nwant: %s", t.Name(), got, tt.want)
			}
		})
	}
}

// runComplete runs the complete login for test purposes
// it gets the complete struct and command line arguments and returns
// the complete options
func runComplete(c *Complete, line string, point int) (completions []string) {
	if point == -1 {
		point = len(line)
	}
	os.Setenv(envLine, line)
	os.Setenv(envPoint, strconv.Itoa(point))
	b := bytes.NewBuffer(nil)
	c.Out = b
	c.Complete()
	completions = parseOutput(b.String())
	return
}

func parseOutput(output string) []string {
	lines := strings.Split(output, "\n")
	options := []string{}
	for _, l := range lines {
		if l != "" {
			options = append(options, l)
		}
	}
	return options
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
