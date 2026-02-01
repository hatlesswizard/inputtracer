package php

import (
	"context"
	"strings"
	"testing"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

func parsePHP(t *testing.T, code string) (*sitter.Node, []byte) {
	t.Helper()
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())
	src := []byte(code)
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		t.Fatalf("failed to parse PHP: %v", err)
	}
	return tree.RootNode(), src
}

func TestBareSuperglobalDetection(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name       string
		code       string
		wantBare   string // expected bare source name, empty if none
		wantCount  int    // total matches expected for this superglobal
	}{
		{
			name:      "bare $_GET as function argument",
			code:      `<?php $forumids = wpfval( $_GET, 'wpff' );`,
			wantBare:  "$_GET (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_POST in array_merge",
			code:      `<?php $all = array_merge($_POST, $defaults);`,
			wantBare:  "$_POST (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_REQUEST as sole argument",
			code:      `<?php process($_REQUEST);`,
			wantBare:  "$_REQUEST (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_COOKIE as function argument",
			code:      `<?php handle($_COOKIE);`,
			wantBare:  "$_COOKIE (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_FILES as function argument",
			code:      `<?php upload($_FILES);`,
			wantBare:  "$_FILES (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_SERVER as function argument",
			code:      `<?php info($_SERVER);`,
			wantBare:  "$_SERVER (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_GET in foreach",
			code:      `<?php foreach($_GET as $k => $v) {}`,
			wantBare:  "$_GET (bare)",
			wantCount: 1,
		},
		{
			name:      "bare $_POST in assignment",
			code:      `<?php $all = $_POST;`,
			wantBare:  "$_POST (bare)",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			bareCount := 0
			for _, m := range matches {
				if m.SourceType == tt.wantBare {
					bareCount++
				}
			}

			if bareCount == 0 {
				t.Errorf("expected bare match %q but got none; all matches: %v", tt.wantBare, sourceTypes(matches))
			}
		})
	}
}

func TestSubscriptDoesNotProduceBareMatch(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		noBareName string // this bare source should NOT appear
	}{
		{
			name:       "subscript $_GET produces no bare match",
			code:       `<?php $x = $_GET['key'];`,
			noBareName: "$_GET (bare)",
		},
		{
			name:       "subscript $_POST produces no bare match",
			code:       `<?php $y = $_POST['name'];`,
			noBareName: "$_POST (bare)",
		},
		{
			name:       "subscript $_REQUEST produces no bare match",
			code:       `<?php $z = $_REQUEST['action'];`,
			noBareName: "$_REQUEST (bare)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			for _, m := range matches {
				if m.SourceType == tt.noBareName {
					t.Errorf("got unexpected bare match %q for subscript-only code", tt.noBareName)
				}
			}

			// Verify the subscript pattern DID match
			subscriptName := strings.TrimSuffix(tt.noBareName, " (bare)")
			found := false
			for _, m := range matches {
				if m.SourceType == subscriptName {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected subscript match %q but got none; all matches: %v", subscriptName, sourceTypes(matches))
			}
		})
	}
}

func TestMixedBareAndSubscript(t *testing.T) {
	matcher := NewMatcher()

	code := `<?php
$forumids = wpfval( $_GET, 'wpff' );
$search = $_GET['data'];
`
	root, src := parsePHP(t, code)
	matches := matcher.FindSources(root, src)

	bareCount := 0
	subscriptCount := 0
	for _, m := range matches {
		switch m.SourceType {
		case "$_GET (bare)":
			bareCount++
		case "$_GET":
			subscriptCount++
		}
	}

	if bareCount != 1 {
		t.Errorf("expected 1 bare $_GET match, got %d; all matches: %v", bareCount, sourceTypes(matches))
	}
	if subscriptCount != 1 {
		t.Errorf("expected 1 subscript $_GET match, got %d; all matches: %v", subscriptCount, sourceTypes(matches))
	}
}

func sourceTypes(matches []common.Match) []string {
	types := make([]string, len(matches))
	for i, m := range matches {
		types[i] = m.SourceType
	}
	return types
}
