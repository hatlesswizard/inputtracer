package php

import (
	"context"
	"slices"
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
		name      string
		code      string
		wantBare  string // expected bare source name, empty if none
		wantCount int    // total matches expected for this superglobal
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
		name       string
		code       string
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

func TestWPRESTRequestArrayAccess(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantKey  string
		wantType string
	}{
		{
			name:     "$request['param'] single quotes",
			code:     `<?php $value = $request['param_name'];`,
			wantKey:  "param_name",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
		{
			name:     `$request["param"] double quotes`,
			code:     `<?php $value = $request["param_name"];`,
			wantKey:  "param_name",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
		{
			name:     "$req['param'] shorthand variable",
			code:     `<?php $val = $req['id'];`,
			wantKey:  "id",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
		{
			name:     "$api_request['param']",
			code:     `<?php $val = $api_request['slug'];`,
			wantKey:  "slug",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
		{
			name:     "$rest_request['param']",
			code:     `<?php $val = $rest_request['page'];`,
			wantKey:  "page",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
		{
			name:     "multiple $request array accesses",
			code:     `<?php $a = $request['id']; $b = $request['name'];`,
			wantKey:  "id",
			wantType: "$request['...'] (WP_REST_Request ArrayAccess)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType && m.Key == tt.wantKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected match type=%q key=%q but got none; all matches: %v (keys: %v)",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

func TestWPRESTRequestArrayAccessNotFalsePositive(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name string
		code string
	}{
		{
			name: "$response['key'] should not match",
			code: `<?php $val = $response['data'];`,
		},
		{
			name: "$result['key'] should not match",
			code: `<?php $val = $result['id'];`,
		},
		{
			name: "$config['key'] should not match",
			code: `<?php $val = $config['setting'];`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			for _, m := range matches {
				if m.SourceType == "$request['...'] (WP_REST_Request ArrayAccess)" {
					t.Errorf("unexpected WP_REST_Request ArrayAccess match for %q", tt.code)
				}
			}
		})
	}
}

func matchKeys(matches []common.Match) []string {
	keys := make([]string, len(matches))
	for i, m := range matches {
		keys[i] = m.Key
	}
	return keys
}

func sourceTypes(matches []common.Match) []string {
	types := make([]string, len(matches))
	for i, m := range matches {
		types[i] = m.SourceType
	}
	return types
}

func TestDrupalFormAPIPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name      string
		code      string
		wantType  string
		wantKey   string
		wantCount int
	}{
		{
			name:     "form_state getValue with key",
			code:     `<?php $val = $form_state->getValue('field_name');`,
			wantType: "$form_state->getValue()",
			wantKey:  "field_name",
		},
		{
			name:     "form_state getValues all",
			code:     `<?php $values = $form_state->getValues();`,
			wantType: "$form_state->getValues()",
		},
		{
			name:     "form_state getUserInput raw",
			code:     `<?php $input = $form_state->getUserInput();`,
			wantType: "$form_state->getUserInput()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected match type=%q key=%q but got none; all matches: %v",
					tt.wantType, tt.wantKey, sourceTypes(matches))
			}
		})
	}
}

func TestJoomlaInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "JInput getString",
			code:     `<?php $name = $this->input->getString('username');`,
			wantType: "->getString()",
			wantKey:  "username",
		},
		{
			name:     "JInput getInt",
			code:     `<?php $id = $this->input->getInt('id');`,
			wantType: "->getInt()",
			wantKey:  "id",
		},
		{
			name:     "JInput getRaw unfiltered",
			code:     `<?php $raw = $app->input->getRaw('data');`,
			wantType: "->getRaw()",
			wantKey:  "data",
		},
		{
			name:     "JInput getArray multiple values",
			code:     `<?php $data = $this->input->getArray(array('field' => 'STRING'));`,
			wantType: "->getArray()",
		},
		{
			name:     "JInput getCmd",
			code:     `<?php $task = $input->getCmd('task');`,
			wantType: "->getCmd()",
			wantKey:  "task",
		},
		{
			name:     "JRequest getVar legacy",
			code:     `<?php $val = JRequest::getVar('param', '', 'post', 'string');`,
			wantType: "JRequest::getVar()",
			wantKey:  "param",
		},
		{
			name:     "JRequest get legacy",
			code:     `<?php $data = JRequest::get('post');`,
			wantType: "JRequest::get()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected match type=%q key=%q but got none; all matches: %v (keys: %v)",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

func TestOpenCartRequestPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "request->get array access",
			code:     `<?php $id = $this->request->get['product_id'];`,
			wantType: "->get[...]",
			wantKey:  "product_id",
		},
		{
			name:     "request->post array access",
			code:     `<?php $name = $this->request->post['firstname'];`,
			wantType: "->post[...]",
			wantKey:  "firstname",
		},
		{
			name:     "request->server array access",
			code:     `<?php $host = $this->request->server['HTTP_HOST'];`,
			wantType: "->server[...]",
			wantKey:  "HTTP_HOST",
		},
		{
			name:     "request->files array access",
			code:     `<?php $file = $this->request->files['upload'];`,
			wantType: "->files[...]",
			wantKey:  "upload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected match type=%q key=%q but got none; all matches: %v (keys: %v)",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

func TestMagentoStylePatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "getParam single parameter",
			code:     `<?php $id = $this->getRequest()->getParam('id');`,
			wantType: "->getParam()",
			wantKey:  "id",
		},
		{
			name:     "getParams all parameters",
			code:     `<?php $params = $request->getParams();`,
			wantType: "->getParams()",
		},
		{
			name:     "getPost typed POST getter",
			code:     `<?php $name = $request->getPost('customer_name');`,
			wantType: "->getPost()",
			wantKey:  "customer_name",
		},
		{
			name:     "getQuery typed GET getter",
			code:     `<?php $page = $request->getQuery('page');`,
			wantType: "->getQuery()",
			wantKey:  "page",
		},
		{
			name:     "getCookie cookie getter",
			code:     `<?php $token = $request->getCookie('session_token');`,
			wantType: "->getCookie()",
			wantKey:  "session_token",
		},
		{
			name:     "getContent JSON body",
			code:     `<?php $body = $request->getContent();`,
			wantType: "->getContent()",
		},
		{
			name:     "getServer server variable",
			code:     `<?php $host = $request->getServer('HTTP_HOST');`,
			wantType: "->getServer()",
			wantKey:  "HTTP_HOST",
		},
		{
			name:     "getFiles file upload",
			code:     `<?php $file = $request->getFiles('image');`,
			wantType: "->getFiles()",
			wantKey:  "image",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected match type=%q key=%q but got none; all matches: %v (keys: %v)",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

func TestWordPressHookPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name      string
		code      string
		wantCount int
	}{
		{
			name:      "add_action wp_ajax logged-in",
			code:      `<?php add_action('wp_ajax_my_action', 'my_callback');`,
			wantCount: 1,
		},
		{
			name:      "add_action wp_ajax_nopriv logged-out",
			code:      `<?php add_action('wp_ajax_nopriv_my_action', 'my_callback');`,
			wantCount: 1,
		},
		{
			name:      "add_shortcode registration",
			code:      `<?php add_shortcode('my-shortcode', array($this, 'render'));`,
			wantCount: 1,
		},
		{
			name:      "apply_filters the_content",
			code:      `<?php apply_filters('the_content', $content);`,
			wantCount: 1,
		},
		{
			name:      "do_action init",
			code:      `<?php do_action('init');`,
			wantCount: 1,
		},
		{
			name:      "sanitize_text_field wrapping $_POST",
			code:      `<?php $name = sanitize_text_field($_POST['name']);`,
			wantCount: 1,
		},
		{
			name:      "absint wrapping $_GET",
			code:      `<?php $id = absint($_GET['id']);`,
			wantCount: 1,
		},
		{
			name:      "get_post_meta database read",
			code:      `<?php $val = get_post_meta($post_id, 'my_key', true);`,
			wantCount: 1,
		},
		{
			name:      "get_user_meta database read",
			code:      `<?php $data = get_user_meta($user_id, 'profile_data', true);`,
			wantCount: 1,
		},
		{
			name:      "WP_REST_Request get_param",
			code:      `<?php $name = $request->get_param('name');`,
			wantCount: 1,
		},
		{
			name:      "WP_REST_Request get_json_params",
			code:      `<?php $body = $request->get_json_params();`,
			wantCount: 1,
		},
		{
			name:      "WP_REST_Request get_query_params",
			code:      `<?php $params = $request->get_query_params();`,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			if len(matches) < tt.wantCount {
				t.Errorf("expected at least %d source(s) but got %d; all matches: %v",
					tt.wantCount, len(matches), sourceTypes(matches))
			}
		})
	}
}

// ──────────────────────────────────────────────────────────────────────────────
// Tests for Haiku gap-analysis driven patterns (builtins.go)
// ──────────────────────────────────────────────────────────────────────────────

func TestFilterInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "filter_input INPUT_GET",
			code:     `<?php $id = filter_input(INPUT_GET, 'id', FILTER_VALIDATE_INT);`,
			wantType: "filter_input(INPUT_GET)",
			wantKey:  "id",
		},
		{
			name:     "filter_input INPUT_POST",
			code:     `<?php $name = filter_input(INPUT_POST, 'username', FILTER_SANITIZE_STRING);`,
			wantType: "filter_input(INPUT_POST)",
			wantKey:  "username",
		},
		{
			name:     "filter_input INPUT_COOKIE",
			code:     `<?php $tok = filter_input(INPUT_COOKIE, 'session_token');`,
			wantType: "filter_input(INPUT_COOKIE)",
			wantKey:  "session_token",
		},
		{
			name:     "filter_input INPUT_SERVER",
			code:     `<?php $host = filter_input(INPUT_SERVER, 'HTTP_HOST');`,
			wantType: "filter_input(INPUT_SERVER)",
			wantKey:  "HTTP_HOST",
		},
		{
			name:     "filter_input INPUT_ENV",
			code:     `<?php $db = filter_input(INPUT_ENV, 'DATABASE_URL');`,
			wantType: "filter_input(INPUT_ENV)",
			wantKey:  "DATABASE_URL",
		},
		{
			name:     "filter_input_array INPUT_GET",
			code:     `<?php $data = filter_input_array(INPUT_GET, ['page' => FILTER_VALIDATE_INT]);`,
			wantType: "filter_input_array(INPUT_GET)",
		},
		{
			name:     "filter_input_array INPUT_POST",
			code:     `<?php $data = filter_input_array(INPUT_POST, FILTER_SANITIZE_STRING);`,
			wantType: "filter_input_array(INPUT_POST)",
		},
		{
			name:     "filter_input_array INPUT_COOKIE",
			code:     `<?php $cookies = filter_input_array(INPUT_COOKIE);`,
			wantType: "filter_input_array(INPUT_COOKIE)",
		},
		{
			name:     "filter_input_array INPUT_SERVER",
			code:     `<?php $srv = filter_input_array(INPUT_SERVER);`,
			wantType: "filter_input_array(INPUT_SERVER)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected type=%q key=%q; got types=%v keys=%v",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

func TestIssetEmptyPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "isset($_POST['key'])",
			code:     `<?php if (isset($_POST['username'])) { echo "set"; }`,
			wantType: "isset($_*[...])",
		},
		{
			name:     "isset($_GET['id'])",
			code:     `<?php if (isset($_GET['id'])) { $id = $_GET['id']; }`,
			wantType: "isset($_*[...])",
		},
		{
			name:     "isset($_REQUEST['action'])",
			code:     `<?php if (isset($_REQUEST['action'])) { do_action(); }`,
			wantType: "isset($_*[...])",
		},
		{
			name:     "isset($_COOKIE['session'])",
			code:     `<?php if (isset($_COOKIE['session_id'])) { resume(); }`,
			wantType: "isset($_*[...])",
		},
		{
			name:     "isset($_SERVER['HTTP_HOST'])",
			code:     `<?php if (isset($_SERVER['HTTP_HOST'])) { $host = $_SERVER['HTTP_HOST']; }`,
			wantType: "isset($_*[...])",
		},
		{
			name:     "empty($_GET['search'])",
			code:     `<?php if (!empty($_GET['search'])) { doSearch(); }`,
			wantType: "empty($_*[...])",
		},
		{
			name:     "empty($_POST['email'])",
			code:     `<?php if (empty($_POST['email'])) { die('email required'); }`,
			wantType: "empty($_*[...])",
		},
		{
			name:     "array_key_exists('k', $_POST)",
			code:     `<?php if (array_key_exists('role', $_POST)) { $role = $_POST['role']; }`,
			wantType: "array_key_exists('key', $_*)",
		},
		{
			name:     "array_key_exists('k', $_GET)",
			code:     `<?php $ok = array_key_exists('page', $_GET);`,
			wantType: "array_key_exists('key', $_*)",
		},
		{
			name:     "in_array($_POST['role'], $allowed)",
			code:     `<?php if (in_array($_POST['role'], ['admin', 'editor'])) { grant(); }`,
			wantType: "in_array($_*[...], ...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestBuiltinWrapperPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "trim($_POST['name'])",
			code:     `<?php $name = trim($_POST['name']);`,
			wantType: "trim/ltrim/rtrim($_*)",
		},
		{
			name:     "ltrim($_GET['path'])",
			code:     `<?php $path = ltrim($_GET['path'], '/');`,
			wantType: "trim/ltrim/rtrim($_*)",
		},
		{
			name:     "htmlspecialchars($_GET['q'])",
			code:     `<?php echo htmlspecialchars($_GET['q'], ENT_QUOTES, 'UTF-8');`,
			wantType: "htmlspecialchars/htmlentities($_*)",
		},
		{
			name:     "htmlentities($_POST['body'])",
			code:     `<?php $body = htmlentities($_POST['body']);`,
			wantType: "htmlspecialchars/htmlentities($_*)",
		},
		{
			name:     "strip_tags($_POST['html'])",
			code:     `<?php $clean = strip_tags($_POST['html'], '<p><br>');`,
			wantType: "strip_tags/nl2br($_*)",
		},
		{
			name:     "addslashes($_GET['q'])",
			code:     `<?php $safe = addslashes($_GET['search']);`,
			wantType: "addslashes/stripslashes($_*)",
		},
		{
			name:     "strtolower($_POST['email'])",
			code:     `<?php $email = strtolower($_POST['email']);`,
			wantType: "strtolower/strtoupper/mb_str*($_*)",
		},
		{
			name:     "urldecode($_GET['url'])",
			code:     `<?php $url = urldecode($_GET['redirect']);`,
			wantType: "urldecode/rawurldecode($_*)",
		},
		{
			name:     "base64_decode($_POST['data'])",
			code:     `<?php $decoded = base64_decode($_POST['payload']);`,
			wantType: "base64_decode/base64_encode($_*)",
		},
		{
			name:     "json_decode($_POST['json'])",
			code:     `<?php $obj = json_decode($_POST['data'], true);`,
			wantType: "json_decode/json_encode($_*)",
		},
		{
			name:     "intval($_GET['id'])",
			code:     `<?php $id = intval($_GET['id']);`,
			wantType: "intval/floatval/abs($_*)",
		},
		{
			name:     "round($_GET['amount'])",
			code:     `<?php $price = round($_GET['amount'], 2);`,
			wantType: "intval/floatval/abs($_*)",
		},
		{
			name:     "md5($_POST['pass'])",
			code:     `<?php $hash = md5($_POST['password']);`,
			wantType: "md5/sha1/hash($_*)",
		},
		{
			name:     "sha1($_GET['token'])",
			code:     `<?php $t = sha1($_GET['token']);`,
			wantType: "md5/sha1/hash($_*)",
		},
		{
			name:     "preg_replace pattern on $_POST",
			code:     `<?php $clean = preg_replace('/[^a-z]/', '', $_POST['name']);`,
			wantType: "preg_replace/str_replace on $_*",
		},
		{
			name:     "sprintf with $_GET arg",
			code:     `<?php $msg = sprintf("Hello %s", $_GET['name']);`,
			wantType: "sprintf/printf with $_*",
		},
		{
			name:     "mb_strtolower($_POST['text'])",
			code:     `<?php $t = mb_strtolower($_POST['text'], 'UTF-8');`,
			wantType: "strtolower/strtoupper/mb_str*($_*)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestStreamInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "fgets(STDIN)",
			code:     `<?php $line = fgets(STDIN);`,
			wantType: "fgets(STDIN)",
		},
		{
			name:     "fread(STDIN, 1024)",
			code:     `<?php $data = fread(STDIN, 1024);`,
			wantType: "fread(STDIN, ...)",
		},
		{
			name:     "stream_get_contents(STDIN)",
			code:     `<?php $input = stream_get_contents(STDIN);`,
			wantType: "stream_get_contents(STDIN)",
		},
		{
			name:     "file_get_contents php://stdin",
			code:     `<?php $data = file_get_contents('php://stdin');`,
			wantType: "file_get_contents(php://stdin)",
		},
		{
			name:     "readline() interactive input",
			code:     `<?php $answer = readline('Enter your name: ');`,
			wantType: "readline()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestCliInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "$argv subscript access",
			code:     `<?php $cmd = $argv[1];`,
			wantType: "$argv[N]",
		},
		{
			name:     "$argv[2] second argument",
			code:     `<?php $path = $argv[2];`,
			wantType: "$argv[N]",
		},
		{
			name:     "$argc argument count",
			code:     `<?php if ($argc > 1) { process($argv[1]); }`,
			wantType: "$argc",
		},
		{
			name:     "getopt short options",
			code:     `<?php $opts = getopt('f:v');`,
			wantType: "getopt()",
		},
		{
			name:     "getopt long options",
			code:     `<?php $opts = getopt('', ['file:', 'verbose']);`,
			wantType: "getopt()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestCastInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "(int) cast on $_GET",
			code:     `<?php $id = (int)$_GET['id'];`,
			wantType: "(int)$_*",
		},
		{
			name:     "(int) cast on $_POST with spaces",
			code:     `<?php $page = ( int ) $_POST['page'];`,
			wantType: "(int)$_*",
		},
		{
			name:     "(float) cast on $_REQUEST",
			code:     `<?php $price = (float)$_REQUEST['price'];`,
			wantType: "(float)$_*",
		},
		{
			name:     "(double) cast on $_GET",
			code:     `<?php $val = (double)$_GET['value'];`,
			wantType: "(float)$_*",
		},
		{
			name:     "(string) cast on $_POST",
			code:     `<?php $name = (string)$_POST['name'];`,
			wantType: "(string)$_*",
		},
		{
			name:     "(bool) cast on $_GET flag",
			code:     `<?php $enabled = (bool)$_GET['enabled'];`,
			wantType: "(bool)$_*",
		},
		{
			name:     "(boolean) cast on $_POST",
			code:     `<?php $flag = (boolean)$_POST['active'];`,
			wantType: "(bool)$_*",
		},
		{
			name:     "(array) cast on $_POST items",
			code:     `<?php $items = (array)$_POST['items'];`,
			wantType: "(array)$_*",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestDynamicInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "extract($_POST) without flags",
			code:     `<?php extract($_POST);`,
			wantType: "extract($_POST)",
		},
		{
			name:     "extract($_GET) without flags",
			code:     `<?php extract($_GET);`,
			wantType: "extract($_GET)",
		},
		{
			name:     "extract($_REQUEST) without flags",
			code:     `<?php extract($_REQUEST);`,
			wantType: "extract($_REQUEST)",
		},
		{
			name:     "extract($_POST) with EXTR_PREFIX_ALL flag",
			code:     `<?php extract($_POST, EXTR_PREFIX_ALL, 'post');`,
			wantType: "extract($_*) with flags",
		},
		{
			name:     "variable variable $$_GET",
			code:     `<?php $varname = $$_GET['var'];`,
			wantType: "$$_GET[...] (variable variables)",
		},
		{
			name:     "variable variable $$_POST",
			code:     `<?php $$_POST['field'] = 'value';`,
			wantType: "$$_GET[...] (variable variables)",
		},
		{
			name:     "list() = $_POST value",
			code:     `<?php list($a, $b) = $_POST['pair'];`,
			wantType: "list(...) = $_*[...]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestLegacyGlobalPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "$HTTP_GET_VARS legacy global",
			code:     `<?php $id = $HTTP_GET_VARS['id'];`,
			wantType: "$HTTP_GET_VARS",
		},
		{
			name:     "$HTTP_POST_VARS legacy global",
			code:     `<?php $name = $HTTP_POST_VARS['name'];`,
			wantType: "$HTTP_POST_VARS",
		},
		{
			name:     "$HTTP_COOKIE_VARS legacy global",
			code:     `<?php $tok = $HTTP_COOKIE_VARS['token'];`,
			wantType: "$HTTP_COOKIE_VARS",
		},
		{
			name:     "$HTTP_SERVER_VARS legacy global",
			code:     `<?php $host = $HTTP_SERVER_VARS['HTTP_HOST'];`,
			wantType: "$HTTP_SERVER_VARS",
		},
		{
			name:     "$HTTP_RAW_POST_DATA legacy global",
			code:     `<?php $body = $HTTP_RAW_POST_DATA;`,
			wantType: "$HTTP_RAW_POST_DATA",
		},
		{
			name:     "parse_str on $_SERVER[QUERY_STRING]",
			code:     `<?php parse_str($_SERVER['QUERY_STRING'], $params);`,
			wantType: "parse_str($_SERVER[QUERY_STRING])",
		},
		{
			name:     "mb_parse_str on $_SERVER[QUERY_STRING]",
			code:     `<?php mb_parse_str($_SERVER['QUERY_STRING'], $params);`,
			wantType: "mb_parse_str($_SERVER[QUERY_STRING])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestArrayOperationPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
	}{
		{
			name:     "array_map over $_POST",
			code:     `<?php $escaped = array_map('htmlspecialchars', $_POST);`,
			wantType: "array_map($fn, $_*)",
		},
		{
			name:     "array_filter on $_GET",
			code:     `<?php $filtered = array_filter($_GET, 'strlen');`,
			wantType: "array_filter($_*)",
		},
		{
			name:     "array_values of $_POST",
			code:     `<?php $vals = array_values($_POST);`,
			wantType: "array_values($_*)",
		},
		{
			name:     "array_keys of $_GET",
			code:     `<?php $keys = array_keys($_GET);`,
			wantType: "array_keys($_*)",
		},
		{
			name:     "implode over $_GET array",
			code:     `<?php $csv = implode(',', $_GET['items']);`,
			wantType: "implode($sep, $_*)",
		},
		{
			name:     "join over $_POST array",
			code:     `<?php $str = join(' ', $_POST['tags']);`,
			wantType: "implode($sep, $_*)",
		},
		{
			name:     "explode superglobal value",
			code:     `<?php $parts = explode(',', $_GET['ids']);`,
			wantType: "explode($sep, $_*[...])",
		},
		{
			name:     "array_merge starting with $_POST",
			code:     `<?php $all = array_merge($_POST, $defaults);`,
			wantType: "array_merge($_*,...)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected type=%q; got types=%v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

func TestEnvironmentInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "getenv with key",
			code:     `<?php $db = getenv('DATABASE_URL');`,
			wantType: "getenv('VAR')",
			wantKey:  "DATABASE_URL",
		},
		{
			name:     "getenv bare call",
			code:     `<?php $path = getenv('PATH');`,
			wantType: "getenv('VAR')",
			wantKey:  "PATH",
		},
		{
			name:     "$_ENV bare array",
			code:     `<?php $all = $_ENV;`,
			wantType: "$_ENV (bare)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected type=%q key=%q; got types=%v keys=%v",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

// TestNoFalsePositivesOnNonInputCode verifies that purely internal PHP code
// (no superglobals, no external input) does not match any input source.
func TestNoFalsePositivesOnNonInputCode(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name string
		code string
	}{
		{
			name: "pure local variable operations",
			code: `<?php $x = 42; $y = trim($x); $z = strtolower($y);`,
		},
		{
			name: "array operations on local arrays",
			code: `<?php $arr = [1, 2, 3]; $filtered = array_filter($arr, 'is_int');`,
		},
		{
			name: "isset on local variable",
			code: `<?php $x = null; if (isset($x)) { echo $x; }`,
		},
		{
			name: "empty on local variable",
			code: `<?php $list = []; if (empty($list)) { echo 'empty'; }`,
		},
		{
			name: "json_decode on local string",
			code: `<?php $data = json_decode('{"key":"val"}', true);`,
		},
		{
			name: "intval on literal",
			code: `<?php $n = intval("42");`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)
			if len(matches) > 0 {
				t.Errorf("expected 0 matches for non-input code, got %d: %v",
					len(matches), sourceTypes(matches))
			}
		})
	}
}

// TestKeyExtractionFromBuiltinWrappers verifies that the key extractor correctly
// pulls the array key from superglobal patterns inside builtin function calls.
func TestKeyExtractionFromBuiltinWrappers(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name    string
		code    string
		wantKey string
	}{
		{
			name:    "trim extracts key from $_POST",
			code:    `<?php $name = trim($_POST['username']);`,
			wantKey: "username",
		},
		{
			name:    "htmlspecialchars extracts key from $_GET",
			code:    `<?php $q = htmlspecialchars($_GET['search']);`,
			wantKey: "search",
		},
		{
			name:    "intval extracts key from $_GET",
			code:    `<?php $id = intval($_GET['id']);`,
			wantKey: "id",
		},
		{
			name:    "base64_decode extracts key from $_POST",
			code:    `<?php $tok = base64_decode($_POST['token']);`,
			wantKey: "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.Key == tt.wantKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected key=%q; got keys=%v (types=%v)",
					tt.wantKey, matchKeys(matches), sourceTypes(matches))
			}
		})
	}
}

// TestSessionPatterns verifies $_SESSION access is detected.
func TestSessionPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name     string
		code     string
		wantType string
		wantKey  string
	}{
		{
			name:     "$_SESSION subscript read",
			code:     `<?php $uid = $_SESSION['user_id'];`,
			wantType: "$_SESSION[...] (user-controlled)",
			wantKey:  "user_id",
		},
		{
			name:     "$_SESSION role read",
			code:     `<?php $role = $_SESSION['role'];`,
			wantType: "$_SESSION[...] (user-controlled)",
			wantKey:  "role",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					if tt.wantKey == "" || m.Key == tt.wantKey {
						found = true
						break
					}
				}
			}
			if !found {
				t.Errorf("expected type=%q key=%q; got types=%v keys=%v",
					tt.wantType, tt.wantKey, sourceTypes(matches), matchKeys(matches))
			}
		})
	}
}

// TestIssetKeyExtraction verifies that isset() and empty() patterns correctly
// extract the array key from the superglobal subscript.
func TestIssetKeyExtraction(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name    string
		code    string
		wantKey string
	}{
		{
			name:    "isset extracts key from $_POST",
			code:    `<?php if (isset($_POST['email'])) {}`,
			wantKey: "email",
		},
		{
			name:    "isset extracts key from $_GET",
			code:    `<?php if (isset($_GET['page'])) {}`,
			wantKey: "page",
		},
		{
			name:    "empty extracts key from $_REQUEST",
			code:    `<?php if (!empty($_REQUEST['token'])) {}`,
			wantKey: "token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.Key == tt.wantKey {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected key=%q; got keys=%v", tt.wantKey, matchKeys(matches))
			}
		})
	}
}

// TestRealWorldInputPatterns runs against representative real-world code snippets
// found in the testapps PHP code during gap analysis.
func TestRealWorldInputPatterns(t *testing.T) {
	matcher := NewMatcher()

	tests := []struct {
		name      string
		code      string
		wantTypes []string
	}{
		{
			name: "WordPress form handler typical pattern",
			code: `<?php
if (isset($_POST['nonce']) && wp_verify_nonce($_POST['nonce'], 'my_action')) {
    $name = sanitize_text_field($_POST['name']);
    $email = sanitize_email($_POST['email']);
    $message = wp_kses_post($_POST['message']);
}`,
			wantTypes: []string{"isset($_*[...])", "$_POST"},
		},
		{
			name: "PHP filter_input typical usage",
			code: `<?php
$page = filter_input(INPUT_GET, 'page', FILTER_VALIDATE_INT, ['options' => ['default' => 1]]);
$query = filter_input(INPUT_GET, 'q', FILTER_SANITIZE_STRING);
$token = filter_input(INPUT_POST, 'csrf_token', FILTER_SANITIZE_STRING);`,
			wantTypes: []string{"filter_input(INPUT_GET)", "filter_input(INPUT_POST)"},
		},
		{
			name: "CLI script typical pattern",
			code: `<?php
if ($argc < 2) {
    echo "Usage: script.php <filename>\n";
    exit(1);
}
$filename = $argv[1];
$opts = getopt('v', ['verbose', 'output:']);`,
			wantTypes: []string{"$argc", "$argv[N]", "getopt()"},
		},
		{
			name: "CGI/REST raw body pattern",
			code: `<?php
$body = file_get_contents('php://input');
$data = json_decode($body, true);
$headers = getallheaders();`,
			wantTypes: []string{"php://input (file_get_contents)", "getallheaders()"},
		},
		{
			name: "Typical form validation with multiple checks",
			code: `<?php
if (!isset($_POST['action'])) { die(); }
$action = trim($_POST['action']);
if (empty($_POST['name'])) { die('name required'); }
$id = (int)$_GET['id'];
if (array_key_exists('redirect', $_POST)) {
    $redirect = urldecode($_POST['redirect']);
}`,
			wantTypes: []string{
				"isset($_*[...])",
				"trim/ltrim/rtrim($_*)",
				"empty($_*[...])",
				"(int)$_*",
				"array_key_exists('key', $_*)",
				"urldecode/rawurldecode($_*)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			foundTypes := make(map[string]bool)
			for _, m := range matches {
				foundTypes[m.SourceType] = true
			}

			for _, wantType := range tt.wantTypes {
				if !foundTypes[wantType] {
					t.Errorf("expected type=%q to be found; got types=%v", wantType, sourceTypes(matches))
				}
			}
		})
	}
}

// TestHTTPAuthPatterns verifies that HTTP authentication credential patterns are
// correctly matched. PHP exposes Basic/Digest auth credentials through special
// $_SERVER keys that differ from the generic HTTP header superglobal access.
func TestHTTPAuthPatterns(t *testing.T) {
	matcher := NewMatcher()
	tests := []struct {
		name      string
		code      string
		wantTypes []string
	}{
		{
			name:      "$_SERVER['PHP_AUTH_USER'] — Basic auth username",
			code:      `<?php $user = $_SERVER['PHP_AUTH_USER'];`,
			wantTypes: []string{"$_SERVER['PHP_AUTH_USER']"},
		},
		{
			name:      "$_SERVER['PHP_AUTH_PW'] — Basic auth password",
			code:      `<?php $pass = $_SERVER['PHP_AUTH_PW'];`,
			wantTypes: []string{"$_SERVER['PHP_AUTH_PW']"},
		},
		{
			name:      "$_SERVER['PHP_AUTH_DIGEST'] — Digest auth header",
			code:      `<?php $digest = $_SERVER['PHP_AUTH_DIGEST'];`,
			wantTypes: []string{"$_SERVER['PHP_AUTH_DIGEST']"},
		},
		{
			name:      "$_SERVER['HTTP_AUTHORIZATION'] — raw Authorization header",
			code:      `<?php $auth = $_SERVER['HTTP_AUTHORIZATION'];`,
			wantTypes: []string{"$_SERVER['HTTP_AUTHORIZATION']"},
		},
		{
			name:      "isset check on PHP_AUTH_USER",
			code:      `<?php if (isset($_SERVER['PHP_AUTH_USER']) && isset($_SERVER['PHP_AUTH_PW'])) { $u = $_SERVER['PHP_AUTH_USER']; $p = $_SERVER['PHP_AUTH_PW']; }`,
			wantTypes: []string{"$_SERVER['PHP_AUTH_USER']", "$_SERVER['PHP_AUTH_PW']"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			foundTypes := make(map[string]bool)
			for _, m := range matches {
				foundTypes[m.SourceType] = true
			}
			for _, want := range tt.wantTypes {
				if !foundTypes[want] {
					t.Errorf("expected pattern %q to be matched; got patterns: %v", want, sourceTypes(matches))
				}
			}
		})
	}
}

// TestFileUploadPatterns verifies that PHP file upload processing functions
// (move_uploaded_file, is_uploaded_file) are correctly matched as FILE input sources.
func TestFileUploadPatterns(t *testing.T) {
	matcher := NewMatcher()
	tests := []struct {
		name      string
		code      string
		wantTypes []string
	}{
		{
			name: "move_uploaded_file with $_FILES",
			code: `<?php move_uploaded_file($_FILES['userfile']['tmp_name'], $dest);`,
			wantTypes: []string{"move_uploaded_file($_FILES[...])"},
		},
		{
			name: "is_uploaded_file with $_FILES",
			code: `<?php if (is_uploaded_file($_FILES['file']['tmp_name'])) { echo 'ok'; }`,
			wantTypes: []string{"is_uploaded_file($_FILES[...])"},
		},
		{
			name:      "move_uploaded_file has File label",
			code:      `<?php move_uploaded_file($_FILES['img']['tmp_name'], '/uploads/img.jpg');`,
			wantTypes: []string{"move_uploaded_file($_FILES[...])"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			foundTypes := make(map[string]bool)
			for _, m := range matches {
				foundTypes[m.SourceType] = true
			}
			for _, want := range tt.wantTypes {
				if !foundTypes[want] {
					t.Errorf("expected pattern %q to be matched; got patterns: %v", want, sourceTypes(matches))
				}
			}
		})
	}
}

// TestCommonInputLabel verifies that the LabelUserInput label is present on
// key patterns to ensure downstream analysis can identify the source category.
func TestCommonInputLabel(t *testing.T) {
	matcher := NewMatcher()

	labelTests := []struct {
		name      string
		code      string
		wantLabel common.InputLabel
	}{
		{
			name:      "isset($_POST) has UserInput label",
			code:      `<?php if (isset($_POST['x'])) {}`,
			wantLabel: common.LabelUserInput,
		},
		{
			name:      "trim($_GET) has UserInput label",
			code:      `<?php $v = trim($_GET['q']);`,
			wantLabel: common.LabelUserInput,
		},
		{
			name:      "(int)$_GET has UserInput label",
			code:      `<?php $n = (int)$_GET['id'];`,
			wantLabel: common.LabelUserInput,
		},
		{
			name:      "$argv[1] has CLI label",
			code:      `<?php $f = $argv[1];`,
			wantLabel: common.LabelCLI,
		},
		{
			name:      "filter_input(INPUT_GET) has HTTPGet label",
			code:      `<?php $v = filter_input(INPUT_GET, 'key');`,
			wantLabel: common.LabelHTTPGet,
		},
		{
			name:      "$HTTP_GET_VARS has HTTPGet label",
			code:      `<?php $v = $HTTP_GET_VARS['id'];`,
			wantLabel: common.LabelHTTPGet,
		},
	}

	for _, tt := range labelTests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if slices.Contains(m.Labels, tt.wantLabel) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected label %q on some match; got matches=%v",
					tt.wantLabel, sourceTypes(matches))
			}
		})
	}
}

// TestRequestFactoryPatterns verifies that static factory methods that create
// request objects from PHP superglobals are correctly detected as input sources.
func TestRequestFactoryPatterns(t *testing.T) {
	matcher := NewMatcher()
	tests := []struct {
		name      string
		code      string
		wantType  string
	}{
		{
			name:     "Symfony Request::createFromGlobals()",
			code:     `<?php $request = Request::createFromGlobals();`,
			wantType: "Request::createFromGlobals()",
		},
		{
			name:     "PSR-7 ServerRequestFactory::fromGlobals()",
			code:     `<?php $request = ServerRequestFactory::fromGlobals();`,
			wantType: "ServerRequestFactory::fromGlobals()",
		},
		{
			name:     "Generic ::fromGlobals() factory",
			code:     `<?php $request = RequestFactory::fromGlobals();`,
			wantType: "::fromGlobals() (generic PSR-7 factory)",
		},
		{
			name:     "PSR-7 instance ->fromGlobals() method",
			code:     `<?php $request = $serverRequestFactory->fromGlobals();`,
			wantType: "->fromGlobals() (PSR-7 instance factory)",
		},
		{
			name:     "PSR-17 createStreamFromFile php://input",
			code:     `<?php $stream = $psr17Factory->createStreamFromFile('php://input');`,
			wantType: "->createStreamFromFile('php://input')",
		},
		{
			name:     "PSR-7 chained ->create()->fromGlobals()",
			code:     `<?php $req = ServerRequestFactory::create()->fromGlobals();`,
			wantType: "->fromGlobals() (PSR-7 instance factory)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected pattern %q; got: %v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}

// TestSymfonyParameterBagPatterns verifies that Symfony HttpFoundation ParameterBag
// chained property access patterns are detected as input sources.
func TestSymfonyParameterBagPatterns(t *testing.T) {
	matcher := NewMatcher()
	tests := []struct {
		name      string
		code      string
		wantType  string
	}{
		{
			name:     "->query->get() — GET parameter",
			code:     `<?php $val = $request->query->get('page');`,
			wantType: "->query->get()",
		},
		{
			name:     "->query->all() — all GET parameters",
			code:     `<?php $vals = $request->query->all();`,
			wantType: "->query->get()",
		},
		{
			name:     "->request->get() — POST parameter",
			code:     `<?php $val = $request->request->get('name');`,
			wantType: "->request->get()",
		},
		{
			name:     "->headers->get() — HTTP header",
			code:     `<?php $auth = $request->headers->get('Authorization');`,
			wantType: "->headers->get()",
		},
		{
			name:     "->cookies->get() — cookie value",
			code:     `<?php $token = $request->cookies->get('session_id');`,
			wantType: "->cookies->get()",
		},
		{
			name:     "->server->get() — server variable",
			code:     `<?php $method = $request->server->get('REQUEST_METHOD');`,
			wantType: "->server->get()",
		},
		{
			name:     "->files->get() — uploaded file",
			code:     `<?php $file = $request->files->get('upload');`,
			wantType: "->files->get()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, src := parsePHP(t, tt.code)
			matches := matcher.FindSources(root, src)

			found := false
			for _, m := range matches {
				if m.SourceType == tt.wantType {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected pattern %q; got: %v", tt.wantType, sourceTypes(matches))
			}
		})
	}
}
