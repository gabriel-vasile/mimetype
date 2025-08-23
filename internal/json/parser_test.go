package json

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/gabriel-vasile/mimetype/internal/scan"
)

// These samples come from https://github.com/nst/JSONTestSuite.
var positives = []struct {
	json   string
	stdlib bool
}{
	{`[[]   ]`, true},
	{`[]`, true},
	{`[""]`, true},
	{`["a"]`, true},
	{`[false]`, true},
	{`[null, 1, "1", {}]`, true},
	{`[null]`, true},
	{`[1
]`, true},
	{` [1]`, true},
	{`[1,null,null,null,2]`, true},
	{`[2] `, true},
	{`[0e+1]`, true},
	{`[0e1]`, true},
	{`[ 4]`, true},
	{`[-0.000000000000000000000000000000000000000000000000000000000000000000000000000001]
`, true},
	{`[20e1]`, true},
	{`[123e65]`, true},
	{`[-0]`, true},
	{`[-123]`, true},
	{`[-1]`, true},
	{`[-0]`, true},
	{`[1E22]`, true},
	{`[1E-2]`, true},
	{`[1E+2]`, true},
	{`[123e45]`, true},
	{`[123.456e78]`, true},
	{`[1e-2]`, true},
	{`[1e+2]`, true},
	{`[123]`, true},
	{`[123.456789]`, true},
	{`{"asd":"sdf"}`, true},
	{`{"a":"b","a":"b"}`, true},
	{`{"a":"b","a":"c"}`, true},
	{`{}`, true},
	{`{"":0}`, true},
	{`{"foo\u0000bar": 42}`, true},
	{`{ "min": -1.0e+28, "max": 1.0e+28 }`, true},
	{`{"asd":"sdf", "dfg":"fgh"}`, true},
	{`{"x":[{"id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}], "id": "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`, true},
	{`{"a":[]}`, true},
	{`{"title":"\u041f\u043e\u043b\u0442\u043e\u0440\u0430 \u0417\u0435\u043c\u043b\u0435\u043a\u043e\u043f\u0430" }`, true},
	{`{
"a": "b"
}`, true},
	{`["\u0060\u012a\u12AB"]`, true},
	{`["\uD801\udc37"]`, true},
	{`["\ud83d\ude39\ud83d\udc8d"]`, true},
	{`["\"\\\/\b\f\n\r\t"]`, true},
	{`["\\u0000"]`, true},
	{`["\""]`, true},
	{`["a/*b*/c/*d//e"]`, true},
	{`["\\a"]`, true},
	{`["\\n"]`, true},
	{`["\u0012"]`, true},
	{`["\uFFFF"]`, true},
	{`["asd"]`, true},
	{`[ "asd"]`, true},
	{`["\uDBFF\uDFFF"]`, true},
	{`["new\u00A0line"]`, true},
	{`["􏿿"]`, true},
	{`["￿"]`, true},
	{`["\u0000"]`, true},
	{`["\u002c"]`, true},
	{`["π"]`, true},
	{`["𛿿"]`, true},
	{`["asd "]`, true},
	{`" "`, true},
	{`["\uD834\uDd1e"]`, true},
	{`["\u0821"]`, true},
	{`["\u0123"]`, true},
	{`[" "]`, true},
	{`[" "]`, true},
	{`["new\u000Aline"]`, true},
	{`["\u0061\u30af\u30EA\u30b9"]`, true},
	{`[""]`, true},
	{`["⍂㈴⍂"]`, true},
	{`["\u005C"]`, true},
	{`["\u0022"]`, true},
	{`["\uA66D"]`, true},
	{`["\uDBFF\uDFFE"]`, true},
	{`["\uD83F\uDFFE"]`, true},
	{`["\u200B"]`, true},
	{`["\u2064"]`, true},
	{`["\uFDD0"]`, true},
	{`["\uFFFE"]`, true},
	{`["€𝄞"]`, true},
	{`["aa"]`, true},
	{`false`, true},
	{`42`, true},
	{`-0.1`, true},
	{`null`, true},
	{`"asd"`, true},
	{`true`, true},
	{`""`, true},
	{`["a"]
`, true},
	{`[true]`, true},
	{` [] `, true},

	// Bug: following samples are invalid JSONs but they are parsed successfully.
	{`    `, false},
	{`["",]`, false},
	{`[1,]`, false},
	{`[-01]`, false},
	{`[-2.]`, false},
	{`[.2e-3]`, false},
	{`[0.e1]`, false},
	{`[2.e+3]`, false},
	{`[2.e-3]`, false},
	{`[2.e3]`, false},
	{`[-012]`, false},
	{`[-.123]`, false},
	{`[1.]`, false},
	{`[.123]`, false},
	{`[012]`, false},
	{`{"�":"0",}`, false},
	{`{"id":0,}`, false},
	{`"`, false},
	{`["new
line"]`, false},
	{`["	"]`, false},
	{`[`, false},
	{`[[`, false},
	{`{`, false},
}

var negatives = []struct {
	name          string
	json          string
	expectParse   int
	expectInspect int
}{
	{"array_1_true_without_comma", `[1 true]`, 1, 3},
	{"array_a_invalid_utf8", `[a�]`, 1, 1},
	{"array_colon_instead_of_comma", `["": 1]`, 1, 3},
	{"array_comma_after_close", `[""],`, 4, 4},
	{"array_comma_and_number", `[,1]`, 1, 1},
	{"array_double_comma", `[1,,2]`, 1, 3},
	{"array_double_extra_comma", `["x",,]`, 1, 5},
	{"array_extra_close", `["x"]]`, 5, 5},
	{"array_incomplete_invalid_value", `[x`, 1, 1},
	{"array_incomplete", `["x"`, 1, 4},
	{"array_inner_array_no_comma", `[3[4]]`, 1, 2},
	{"array_invalid_utf8", `[�]`, 1, 1},
	{"array_items_separated_by_semicolon", `[1:2]`, 1, 2},
	{"array_just_comma", `[,]`, 1, 1},
	{"array_just_minus", `[-]`, 1, 2},
	{"array_missing_value", `[   , ""]`, 1, 4},
	{"array_newlines_unclosed", `["a",
4
,1,`, 1, 11},
	{"array_number_and_several_commas", `[1,,]`, 1, 3},
	{"array_spaces_vertical_tab_formfeed", "\x5b\x22\x0b\x61\x22\x5c\x66\x5d", 1, 5},
	{"array_star_inside", `[*]`, 1, 1},
	{"array_unclosed", `[""`, 1, 3},
	{"array_unclosed_trailing_comma", `[1,`, 1, 3},
	{"array_unclosed_with_new_lines", "\x5b\x31\x2c\x0a\x31\x0a\x2c\x31", 1, 8},
	{"array_unclosed_with_object_inside", `[{}`, 1, 3},
	{"incomplete_false", `[fals]`, 1, 5},
	{"incomplete_null", `[nul]`, 1, 4},
	{"incomplete_true", `[tru]`, 1, 4},
	{"multidigit_number_then_00", "\x31\x32\x33\x00", 3, 3},
	{"number_0.1.2", `[0.1.2]`, 1, 4},
	{"number_0.3e+", `[0.3e+]`, 1, 6},
	{"number_0.3e", `[0.3e]`, 1, 5},
	{"number_0_capital_E+", `[0E+]`, 1, 4},
	{"number_0_capital_E", `[0E]`, 1, 3},
	{"number_0e+", `[0e+]`, 1, 4},
	{"number_0e", `[0e]`, 1, 3},
	{"number_1_000", `[1 000.0]`, 1, 3},
	{"number_1.0e+", `[1.0e+]`, 1, 6},
	{"number_1.0e-", `[1.0e-]`, 1, 6},
	{"number_1.0e", `[1.0e]`, 1, 5},
	{"number_-1.0.", `[-1.0.]`, 1, 5},
	{"number_1eE2", `[1eE2]`, 1, 3},
	{"number_+1", `[+1]`, 1, 1},
	{"number_.-1", `[.-1]`, 1, 2},
	{"number_9.e+", `[9.e+]`, 1, 5},
	{"number_expression", `[1+2]`, 1, 2},
	{"number_hex_1_digit", `[0x1]`, 1, 2},
	{"number_hex_2_digits", `[0x42]`, 1, 2},
	{"number_infinity", `[Infinity]`, 1, 1},
	{"number_+Inf", `[+Inf]`, 1, 1},
	{"number_Inf", `[Inf]`, 1, 1},
	{"number_invalid+-", `[0e+-1]`, 1, 4},
	{"number_invalid-negative-real", `[-123.123foo]`, 1, 9},
	{"number_invalid-utf-8-in-bigger-int", `[123�]`, 1, 4},
	{"number_invalid-utf-8-in-exponent", `[1e1�]`, 1, 4},
	{"number_invalid-utf-8-in-int", "\x5b\x30\xe5\x5d\x0a", 1, 2},
	{"number_++", `[++1234]`, 1, 1},
	{"number_minus_infinity", `[-Infinity]`, 1, 2},
	{"number_minus_sign_with_trailing_garbage", `[-foo]`, 1, 2},
	{"number_minus_space_1", `[- 1]`, 1, 2},
	{"number_-NaN", `[-NaN]`, 1, 2},
	{"number_NaN", `[NaN]`, 1, 1},
	{"number_neg_with_garbage_at_end", `[-1x]`, 1, 3},
	{"number_real_garbage_after_e", `[1ea]`, 1, 3},
	{"number_real_with_invalid_utf8_after_e", `[1e�]`, 1, 3},
	{"number_U+FF11_fullwidth_digit_one", `[１]`, 1, 1},
	{"number_with_alpha_char", `[1.8011670033376514H-308]`, 1, 19},
	{"number_with_alpha", `[1.2a-3]`, 1, 4},
	{"object_bad_value", `["x", truth]`, 1, 9},
	{"object_bracket_key", "\x7b\x5b\x3a\x20\x22\x78\x22\x7d\x0a", 1, 1},
	{"object_comma_instead_of_colon", `{"x", null}`, 1, 4},
	{"object_double_colon", `{"x"::"b"}`, 1, 5},
	{"object_emoji", `{🇨🇭}`, 1, 1},
	{"object_garbage_at_end", `{"a":"a" 123}`, 1, 9},
	{"object_key_with_single_quotes", `{key: 'value'}`, 1, 1},
	{"object_missing_colon", `{"a" b}`, 1, 5},
	{"object_missing_key", `{:"b"}`, 1, 1},
	{"object_missing_semicolon", `{"a" "b"}`, 1, 5},
	{"object_missing_value", `{"a":`, 1, 5},
	{"object_no-colon", `{"a"`, 1, 4},
	{"object_non_string_key_but_huge_number_instead", `{9999E9999:1}`, 1, 1},
	{"object_non_string_key", `{1:1}`, 1, 1},
	{"object_repeated_null_null", `{null:null,null:null}`, 1, 1},
	{"object_several_trailing_commas", `{"id":0,,,,,}`, 1, 8},
	{"object_single_quote", `{'a':0}`, 1, 1},
	{"object_trailing_comment", `{"a":"b"}/**/`, 9, 9},
	{"object_trailing_comment_open", `{"a":"b"}/**//`, 9, 9},
	{"object_trailing_comment_slash_open_incomplete", `{"a":"b"}/`, 9, 9},
	{"object_trailing_comment_slash_open", `{"a":"b"}//`, 9, 9},
	{"object_two_commas_in_a_row", `{"a":"b",,"c":"d"}`, 1, 9},
	{"object_unquoted_key", `{a: "b"}`, 1, 1},
	{"object_unterminated-value", `{"a":"a`, 1, 7},
	{"object_with_single_string", `{ "foo" : "bar", "a" }`, 1, 21},
	{"object_with_trailing_garbage", `{"a":"b"}#`, 9, 9},
	{"single_space", ` `, 0, 1},
	{"string_1_surrogate_then_escape", `["\uD800\"]`, 1, 11},
	{"string_1_surrogate_then_escape_u1", `["\uD800\u1"]`, 1, 11},
	{"string_1_surrogate_then_escape_u1x", `["\uD800\u1x"]`, 1, 11},
	{"string_1_surrogate_then_escape_u", `["\uD800\u"]`, 1, 10},
	{"string_accentuated_char_no_quotes", `[é]`, 1, 1},
	{"string_backslash_00", "\x5b\x22\x5c\x00\x22\x5d", 1, 3},
	{"string_escaped_backslash_bad", `["\\\"]`, 1, 7},
	{"string_escaped_ctrl_char_tab", "\x5b\x22\x5c\x09\x22\x5d", 1, 3},
	{"string_escaped_emoji", `["\🌀"]`, 1, 3},
	{"string_escape_x", `["\x00"]`, 1, 3},
	{"string_incomplete_escaped_character", `["\u00A"]`, 1, 7},
	{"string_incomplete_escape", `["\"]`, 1, 5},
	{"string_incomplete_surrogate_escape_invalid", `["\uD800\uD800\x"]`, 1, 15},
	{"string_incomplete_surrogate", `["\uD834\uDd"]`, 1, 12},
	{"string_invalid_backslash_esc", `["\a"]`, 1, 3},
	{"string_invalid_unicode_escape", `["\uqqqq"]`, 1, 4},
	{"string_invalid_utf8_after_escape", `["\�"]`, 1, 3},
	{"string_invalid-utf-8-in-escape", `["\u�"]`, 1, 4},
	{"string_leading_uescaped_thinspace", `[\u0020"asd"]`, 1, 1},
	{"string_no_quotes_with_bad_escape", `[\n]`, 1, 1},
	{"string_single_quote", `['single quote']`, 1, 1},
	{"string_single_string_no_double_quotes", `abc`, 0, 0},
	{"string_start_escape_unclosed", `["\`, 1, 3},
	{"string_unicode_CapitalU", `"\UA66D"`, 1, 2},
	{"string_with_trailing_garbage", `""x`, 2, 2},
	{"structure_angle_bracket_.", `<.>`, 0, 0},
	{"structure_angle_bracket_null", `[<null>]`, 1, 1},
	{"structure_array_trailing_garbage", `[1]x`, 3, 3},
	{"structure_array_with_extra_array_close", `[1]]`, 3, 3},
	{"structure_array_with_unclosed_string", `["asd]`, 1, 6},
	{"structure_ascii-unicode-identifier", `aå`, 0, 0},
	{"structure_capitalized_True", `[True]`, 1, 1},
	{"structure_close_unopened_array", `1]`, 1, 1},
	{"structure_comma_instead_of_closing_brace", `{"x": true,`, 1, 11},
	{"structure_double_array", `[][]`, 2, 2},
	{"structure_end_array", `]`, 0, 0},
	{"structure_incomplete_UTF8_BOM", `�{}`, 0, 0},
	{"structure_lone-invalid-utf-8", `�`, 0, 0},
	{"structure_null-byte-outside-string", "\x5b\x00\x5d", 1, 1},
	{"structure_number_with_trailing_garbage", `2@`, 1, 1},
	{"structure_object_followed_by_closing_object", `{}}`, 2, 2},
	{"structure_object_unclosed_no_value", `{"":`, 1, 4},
	{"structure_object_with_comment", `{"a":/*comment*/"b"}`, 1, 5},
	{"structure_object_with_trailing_garbage", `{"a": true} "x"`, 12, 12},
	{"structure_open_array_apostrophe", `['`, 1, 1},
	{"structure_open_array_comma", `[,`, 1, 1},
	{"structure_open_array_open_object", `[{`, 1, 2},
	{"structure_open_array_open_string", `["a`, 1, 3},
	{"structure_open_array_string", `["a"`, 1, 4},
	{"structure_open_object_close_array", `{]`, 1, 1},
	{"structure_open_object_comma", `{,`, 1, 1},
	{"structure_open_object_open_array", `{[`, 1, 1},
	{"structure_open_object_open_string", `{"a`, 1, 3},
	{"structure_open_object_string_with_apostrophes", `{'a'`, 1, 1},
	{"structure_open_open", `["\{["\{["\{["\{`, 1, 3},
	{"structure_single_eacute", `�`, 0, 0},
	{"structure_single_star", `*`, 0, 0},
	{"structure_trailing_#", `{"a":"b"}#{}`, 9, 9},
	{"structure_U+2060_word_joined", "\x5b\xe2\x81\xa0\x5d", 1, 1},
	{"structure_uescaped_LF_before_string", `[\u000A""]`, 1, 1},
	{"structure_unclosed_array", `[1`, 1, 2},
	{"structure_unclosed_array_partial_null", `[ false, nul`, 1, 12},
	{"structure_unclosed_array_unfinished_false", `[ true, fals`, 1, 12},
	{"structure_unclosed_array_unfinished_true", `[ false, tru`, 1, 12},
	{"structure_unclosed_object", `{"asd":"asd"`, 1, 12},
	{"structure_unicode-identifier", `å`, 0, 0},
	{"structure_UTF8_BOM_no_data", "\xef\xbb\xbf", 0, 0},
	{"structure_whitespace_formfeed", "\x5b\x0c\x5d", 1, 1},
	{"structure_whitespace_U+2060_word_joiner", "\x5b\xe2\x81\xa0\x5d", 1, 1},
}

func TestConsumeString(t *testing.T) {
	tCases := []struct {
		name     string
		data     string
		expected int
	}{
		{"ascii string", `foo"`, 4},
		{"utf-8 string one char", `ß"`, 3},
		{"utf-8 string multiple chars", `ßßßß"`, 9},
		{"empty string", ``, 0},
		{"non-ending ascii string", `a`, 0},
		{"non-ending utf-8 string", `ß`, 0},
		{"escaped ascii string", "\\b a\"", 5},
		{"escaped utf-8 string", "\\b ß\"", 6},
	}

	for _, tt := range tCases {
		t.Run(tt.name, func(t *testing.T) {
			p := &parserState{}
			got := p.consumeString([]byte(tt.data))
			if got != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, got)
			}
		})
	}
}

func TestConsumeNumber(t *testing.T) {
	tCases := []struct {
		data     string
		expected int
	}{
		{`123`, 3},
		{`123.1`, 5},
		{`123.`, 4},
		{`.123`, 4},
		{`.`, 0},
		{`..`, 0},
		{`e`, 0},
		{`1e1`, 3},
		{`1.1e1`, 5},
		{`.1e1`, 4},
		{"", 0},
		{`"NaN"`, 0},
		{`"Infinity"`, 0},
		{`"-Infinity"`, 0},
		{".0", 2},
		{"0", 1},
		{"-0", 2},
		{"+0", 0},
		{"1", 1},
		{"-1", 2},
		{"00", 2},
		{"-00", 3},
		{"01", 2},
		{"-01", 3},
		{"0i", 1},
		{"-0i", 2},
		{"0f", 1},
		{"-0f", 2},
		{"9876543210", 10},
		{"-9876543210", 11},
		{"9876543210x", 10},
		{"-9876543210x", 11},
		{" 9876543210", 0},
		{"- 9876543210", 0},
		{strings.Repeat("9876543210", 1000), 10000},
		{"-" + strings.Repeat("9876543210", 1000), 1 + 10000},
		{"0.", 2},
		{"-0.", 3},
		{"0e", 0},
		{"-0e", 0},
		{"0E", 0},
		{"-0E", 0},
		{"0.0", 3},
		{"-0.0", 4},
		{"0e0", 3},
		{"-0e0", 4},
		{"0E0", 3},
		{"-0E0", 4},
		{"0.0123456789", 12},
		{"-0.0123456789", 13},
		{"1.f", 2},
		{"-1.f", 3},
		{"1.e", 0},
		{"-1.e", 0},
		{"1e0", 3},
		{"-1e0", 4},
		{"1E0", 3},
		{"-1E0", 4},
		{"1Ex", 0},
		{"-1Ex", 0},
		{"1e-0", 4},
		{"-1e-0", 5},
		{"1e+0", 4},
		{"-1e+0", 5},
		{"1E-0", 4},
		{"-1E-0", 5},
		{"1E+0", 4},
		{"-1E+0", 5},
		{"1E+00500", 8},
		{"-1E+00500", 9},
		{"1E+00500x", 8},
		{"-1E+00500x", 9},
		{"9876543210.0123456789e+01234589x", 31},
		{"-9876543210.0123456789e+01234589x", 32},
		{"1_000_000", 1},
		{"0x12ef", 1},
		{"0x1p-2", 1},
	}

	p := &parserState{}
	for _, tt := range tCases {
		tname := tt.data
		if len(tname) > 10 {
			tname = tname[:10] + "..."
		}
		t.Run(tname, func(t *testing.T) {
			got := p.consumeNumber([]byte(tt.data))
			if got != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, got)
			}
		})
	}
}

func TestConsumeArray(t *testing.T) {
	tCases := []struct {
		name     string
		data     string
		expected int
	}{
		{"empty array", `]`, 1},
		{"empty array spaces", ` ]`, 2},
		{"one int array", `1]`, 2},
		{"one int array spaces", ` 1 ]`, 4},
		{"two ints array", `1,2]`, 4},
		{"two ints array spaces", ` 1 , 2 ]`, 8},
		{"everything array", `[], {}, true, false, null, 1, "abc"]`, 36},
		{"everything array v2", `[1,2,3], {"a":"b"}, true, false, null, 1, "abc"]`, 48},
		{"escaped \"", `"\""]`, 5},
		{"hex", `"\uA66D"]`, 9},
		{"unfinished string", `"\uFFF`, 0},
	}

	p := &parserState{}
	for _, tt := range tCases {
		t.Run(tt.name, func(t *testing.T) {
			got := p.consumeArray([]byte(tt.data), nil, 1)
			if got != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, got)
			}
		})
	}
}

func TestQueryObject(t *testing.T) {
	tCases := []struct {
		json         string
		query        query
		expectedFind bool
	}{{
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("fool")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("afoo")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte(""), []byte("foo")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("bar"), []byte("foo")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "foo"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("bar"), []byte("foo")},
		},
		expectedFind: false,
	}, {
		json: `[{"foo": {"bar": "baz"}}]`,
		query: query{
			SearchPath: [][]byte{[]byte("foo"), []byte("bar")},
		},
		expectedFind: false,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("foo"), []byte("bar")},
		},
		expectedFind: true,
	}, {
		json: `{"foo": {"bar": "baz"}}`,
		query: query{
			SearchPath: [][]byte{[]byte("foo"), []byte("bar")},
			SearchVals: [][]byte{[]byte(`"baz"`)},
		},
		expectedFind: true,
	}, {
		json: `{"foo": {"foo": {"bar": "baz"}}}`,
		query: query{
			SearchPath: [][]byte{[]byte("foo"), []byte("bar")},
			SearchVals: [][]byte{[]byte(`"baz"`)},
		},
		expectedFind: false,
	}}

	for _, tt := range tCases {
		t.Run(tt.json, func(t *testing.T) {
			p := &parserState{}
			p.consumeAny([]byte(tt.json), []query{tt.query}, 0)
			if tt.expectedFind != p.querySatisfied {
				t.Errorf("expectedFind: %v, got: %v", tt.expectedFind, p.querySatisfied)
			}
		})
	}
}

func TestConsumeObject(t *testing.T) {
	tCases := []struct {
		name     string
		data     string
		expected int
	}{
		{"empty object", `}`, 1},
		{"object", `"a":"b"}`, 8},
		{"panic found with fuzz", "\"\":0", 0},
	}

	p := &parserState{}
	for _, tt := range tCases {
		t.Run(tt.name, func(t *testing.T) {
			got := p.consumeObject([]byte(tt.data), nil, 1)
			if got != tt.expected {
				t.Errorf("expected: %v, got: %v", tt.expected, got)
			}
		})
	}
}

func TestConsumeConst(t *testing.T) {
	tCases := []struct {
		b       string
		cnst    string
		expect  int
		inspect int
	}{
		{"", "", 0, 0},
		{"", "true", 0, 0},
		{"true", "", 0, 0},
		{"t", "true", 0, 1},
		{"tr", "true", 0, 2},
		{"tru", "true", 0, 3},
		{"true", "true", 4, 4},
		{"truex", "true", 4, 4},
	}

	for _, tt := range tCases {
		p := &parserState{}
		t.Run(tt.b+" -- "+tt.cnst, func(t *testing.T) {
			got := p.consumeConst([]byte(tt.b), []byte(tt.cnst))
			if got != tt.expect {
				t.Errorf("expected: %v, got %v", tt.expect, got)
			}
			if p.ib != tt.inspect {
				t.Errorf("expected to inspect: %v, got %v", tt.inspect, p.ib)
			}
		})
	}
}

// Truncate inputs at each possible index and test if decoder parses
// the truncated part successfully.
func testTruncating(t *testing.T, jsonString string) {
	t.Helper()
	p := &parserState{}
	for i := 1; i <= len(jsonString); i++ {
		b := scan.Bytes(jsonString[:i])
		b.TrimRWS()
		p.reset()
		_ = p.consumeAny(b, nil, 0)
		if p.ib != len(b) {
			t.Errorf("truncated positives should be fully parsed %v \n"+
				"got: %d want: %d", string(b), p.ib, len(b))
		}
	}
}

func TestPositives(t *testing.T) {
	for _, tt := range positives {
		testTruncating(t, tt.json)
	}
}

func TestPositivesCompacted(t *testing.T) {
	for _, tt := range positives {
		if !tt.stdlib {
			continue
		}
		buf := &bytes.Buffer{}
		if err := json.Compact(buf, []byte(tt.json)); err != nil {
			t.Errorf("Compact should always be successful: %s %s", tt.json, err)
		}
		testTruncating(t, buf.String())
	}
}

func TestPositivesIndented(t *testing.T) {
	indents := [][2]string{
		{"", " "},
		{" ", " "},
		{" ", "\t"},
		{"\t", "\t"},
		{"\t", " \t"},
		{"", "\r\n"},
		{"", " \r\n"},
	}
	for _, tt := range positives {
		if !tt.stdlib {
			continue
		}
		for _, indent := range indents {
			buf := &bytes.Buffer{}
			if err := json.Indent(buf, []byte(tt.json), indent[0], indent[1]); err != nil {
				t.Errorf("Indent should always be successful: %s %s", tt.json, err)
			}
			testTruncating(t, buf.String())
		}
	}
}

func TestNegatives(t *testing.T) {
	p := &parserState{}
	for _, tt := range negatives {
		t.Run(tt.name, func(t *testing.T) {
			p.reset()
			got := p.consumeAny([]byte(tt.json), nil, 0)
			if got != tt.expectParse {
				t.Errorf("unexpected parsed length got: %d want:%d", got, tt.expectParse)
			}
			if p.ib != tt.expectInspect {
				t.Errorf("unexpected inspected length got: %d want:%d\nin:%s", p.ib, tt.expectInspect, tt.json)
			}
		})
	}
}

func TestMaxRecursion(t *testing.T) {
	tCases := []struct {
		maxRecursion    int
		input           string
		expectParsed    int
		expectInspected int
	}{
		{0, `[]`, 2, 2},
		{0, `[[[]]]`, 6, 6},
		{0, strings.Repeat("[", 10000) + strings.Repeat("]", 10000), 20000, 20000},
		{3, `[[[[[]]]]]`, 1, 4}, // max recursion is 3 so we need to inspect 4 opening brackets
	}
	for _, tt := range tCases {
		tname := tt.input
		if len(tname) > 10 {
			tname = tname[:10] + "..."
		}
		t.Run(tname, func(t *testing.T) {
			p := &parserState{
				maxRecursion: tt.maxRecursion,
			}
			got := p.consumeAny([]byte(tt.input), nil, 0)
			if got != tt.expectParsed {
				t.Errorf("parsed: got: %d expected: %d", got, tt.expectParsed)
			}
			if p.ib != tt.expectInspected {
				t.Errorf("inspected: got: %d expected: %d", p.ib, tt.expectInspected)
			}
		})
	}
}

func TestStack(t *testing.T) {
	tCases := []struct {
		name     string
		data     string
		expected string
	}{
		{"empty", ` `, ""},
		{"a string", `"abc"`, ""},
		{"an int", `123`, ""},
		{"true", `true`, ""},
		{"false", `false`, ""},
		// Input must be an incomplete JSON because the stack is popped otherwise.
		{"arr", `[`, "["},
		// Put a § between each segment of the stack.
		{"arrr", `[[`, "[§["},
		{"arrrr", `[[[`, "[§[§["},
		{"arrr popped once", `[[[]`, "[§["},
		{"obj", `{`, ""},
		{"obj key", `{"abc":1`, "abc"},
		{"obj key twice", `{"abc":{"def":1`, "abc§def"},
		{"obj key twice but popped", `{"abc":{"def":1}`, "abc"},
		{"obj key twice and arr", `{"abc":{"def":[`, "abc§def§["},
		{"hacky", `{"abc":{"def[":`, "abc§def["},
	}

	join := func(bs [][]byte) string {
		ret := []string{}
		for _, b := range bs {
			ret = append(ret, string(b))
		}
		return strings.Join(ret, "§")
	}
	for _, tt := range tCases {
		t.Run(tt.name, func(t *testing.T) {
			p := &parserState{}
			p.consumeAny([]byte(tt.data), []query{{}}, 0)
			if got := join(p.currPath); got != tt.expected {
				t.Errorf("expected: %s, got: %s", tt.expected, got)
			}
		})
	}
}

func TestCurrPathBounded(t *testing.T) {
	// currPath is bounded to 128.
	count := 129
	// input has to be an incomplete json, so that currPath does not get popped.
	input := []byte(strings.Repeat("[", count))

	for i := 0; i < 100; i++ {
		Parse(QueryGeo, input)
		// It's not guaranteed that p is the same parser object used by the
		// Parse call above. Reason: go runs tests packages concurrently. If
		// another package calls Parse in tests, that can interfere with parserPool.
		// Running the test several times in loop mitigates that.
		p := parserPool.Get().(*parserState)
		if len(p.currPath) > 128 {
			t.Errorf("expected currPath be purged if >128")
		}
	}
}

var sample = []byte(` { "type": "Feature", "fruit": "Apple", "size": "Large", "color": "Red" } `)

func BenchmarkParse(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _, _, query := Parse(QueryGeo, sample)
		if !query {
			b.Error("query should be satisfied")
		}
	}
}

func BenchmarkJSONStdlibDecoder(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d := json.NewDecoder(bytes.NewReader(sample))
		for {
			_, err := d.Token()
			if err != nil {
				break
			}
		}
	}
}
func BenchmarkJSONOurParser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		p := &parserState{}
		p.consumeAny(sample, nil, 0)
	}
}

func FuzzJson(f *testing.F) {
	for _, p := range positives {
		f.Add([]byte(p.json), true)
	}
	p := &parserState{}
	f.Fuzz(func(t *testing.T, data []byte, reset bool) {
		if reset {
			p.reset()
		}
		p.consumeString(data)
		p.consumeNumber(data)
		p.consumeArray(data, nil, 1)
		p.consumeObject(data, nil, 1)
		p.consumeAny(data, nil, 1)
	})
}
