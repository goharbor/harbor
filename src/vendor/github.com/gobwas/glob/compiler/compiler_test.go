package compiler

import (
	"github.com/gobwas/glob/match"
	"github.com/gobwas/glob/match/debug"
	"github.com/gobwas/glob/syntax/ast"
	"reflect"
	"testing"
)

var separators = []rune{'.'}

func TestCommonChildren(t *testing.T) {
	for i, test := range []struct {
		nodes []*ast.Node
		left  []*ast.Node
		right []*ast.Node
	}{
		{
			nodes: []*ast.Node{
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"z"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
				),
			},
		},
		{
			nodes: []*ast.Node{
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"z"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
				),
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
				),
			},
			left: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"a"}),
			},
			right: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"c"}),
			},
		},
		{
			nodes: []*ast.Node{
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
					ast.NewNode(ast.KindText, ast.Text{"d"}),
				),
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
					ast.NewNode(ast.KindText, ast.Text{"d"}),
				),
			},
			left: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"a"}),
				ast.NewNode(ast.KindText, ast.Text{"b"}),
			},
			right: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"c"}),
				ast.NewNode(ast.KindText, ast.Text{"d"}),
			},
		},
		{
			nodes: []*ast.Node{
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
				),
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"b"}),
					ast.NewNode(ast.KindText, ast.Text{"c"}),
				),
			},
			left: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"a"}),
				ast.NewNode(ast.KindText, ast.Text{"b"}),
			},
			right: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"c"}),
			},
		},
		{
			nodes: []*ast.Node{
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"d"}),
				),
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"d"}),
				),
				ast.NewNode(ast.KindNothing, nil,
					ast.NewNode(ast.KindText, ast.Text{"a"}),
					ast.NewNode(ast.KindText, ast.Text{"e"}),
				),
			},
			left: []*ast.Node{
				ast.NewNode(ast.KindText, ast.Text{"a"}),
			},
			right: []*ast.Node{},
		},
	} {
		left, right := commonChildren(test.nodes)
		if !nodesEqual(left, test.left) {
			t.Errorf("[%d] left, right := commonChildren(); left = %v; want %v", i, left, test.left)
		}
		if !nodesEqual(right, test.right) {
			t.Errorf("[%d] left, right := commonChildren(); right = %v; want %v", i, right, test.right)
		}
	}
}

func nodesEqual(a, b []*ast.Node) bool {
	if len(a) != len(b) {
		return false
	}
	for i, av := range a {
		if !av.Equal(b[i]) {
			return false
		}
	}
	return true
}

func TestGlueMatchers(t *testing.T) {
	for id, test := range []struct {
		in  []match.Matcher
		exp match.Matcher
	}{
		{
			[]match.Matcher{
				match.NewSuper(),
				match.NewSingle(nil),
			},
			match.NewMin(1),
		},
		{
			[]match.Matcher{
				match.NewAny(separators),
				match.NewSingle(separators),
			},
			match.EveryOf{match.Matchers{
				match.NewMin(1),
				match.NewContains(string(separators), true),
			}},
		},
		{
			[]match.Matcher{
				match.NewSingle(nil),
				match.NewSingle(nil),
				match.NewSingle(nil),
			},
			match.EveryOf{match.Matchers{
				match.NewMin(3),
				match.NewMax(3),
			}},
		},
		{
			[]match.Matcher{
				match.NewList([]rune{'a'}, true),
				match.NewAny([]rune{'a'}),
			},
			match.EveryOf{match.Matchers{
				match.NewMin(1),
				match.NewContains("a", true),
			}},
		},
	} {
		act, err := compileMatchers(test.in)
		if err != nil {
			t.Errorf("#%d convert matchers error: %s", id, err)
			continue
		}

		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers result:\nact: %#v;\nexp: %#v", id, act, test.exp)
			continue
		}
	}
}

func TestCompileMatchers(t *testing.T) {
	for id, test := range []struct {
		in  []match.Matcher
		exp match.Matcher
	}{
		{
			[]match.Matcher{
				match.NewSuper(),
				match.NewSingle(separators),
				match.NewText("c"),
			},
			match.NewBTree(
				match.NewText("c"),
				match.NewBTree(
					match.NewSingle(separators),
					match.NewSuper(),
					nil,
				),
				nil,
			),
		},
		{
			[]match.Matcher{
				match.NewAny(nil),
				match.NewText("c"),
				match.NewAny(nil),
			},
			match.NewBTree(
				match.NewText("c"),
				match.NewAny(nil),
				match.NewAny(nil),
			),
		},
		{
			[]match.Matcher{
				match.NewRange('a', 'c', true),
				match.NewList([]rune{'z', 't', 'e'}, false),
				match.NewText("c"),
				match.NewSingle(nil),
			},
			match.NewRow(
				4,
				match.Matchers{
					match.NewRange('a', 'c', true),
					match.NewList([]rune{'z', 't', 'e'}, false),
					match.NewText("c"),
					match.NewSingle(nil),
				}...,
			),
		},
	} {
		act, err := compileMatchers(test.in)
		if err != nil {
			t.Errorf("#%d convert matchers error: %s", id, err)
			continue
		}

		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers result:\nact: %#v\nexp: %#v", id, act, test.exp)
			continue
		}
	}
}

func TestConvertMatchers(t *testing.T) {
	for id, test := range []struct {
		in, exp []match.Matcher
	}{
		{
			[]match.Matcher{
				match.NewRange('a', 'c', true),
				match.NewList([]rune{'z', 't', 'e'}, false),
				match.NewText("c"),
				match.NewSingle(nil),
				match.NewAny(nil),
			},
			[]match.Matcher{
				match.NewRow(
					4,
					[]match.Matcher{
						match.NewRange('a', 'c', true),
						match.NewList([]rune{'z', 't', 'e'}, false),
						match.NewText("c"),
						match.NewSingle(nil),
					}...,
				),
				match.NewAny(nil),
			},
		},
		{
			[]match.Matcher{
				match.NewRange('a', 'c', true),
				match.NewList([]rune{'z', 't', 'e'}, false),
				match.NewText("c"),
				match.NewSingle(nil),
				match.NewAny(nil),
				match.NewSingle(nil),
				match.NewSingle(nil),
				match.NewAny(nil),
			},
			[]match.Matcher{
				match.NewRow(
					3,
					match.Matchers{
						match.NewRange('a', 'c', true),
						match.NewList([]rune{'z', 't', 'e'}, false),
						match.NewText("c"),
					}...,
				),
				match.NewMin(3),
			},
		},
	} {
		act := minimizeMatchers(test.in)
		if !reflect.DeepEqual(act, test.exp) {
			t.Errorf("#%d unexpected convert matchers 2 result:\nact: %#v\nexp: %#v", id, act, test.exp)
			continue
		}
	}
}

func TestCompiler(t *testing.T) {
	for id, test := range []struct {
		ast    *ast.Node
		result match.Matcher
		sep    []rune
	}{
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			sep:    separators,
			result: match.NewAny(separators),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewSuper(),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
			),
			result: match.NewSuper(),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSingle, nil),
			),
			sep:    separators,
			result: match.NewSingle(separators),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{
					Lo:  'a',
					Hi:  'z',
					Not: true,
				}),
			),
			result: match.NewRange('a', 'z', true),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindList, ast.List{
					Chars: "abc",
					Not:   true,
				}),
			),
			result: match.NewList([]rune{'a', 'b', 'c'}, true),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.EveryOf{Matchers: match.Matchers{
				match.NewMin(3),
				match.NewContains(string(separators), true),
			}},
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			result: match.NewMin(3),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.NewBTree(
				match.NewRow(
					4,
					match.Matchers{
						match.NewText("abc"),
						match.NewSingle(separators),
					}...,
				),
				match.NewAny(separators),
				nil,
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"/"}),
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindText, ast.Text{"z"}),
					ast.NewNode(ast.KindText, ast.Text{"ab"}),
				),
				ast.NewNode(ast.KindSuper, nil),
			),
			sep: separators,
			result: match.NewBTree(
				match.NewText("/"),
				nil,
				match.NewBTree(
					match.NewAnyOf(match.NewText("z"), match.NewText("ab")),
					nil,
					match.NewSuper(),
				),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSingle, nil),
			),
			sep: separators,
			result: match.NewBTree(
				match.NewRow(
					5,
					match.Matchers{
						match.NewSingle(separators),
						match.NewText("abc"),
						match.NewSingle(separators),
					}...,
				),
				match.NewSuper(),
				nil,
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewSuffix("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewPrefix("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"def"}),
			),
			result: match.NewPrefixSuffix("abc", "def"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewContains("abc", false),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindAny, nil),
				ast.NewNode(ast.KindAny, nil),
			),
			sep: separators,
			result: match.NewBTree(
				match.NewText("abc"),
				match.NewAny(separators),
				match.NewAny(separators),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
				ast.NewNode(ast.KindSuper, nil),
				ast.NewNode(ast.KindSingle, nil),
			),
			result: match.NewBTree(
				match.NewText("abc"),
				match.NewMin(1),
				match.NewMin(1),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindText, ast.Text{"abc"}),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindAnyOf, nil,
							ast.NewNode(ast.KindPattern, nil,
								ast.NewNode(ast.KindText, ast.Text{"abc"}),
							),
						),
					),
				),
			),
			result: match.NewText("abc"),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindSingle, nil),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "def"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
					),
				),
			),
			result: match.NewBTree(
				match.NewText("abc"),
				nil,
				match.AnyOf{Matchers: match.Matchers{
					match.NewSingle(nil),
					match.NewList([]rune{'d', 'e', 'f'}, false),
					match.NewNothing(),
				}},
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'z'}),
				ast.NewNode(ast.KindRange, ast.Range{Lo: 'a', Hi: 'x', Not: true}),
				ast.NewNode(ast.KindAny, nil),
			),
			result: match.NewBTree(
				match.NewRow(
					2,
					match.Matchers{
						match.NewRange('a', 'z', false),
						match.NewRange('a', 'x', true),
					}...,
				),
				nil,
				match.NewSuper(),
			),
		},
		{
			ast: ast.NewNode(ast.KindPattern, nil,
				ast.NewNode(ast.KindAnyOf, nil,
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "abc"}),
						ast.NewNode(ast.KindText, ast.Text{"ghi"}),
					),
					ast.NewNode(ast.KindPattern, nil,
						ast.NewNode(ast.KindText, ast.Text{"abc"}),
						ast.NewNode(ast.KindList, ast.List{Chars: "def"}),
						ast.NewNode(ast.KindText, ast.Text{"ghi"}),
					),
				),
			),
			result: match.NewRow(
				7,
				match.Matchers{
					match.NewText("abc"),
					match.AnyOf{Matchers: match.Matchers{
						match.NewList([]rune{'a', 'b', 'c'}, false),
						match.NewList([]rune{'d', 'e', 'f'}, false),
					}},
					match.NewText("ghi"),
				}...,
			),
		},
	} {
		m, err := Compile(test.ast, test.sep)
		if err != nil {
			t.Errorf("compilation error: %s", err)
			continue
		}

		if !reflect.DeepEqual(m, test.result) {
			t.Errorf("[%d] Compile():\nexp: %#v\nact: %#v\n\ngraphviz:\nexp:\n%s\nact:\n%s\n", id, test.result, m, debug.Graphviz("", test.result.(match.Matcher)), debug.Graphviz("", m.(match.Matcher)))
			continue
		}
	}
}
