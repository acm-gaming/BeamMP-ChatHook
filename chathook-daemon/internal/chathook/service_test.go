package chathook

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestCleanseString(t *testing.T) {
	input := "^2Hello ^rworld^"
	if got := cleanseString(input); got != "Hello world" {
		t.Fatalf("unexpected cleaned string: %q", got)
	}
}

func TestCutServerName(t *testing.T) {
	input := "サーバー" + strings.Repeat("x", 100)
	output := cutServerName(input)
	if len([]rune(output)) != maxServerNameRunes {
		t.Fatalf("expected %d runes, got %d", maxServerNameRunes, len([]rune(output)))
	}
	if !strings.HasSuffix(output, "...") {
		t.Fatalf("expected ellipsis suffix, got %q", output)
	}
}

func TestIsGuest(t *testing.T) {
	if !isGuest("guest1234567") {
		t.Fatalf("expected guest to be detected")
	}
	if isGuest("player123") {
		t.Fatalf("did not expect guest detection")
	}
}

func TestCountryFlagEmoji(t *testing.T) {
	if got := countryFlagEmoji("de"); got != "🇩🇪" {
		t.Fatalf("unexpected flag: %q", got)
	}
	if got := countryFlagEmoji(""); got != "" {
		t.Fatalf("expected empty output for empty code, got %q", got)
	}
}

func TestDecodePacket(t *testing.T) {
	wire := packet{
		ServerName:  "Test",
		PlayerCount: 1,
		PlayerMax:   4,
		PlayerDif:   0,
		Version:     ProtocolVersion,
		Contents: []content{
			{Type: 2, Content: json.RawMessage(`{}`)},
		},
	}
	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	decoded, err := decodePacket([]byte(encoded))
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if decoded.ServerName != "Test" {
		t.Fatalf("unexpected server name %q", decoded.ServerName)
	}
}

func TestDecodePacketMissingFields(t *testing.T) {
	wire := map[string]any{
		"server_name":  "Test",
		"player_count": 1,
		"player_max":   4,
		"version":      ProtocolVersion,
		"contents":     []map[string]any{{"type": 2, "content": map[string]any{}}},
	}
	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if _, err := decodePacket([]byte(encoded)); err == nil {
		t.Fatalf("expected error when player_dif is missing")
	}
}

func TestDecodePacketRequiresContents(t *testing.T) {
	wire := map[string]any{
		"server_name":  "Test",
		"player_count": 1,
		"player_max":   4,
		"player_dif":   0,
		"version":      ProtocolVersion,
	}
	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if _, err := decodePacket([]byte(encoded)); err == nil {
		t.Fatalf("expected error when contents is missing")
	}
}

func TestDecodePacketAllContentTypes(t *testing.T) {
	contents := []content{
		{Type: 1, Content: json.RawMessage(`{"player_name":"Alice","chat_message":"hi"}`)},
		{Type: 2, Content: json.RawMessage(`{}`)},
		{Type: 3, Content: json.RawMessage(`{"player_name":"Bob","ip":"127.0.0.1"}`)},
		{Type: 4, Content: json.RawMessage(`{"player_name":"Cara","early":false}`)},
		{Type: 5, Content: json.RawMessage(`{}`)},
		{Type: 6, Content: json.RawMessage(`{"script_ref":"","chat_message":"ok"}`)},
		{Type: 7, Content: json.RawMessage(`{"player_name":"Dee"}`)},
		{Type: 8, Content: json.RawMessage(`{"script_ref":"","chat_message":"ok"}`)},
	}

	wire := packet{
		ServerName:  "Test",
		PlayerCount: 1,
		PlayerMax:   4,
		PlayerDif:   0,
		Version:     ProtocolVersion,
		Contents:    contents,
	}
	data, err := json.Marshal(wire)
	if err != nil {
		t.Fatalf("marshal packet: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	decoded, err := decodePacket([]byte(encoded))
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if len(decoded.Contents) != len(contents) {
		t.Fatalf("expected %d contents, got %d", len(contents), len(decoded.Contents))
	}
}

func TestDecodePlayerLeftRequiresEarly(t *testing.T) {
	item := content{Type: 4, Content: json.RawMessage(`{"player_name":"Test"}`)}
	if _, err := decodePlayerLeft(item); err == nil {
		t.Fatalf("expected error when early flag is missing")
	}
}

func TestDecodeChatMessageMissingFields(t *testing.T) {
	item := content{Type: 1, Content: json.RawMessage(`{"player_name":"Test"}`)}
	if _, err := decodeChatMessage(item); err == nil {
		t.Fatalf("expected error when chat_message is missing")
	}
}

func TestDecodeChatMessageSanitizes(t *testing.T) {
	item := content{Type: 1, Content: json.RawMessage(`{"player_name":"Test","chat_message":"@hi http://test https://example discord.gg/abc"}`)}
	msg, err := decodeChatMessage(item)
	if err != nil {
		t.Fatalf("decode chat: %v", err)
	}
	if msg.ChatMessage != "hi test example discord-gg/abc" {
		t.Fatalf("unexpected sanitized message %q", msg.ChatMessage)
	}
}

func TestDecodeScriptMessageMissingFields(t *testing.T) {
	item := content{Type: 6, Content: json.RawMessage(`{"chat_message":"test"}`)}
	if _, err := decodeScriptMessage(item); err == nil {
		t.Fatalf("expected error when script_ref is missing")
	}
}

func TestDecodeScriptMessageAllowsEmptyRef(t *testing.T) {
	item := content{Type: 6, Content: json.RawMessage(`{"script_ref":"","chat_message":"hi"}`)}
	msg, err := decodeScriptMessage(item)
	if err != nil {
		t.Fatalf("decode script: %v", err)
	}
	if msg.ScriptRef != "" {
		t.Fatalf("expected empty script_ref")
	}
}

func TestDecodePlayerJoiningValid(t *testing.T) {
	item := content{Type: 7, Content: json.RawMessage(`{"player_name":"Test"}`)}
	if _, err := decodePlayerJoining(item); err != nil {
		t.Fatalf("decode joining: %v", err)
	}
}

func TestDecodePlayerJoinValid(t *testing.T) {
	item := content{Type: 3, Content: json.RawMessage(`{"player_name":"Test","ip":"127.0.0.1"}`)}
	if _, err := decodePlayerJoin(item); err != nil {
		t.Fatalf("decode join: %v", err)
	}
}

func TestDecodePlayerLeftValid(t *testing.T) {
	item := content{Type: 4, Content: json.RawMessage(`{"player_name":"Test","early":false}`)}
	if _, err := decodePlayerLeft(item); err != nil {
		t.Fatalf("decode left: %v", err)
	}
}

func TestChatRateLimiterDisabled(t *testing.T) {
	limiter := newChatRateLimiter(0, 0, nil)
	for i := 0; i < 100; i++ {
		if !limiter.Allow("ServerA", "PlayerA") {
			t.Fatalf("expected disabled limiter to allow all messages")
		}
	}
}

func TestChatRateLimiterPerPlayerPerServer(t *testing.T) {
	now := time.Unix(1000, 0)
	limiter := newChatRateLimiter(2, 10*time.Second, func() time.Time { return now })

	if !limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("first message should pass")
	}
	if !limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("second message should pass")
	}
	if limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("third message in same window should be limited")
	}

	if !limiter.Allow("ServerA", "PlayerB") {
		t.Fatalf("different player should have independent quota")
	}
	if !limiter.Allow("ServerB", "PlayerA") {
		t.Fatalf("same player on different server should have independent quota")
	}
}

func TestChatRateLimiterWindowReset(t *testing.T) {
	now := time.Unix(2000, 0)
	limiter := newChatRateLimiter(1, 5*time.Second, func() time.Time { return now })

	if !limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("first message should pass")
	}
	if limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("second message in same window should be limited")
	}

	now = now.Add(6 * time.Second)
	if !limiter.Allow("ServerA", "PlayerA") {
		t.Fatalf("message after window reset should pass")
	}
}
