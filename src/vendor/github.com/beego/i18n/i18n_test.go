package i18n

import "testing"

const enusIni = `
[base]
say_hi = hello %s
hello_world = hello world!
`

const zhcnIni = `
[base]
say_hi = 你好 %s
hello_world = 你好世界！
`

func TestSetMessageData(t *testing.T) {
	var err error
	enusData := []byte(enusIni)
	if err = SetMessageData("en-US", enusData); err != nil {
		t.Fatal(err)
	}

	zhcnData := []byte(zhcnIni)
	if err = SetMessageData("zh-CN", zhcnData); err != nil {
		t.Fatal(err)
	}

	enVal := Tr("en-US", "base.say_hi", "someone")
	if enVal != "hello someone" {
		t.Fatalf("Expect %q, actual %q", "hello someone", enVal)
	}

	zhVal := Tr("zh-CN", "base.hello_world")
	if zhVal != "你好世界！" {
		t.Fatalf("Expect %q, actual %q", "你好世界！", zhVal)
	}
}
