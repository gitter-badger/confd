package template

import (
	"path"
	"reflect"
	"runtime"
	"testing"
	"time"
)

type tstCompareType int

const (
	tstEq tstCompareType = iota
	tstNe
	tstGt
	tstGe
	tstLt
	tstLe
)

func tstIsEq(tp tstCompareType) bool {
	return tp == tstEq || tp == tstGe || tp == tstLe
}

func tstIsGt(tp tstCompareType) bool {
	return tp == tstGt || tp == tstGe
}

func tstIsLt(tp tstCompareType) bool {
	return tp == tstLt || tp == tstLe
}

var templateExtFuncTests = []templateTest{

	templateTest{
		desc: "add test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{add .Value 1}}
{{end}}
`,
		expected: `

key: key
val: 2

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "1")
		},
	},
	templateTest{
		desc: "sub test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{sub .Value 1}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "2")
		},
	},
	templateTest{
		desc: "div test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{div .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "div test2",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{div .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1.5

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3.00")
		},
	},
	templateTest{
		desc: "mul test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{mul .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 6

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},

	templateTest{
		desc: "gt test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{if gt .Value 2}}gt{{end}}
{{end}}
`,
		expected: `

key: key
val: gt

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "lt test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{if lt .Value 2}}lt{{end}}
{{end}}
`,
		expected: `

key: key
val: lt

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "1")
		},
	},
	templateTest{
		desc: "mod test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{with get "/test/key"}}
key: {{base .Key}}
val: {{mod .Value 2}}
{{end}}
`,
		expected: `

key: key
val: 1

`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key", "3")
		},
	},
	templateTest{
		desc: "max test",
		toml: `
[template]
src = "test.conf.tmpl"
dest = "./tmp/test.conf"
keys = [
    "/test/key",
]
`,
		tmpl: `
{{$nodes := gets "/test/key/*"}}
len: {{len $nodes}}
max: {{max (len $nodes) 3}}
`,
		expected: `

len: 2
max: 3
`,
		updateStore: func(tr *TemplateResource) {
			tr.store.Set("/test/key/n1", "v1")
			tr.store.Set("/test/key/n2", "v2")
		},
	},
}

func TestFuncsInTemplate(t *testing.T) {
	for _, tt := range templateExtFuncTests {
		ExecuteTestTemplate(tt, t)
	}
}

func TestCompare(t *testing.T) {
	for _, this := range []struct {
		tstCompareType
		funcUnderTest func(a, b interface{}) bool
	}{
		{tstGt, gt},
		{tstLt, lt},
		{tstGe, ge},
		{tstLe, le},
		{tstEq, eq},
		{tstNe, ne},
	} {
		doTestCompare(t, this.tstCompareType, this.funcUnderTest)
	}
}

func toTime(value string) time.Time {
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		println(err.Error())
		t = time.Now()
	}
	return t
}

func doTestCompare(t *testing.T, tp tstCompareType, funcUnderTest func(a, b interface{}) bool) {
	for i, this := range []struct {
		left            interface{}
		right           interface{}
		expectIndicator int
	}{
		{5, 8, -1},
		{8, 5, 1},
		{5, 5, 0},
		{int(5), int64(5), 0},
		{int32(5), int(5), 0},
		{int16(4), int(5), -1},
		{uint(15), uint64(15), 0},
		{-2, 1, -1},
		{2, -5, 1},
		{0.0, 1.23, -1},
		{1.1, 1.1, 0},
		{float32(1.0), float64(1.0), 0},
		{1.23, 0.0, 1},
		{"5", "5", 0},
		{"5", 5, 0},
		{5, "5", 0},
		{5.1, "5.1", 0},
		{5.0, "5", 0},
		{5, "5.0", 0},
		{"8", 5, 1},
		{5, "8", -1},
		{"8", "5", 1},
		{"8", "5.1", 1},
		{"8", 5.1, 1},
		{8, "5.1", 1},
		{"5", "0001", 1},
		{"a", "a", 0},
		{"a", "b", -1},
		{"b", "a", 1},
		{[]int{100, 99}, []int{1, 2, 3, 4}, -1},
		{toTime("2015-11-20"), toTime("2015-11-20"), 0},
		{toTime("2015-11-19"), toTime("2015-11-20"), -1},
		{toTime("2015-11-20"), toTime("2015-11-19"), 1},
	} {
		result := funcUnderTest(this.left, this.right)
		success := false

		if this.expectIndicator == 0 {
			if tstIsEq(tp) {
				success = result
			} else {
				success = !result
			}
		}

		if this.expectIndicator < 0 {
			success = result && (tstIsLt(tp) || tp == tstNe)
			success = success || (!result && !tstIsLt(tp))
		}

		if this.expectIndicator > 0 {
			success = result && (tstIsGt(tp) || tp == tstNe)
			success = success || (!result && (!tstIsGt(tp) || tp != tstNe))
		}

		if !success {
			t.Errorf("[%d][%s] %v compared to %v: %t", i, path.Base(runtime.FuncForPC(reflect.ValueOf(funcUnderTest).Pointer()).Name()), this.left, this.right, result)
		}
	}
}

func TestMod(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		expect interface{}
	}{
		{3, 2, int64(1)},
		{"3", 2, int64(1)},
		{3, "2", int64(1)},
		{3, 1, int64(0)},
		{3, 0, false},
		{0, 3, int64(0)},
		{3.1, 2, false},
		{3, 2.1, false},
		{3.1, 2.1, false},
		{int8(3), int8(2), int64(1)},
		{int16(3), int16(2), int64(1)},
		{int32(3), int32(2), int64(1)},
		{int64(3), int64(2), int64(1)},
	} {
		result, err := mod(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] modulo didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] modulo got %v but expected %v", i, result, this.expect)
			}
		}
	}
}

func TestMaxAndMin(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		expect interface{}
	}{
		{3, 2, float64(3)},
		{"3", 2, float64(3)},
		{3, "2", float64(3)},
		{3.1, 3, float64(3.1)},
		{3, "a", false},
		{int8(3), int8(2), float64(3)},
		{int16(3), int16(2), float64(3)},
		{int32(3), int32(2), float64(3)},
		{int64(3), int64(2), float64(3)},
		{float64(3.0001), float64(3.00011), float64(3.00011)},
	} {
		result, err := max(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] max didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] max got %v but expected %v", i, result, this.expect)
			}
		}

		result, err = min(this.a, this.b)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] min didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] min not expected %v", i, result)
			}
		}
	}
}

func TestTimeUnix(t *testing.T) {
	var sec int64 = 1234567890
	tv := reflect.ValueOf(time.Unix(sec, 0))
	i := 1

	res := toTimeUnix(tv)
	if sec != res {
		t.Errorf("[%d] timeUnix got %v but expected %v", i, res, sec)
	}

	i++
	func(t *testing.T) {
		defer func() {
			if err := recover(); err == nil {
				t.Errorf("[%d] timeUnix didn't return an expected error", i)
			}
		}()
		iv := reflect.ValueOf(sec)
		toTimeUnix(iv)
	}(t)
}

func TestDoArithmetic(t *testing.T) {
	for i, this := range []struct {
		a      interface{}
		b      interface{}
		op     rune
		expect interface{}
	}{
		{3, 2, '+', int64(5)},
		{3, 2, '-', int64(1)},
		{3, 2, '*', int64(6)},
		{3, 2, '/', int64(1)},
		{3.0, 2, '+', float64(5)},
		{3.0, 2, '-', float64(1)},
		{3.0, 2, '*', float64(6)},
		{3.0, 2, '/', float64(1.5)},
		{3, 2.0, '+', float64(5)},
		{3, 2.0, '-', float64(1)},
		{3, 2.0, '*', float64(6)},
		{3, 2.0, '/', float64(1.5)},
		{3.0, 2.0, '+', float64(5)},
		{3.0, 2.0, '-', float64(1)},
		{3.0, 2.0, '*', float64(6)},
		{3.0, 2.0, '/', float64(1.5)},
		{uint(3), uint(2), '+', int64(5)},
		{uint(3), uint(2), '-', int64(1)},
		{uint(3), uint(2), '*', int64(6)},
		{uint(3), uint(2), '/', int64(1)},
		{uint(3), 2, '+', int64(5)},
		{uint(3), 2, '-', int64(1)},
		{uint(3), 2, '*', int64(6)},
		{uint(3), 2, '/', int64(1)},
		{3, uint(2), '+', int64(5)},
		{3, uint(2), '-', int64(1)},
		{3, uint(2), '*', int64(6)},
		{3, uint(2), '/', int64(1)},
		{uint(3), -2, '+', int64(1)},
		{uint(3), -2, '-', int64(5)},
		{uint(3), -2, '*', int64(-6)},
		{uint(3), -2, '/', int64(-1)},
		{-3, uint(2), '+', int64(-1)},
		{-3, uint(2), '-', int64(-5)},
		{-3, uint(2), '*', int64(-6)},
		{-3, uint(2), '/', int64(-1)},
		{uint(3), 2.0, '+', float64(5)},
		{uint(3), 2.0, '-', float64(1)},
		{uint(3), 2.0, '*', float64(6)},
		{uint(3), 2.0, '/', float64(1.5)},
		{3.0, uint(2), '+', float64(5)},
		{3.0, uint(2), '-', float64(1)},
		{3.0, uint(2), '*', float64(6)},
		{3.0, uint(2), '/', float64(1.5)},
		{0, 0, '+', 0},
		{0, 0, '-', 0},
		{0, 0, '*', 0},
		{"foo", "bar", '+', false},
		{3, 0, '/', false},
		{3.0, 0, '/', false},
		{3, 0.0, '/', false},
		{uint(3), uint(0), '/', false},
		{3, uint(0), '/', false},
		{-3, uint(0), '/', false},
		{uint(3), 0, '/', false},
		{3.0, uint(0), '/', false},
		{uint(3), 0.0, '/', false},
		{3, "foo", '+', false},
		{3.0, "foo", '+', false},
		{uint(3), "foo", '+', false},
		{"foo", 3, '+', false},
		{"foo", "bar", '-', false},
		{"3", "2", '+', int64(5)},
		{"3", "2", '-', int64(1)},
		{"3", "2", '*', int64(6)},
		{"3", "2", '/', int64(1)},
		{"3.0", "2", '+', float64(5)},
		{"3.0", "2", '-', float64(1)},
		{"3.0", "2", '*', float64(6)},
		{"3.0", "2", '/', float64(1.5)},
		{"3", "2.0", '+', float64(5)},
		{"3", "2.0", '-', float64(1)},
		{"3", "2.0", '*', float64(6)},
		{"3", "2.0", '/', float64(1.5)},
		{"3.0", "2.0", '+', float64(5)},
		{"3.0", "2.0", '-', float64(1)},
		{"3.0", "2.0", '*', float64(6)},
		{"3.0", "2.0", '/', float64(1.5)},
	} {
		result, err := DoArithmetic(this.a, this.b, this.op)
		if b, ok := this.expect.(bool); ok && !b {
			if err == nil {
				t.Errorf("[%d] doArithmetic didn't return an expected error", i)
			}
		} else {
			if err != nil {
				t.Errorf("[%d] failed: %s", i, err)
				continue
			}
			if !reflect.DeepEqual(result, this.expect) {
				t.Errorf("[%d] doArithmetic [%v %s %v ] got %v but expected %v", i, this.a, string(this.op), this.b, result, this.expect)
			}
		}
	}
}
