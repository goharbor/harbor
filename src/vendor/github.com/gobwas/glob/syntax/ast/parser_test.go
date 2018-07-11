package ast

import (
	"github.com/gobwas/glob/syntax/lexer"
	"reflect"
	"testing"
)

type stubLexer struct {
	tokens []lexer.Token
	pos    int
}

func (s *stubLexer) Next() (ret lexer.Token) {
	if s.pos == len(s.tokens) {
		return lexer.Token{lexer.EOF, ""}
	}
	ret = s.tokens[s.pos]
	s.pos++
	return
}

func TestParseString(t *testing.T) {
	for id, test := range []struct {
		tokens []lexer.Token
		tree   *Node
	}{
		{
			//pattern: "abc",
			tokens: []lexer.Token{
				{lexer.Text, "abc"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "abc"}),
			),
		},
		{
			//pattern: "a*c",
			tokens: []lexer.Token{
				{lexer.Text, "a"},
				{lexer.Any, "*"},
				{lexer.Text, "c"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindAny, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			//pattern: "a**c",
			tokens: []lexer.Token{
				{lexer.Text, "a"},
				{lexer.Super, "**"},
				{lexer.Text, "c"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSuper, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			//pattern: "a?c",
			tokens: []lexer.Token{
				{lexer.Text, "a"},
				{lexer.Single, "?"},
				{lexer.Text, "c"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "a"}),
				NewNode(KindSingle, nil),
				NewNode(KindText, Text{Text: "c"}),
			),
		},
		{
			//pattern: "[!a-z]",
			tokens: []lexer.Token{
				{lexer.RangeOpen, "["},
				{lexer.Not, "!"},
				{lexer.RangeLo, "a"},
				{lexer.RangeBetween, "-"},
				{lexer.RangeHi, "z"},
				{lexer.RangeClose, "]"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindRange, Range{Lo: 'a', Hi: 'z', Not: true}),
			),
		},
		{
			//pattern: "[az]",
			tokens: []lexer.Token{
				{lexer.RangeOpen, "["},
				{lexer.Text, "az"},
				{lexer.RangeClose, "]"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindList, List{Chars: "az"}),
			),
		},
		{
			//pattern: "{a,z}",
			tokens: []lexer.Token{
				{lexer.TermsOpen, "{"},
				{lexer.Text, "a"},
				{lexer.Separator, ","},
				{lexer.Text, "z"},
				{lexer.TermsClose, "}"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
					),
				),
			),
		},
		{
			//pattern: "/{z,ab}*",
			tokens: []lexer.Token{
				{lexer.Text, "/"},
				{lexer.TermsOpen, "{"},
				{lexer.Text, "z"},
				{lexer.Separator, ","},
				{lexer.Text, "ab"},
				{lexer.TermsClose, "}"},
				{lexer.Any, "*"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindText, Text{Text: "/"}),
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "z"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "ab"}),
					),
				),
				NewNode(KindAny, nil),
			),
		},
		{
			//pattern: "{a,{x,y},?,[a-z],[!qwe]}",
			tokens: []lexer.Token{
				{lexer.TermsOpen, "{"},
				{lexer.Text, "a"},
				{lexer.Separator, ","},
				{lexer.TermsOpen, "{"},
				{lexer.Text, "x"},
				{lexer.Separator, ","},
				{lexer.Text, "y"},
				{lexer.TermsClose, "}"},
				{lexer.Separator, ","},
				{lexer.Single, "?"},
				{lexer.Separator, ","},
				{lexer.RangeOpen, "["},
				{lexer.RangeLo, "a"},
				{lexer.RangeBetween, "-"},
				{lexer.RangeHi, "z"},
				{lexer.RangeClose, "]"},
				{lexer.Separator, ","},
				{lexer.RangeOpen, "["},
				{lexer.Not, "!"},
				{lexer.Text, "qwe"},
				{lexer.RangeClose, "]"},
				{lexer.TermsClose, "}"},
				{lexer.EOF, ""},
			},
			tree: NewNode(KindPattern, nil,
				NewNode(KindAnyOf, nil,
					NewNode(KindPattern, nil,
						NewNode(KindText, Text{Text: "a"}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindAnyOf, nil,
							NewNode(KindPattern, nil,
								NewNode(KindText, Text{Text: "x"}),
							),
							NewNode(KindPattern, nil,
								NewNode(KindText, Text{Text: "y"}),
							),
						),
					),
					NewNode(KindPattern, nil,
						NewNode(KindSingle, nil),
					),
					NewNode(KindPattern, nil,
						NewNode(KindRange, Range{Lo: 'a', Hi: 'z', Not: false}),
					),
					NewNode(KindPattern, nil,
						NewNode(KindList, List{Chars: "qwe", Not: true}),
					),
				),
			),
		},
	} {
		lexer := &stubLexer{tokens: test.tokens}
		result, err := Parse(lexer)
		if err != nil {
			t.Errorf("[%d] unexpected error: %s", id, err)
		}
		if !reflect.DeepEqual(test.tree, result) {
			t.Errorf("[%d] Parse():\nact:\t%s\nexp:\t%s\n", id, result, test.tree)
		}
	}
}
