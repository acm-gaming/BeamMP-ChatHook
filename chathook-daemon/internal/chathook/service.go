package chathook

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/biter777/countries"
	"github.com/charmbracelet/log"
)

const (
	Version         = 1
	ProtocolVersion = 4

	defaultForumBaseURL = "https://forum.beammp.com"
	defaultIPAPIBaseURL = "http://ip-api.com"
	maxUDPPacketSize    = 64000
	maxServerNameRunes  = 80
)

type Service struct {
	config       Config
	logger       *log.Logger
	httpClient   *http.Client
	forumBaseURL string
	ipAPIBaseURL string

	profileCache map[string]string
	mu           sync.Mutex
}

type packet struct {
	ServerName  string    `json:"server_name"`
	PlayerCount int       `json:"player_count"`
	PlayerMax   int       `json:"player_max"`
	PlayerDif   int       `json:"player_dif"`
	Version     uint8     `json:"version"`
	Contents    []content `json:"contents"`
}

type packetWire struct {
	ServerName  *string   `json:"server_name"`
	PlayerCount *int      `json:"player_count"`
	PlayerMax   *int      `json:"player_max"`
	PlayerDif   *int      `json:"player_dif"`
	Version     *uint8    `json:"version"`
	Contents    []content `json:"contents"`
}

type content struct {
	Type    int             `json:"type"`
	Content json.RawMessage `json:"content"`
}

type chatMessage struct {
	PlayerName  string `json:"player_name"`
	ChatMessage string `json:"chat_message"`
}

type scriptMessage struct {
	ScriptRef   string `json:"script_ref"`
	ChatMessage string `json:"chat_message"`
}

type playerJoining struct {
	PlayerName string `json:"player_name"`
}

type playerJoin struct {
	PlayerName string `json:"player_name"`
	IP         string `json:"ip"`
}

type playerLeft struct {
	PlayerName string `json:"player_name"`
	Early      bool   `json:"early"`
}

type webhookPayload struct {
	Username  string         `json:"username,omitempty"`
	AvatarURL string         `json:"avatar_url,omitempty"`
	Content   string         `json:"content,omitempty"`
	Embeds    []webhookEmbed `json:"embeds,omitempty"`
}

type webhookEmbed struct {
	Thumbnail *webhookThumbnail `json:"thumbnail,omitempty"`
	Color     uint32            `json:"color,omitempty"`
	Fields    []webhookField    `json:"fields,omitempty"`
}

type webhookThumbnail struct {
	URL string `json:"url"`
}

type webhookField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type forumResponse struct {
	User struct {
		AvatarTemplate string `json:"avatar_template"`
	} `json:"user"`
}

type ipAPIResponse struct {
	Status      string `json:"status"`
	CountryCode string `json:"countryCode"`
	Proxy       bool   `json:"proxy"`
	Hosting     bool   `json:"hosting"`
}

var chatSanitizer = strings.NewReplacer(
	"@", "",
	"http://", "",
	"https://", "",
	"discord.gg", "discord-gg",
)

func NewService(config Config, logger *log.Logger) *Service {
	if logger == nil {
		logger = log.New(io.Discard)
	}
	return &Service{
		config: config,
		logger: logger,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		forumBaseURL: defaultForumBaseURL,
		ipAPIBaseURL: defaultIPAPIBaseURL,
		profileCache: make(map[string]string),
	}
}

func (s *Service) Listen(ctx context.Context) error {
	listener, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4zero, Port: s.config.UDPPort})
	if err != nil {
		return fmt.Errorf("listen udp: %w", err)
	}
	defer listener.Close()

	buffer := make([]byte, maxUDPPacketSize)
	for {
		if err := listener.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			return fmt.Errorf("set udp read deadline: %w", err)
		}

		readBytes, _, err := listener.ReadFromUDP(buffer)
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				continue
			}
			return fmt.Errorf("udp read: %w", err)
		}

		if err := s.HandlePacket(ctx, buffer[:readBytes]); err != nil {
			s.logger.Warn("packet handling failed", "error", err)
		}
	}
}

func (s *Service) SendStartupHello(ctx context.Context) error {
	return s.sendWebhook(ctx, webhookPayload{
		Username:  cutServerName("BeamMP ChatHook"),
		AvatarURL: s.config.AvatarURL,
		Content:   fmt.Sprintf("### 🌺 Hello from [*BeamMP ChatHook*](https://github.com/OfficialLambdax/BeamMP-ChatHook) v%d o/", Version),
	})
}

func (s *Service) HandlePacket(ctx context.Context, raw []byte) error {
	message, err := decodePacket(raw)
	if err != nil {
		return err
	}

	var scriptBuffer strings.Builder
	for _, item := range message.Contents {
		if err := s.handleContent(ctx, message, item, &scriptBuffer); err != nil {
			s.logger.Warn("handling content failed", "type", item.Type, "server", message.ServerName, "error", err)
		}
	}

	if scriptBuffer.Len() > 0 {
		content := strings.TrimSuffix(scriptBuffer.String(), "\n")
		return s.sendServerContent(ctx, message.ServerName, content)
	}

	return nil
}

func decodePacket(raw []byte) (packet, error) {
	decodedBase64, err := base64.StdEncoding.DecodeString(string(raw))
	if err != nil {
		return packet{}, fmt.Errorf("decode base64 packet: %w", err)
	}

	var wire packetWire
	if err := json.Unmarshal(decodedBase64, &wire); err != nil {
		return packet{}, fmt.Errorf("decode json packet: %w", err)
	}

	if wire.ServerName == nil || wire.PlayerCount == nil || wire.PlayerMax == nil || wire.PlayerDif == nil ||
		wire.Version == nil || wire.Contents == nil {
		return packet{}, fmt.Errorf("invalid message pack format: %s", string(decodedBase64))
	}

	message := packet{
		ServerName:  *wire.ServerName,
		PlayerCount: *wire.PlayerCount,
		PlayerMax:   *wire.PlayerMax,
		PlayerDif:   *wire.PlayerDif,
		Version:     *wire.Version,
		Contents:    wire.Contents,
	}

	if message.Version != ProtocolVersion {
		return packet{}, fmt.Errorf("invalid message pack version: %s", string(decodedBase64))
	}
	message.ServerName = cleanseString(message.ServerName)

	return message, nil
}

func (s *Service) handleContent(ctx context.Context, message packet, item content, scriptBuffer *strings.Builder) error {
	switch item.Type {
	case 1:
		chat, err := decodeChatMessage(item)
		if err != nil {
			return err
		}
		return s.sendChatMessage(ctx, message, chat)
	case 2:
		return s.sendServerOnline(ctx, message)
	case 3:
		player, err := decodePlayerJoin(item)
		if err != nil {
			return err
		}
		return s.sendPlayerJoin(ctx, message, player)
	case 4:
		player, err := decodePlayerLeft(item)
		if err != nil {
			return err
		}
		return s.sendPlayerLeft(ctx, message, player)
	case 5:
		return s.sendServerReload(ctx, message)
	case 6:
		script, err := decodeScriptMessage(item)
		if err != nil {
			return err
		}
		appendScriptMessage(scriptBuffer, script)
		return nil
	case 7:
		player, err := decodePlayerJoining(item)
		if err != nil {
			return err
		}
		return s.sendPlayerJoining(ctx, message, player)
	case 8:
		script, err := decodeScriptMessage(item)
		if err != nil {
			return err
		}
		return s.sendScriptMessageNoBuf(ctx, message, script)
	default:
		return fmt.Errorf("invalid option from %s", message.ServerName)
	}
}

func decodeChatMessage(item content) (chatMessage, error) {
	var message struct {
		PlayerName  *string `json:"player_name"`
		ChatMessage *string `json:"chat_message"`
	}
	if err := json.Unmarshal(item.Content, &message); err != nil {
		return chatMessage{}, err
	}
	if message.PlayerName == nil || message.ChatMessage == nil {
		return chatMessage{}, errors.New("invalid format for chat message")
	}
	return chatMessage{
		PlayerName:  *message.PlayerName,
		ChatMessage: chatSanitizer.Replace(*message.ChatMessage),
	}, nil
}

func decodeScriptMessage(item content) (scriptMessage, error) {
	var message struct {
		ScriptRef   *string `json:"script_ref"`
		ChatMessage *string `json:"chat_message"`
	}
	if err := json.Unmarshal(item.Content, &message); err != nil {
		return scriptMessage{}, err
	}
	if message.ScriptRef == nil || message.ChatMessage == nil {
		return scriptMessage{}, errors.New("invalid format for script message")
	}

	return scriptMessage{
		ScriptRef:   strings.ReplaceAll(*message.ScriptRef, "@", ""),
		ChatMessage: strings.ReplaceAll(*message.ChatMessage, "@", ""),
	}, nil
}

func decodePlayerJoining(item content) (playerJoining, error) {
	var message playerJoining
	if err := json.Unmarshal(item.Content, &message); err != nil {
		return playerJoining{}, err
	}
	if message.PlayerName == "" {
		return playerJoining{}, errors.New("invalid format for player joining")
	}
	return message, nil
}

func decodePlayerJoin(item content) (playerJoin, error) {
	var message playerJoin
	if err := json.Unmarshal(item.Content, &message); err != nil {
		return playerJoin{}, err
	}
	if message.PlayerName == "" || message.IP == "" {
		return playerJoin{}, errors.New("invalid format for player join")
	}
	return message, nil
}

func decodePlayerLeft(item content) (playerLeft, error) {
	var message struct {
		PlayerName *string `json:"player_name"`
		Early      *bool   `json:"early"`
	}
	if err := json.Unmarshal(item.Content, &message); err != nil {
		return playerLeft{}, err
	}
	if message.PlayerName == nil || message.Early == nil {
		return playerLeft{}, errors.New("invalid format for player left")
	}
	if *message.PlayerName == "" {
		return playerLeft{}, errors.New("invalid format for player left")
	}
	return playerLeft{PlayerName: *message.PlayerName, Early: *message.Early}, nil
}

func (s *Service) sendPlayerJoining(ctx context.Context, message packet, player playerJoining) error {
	content := fmt.Sprintf("> -# - 🕵️ ***%s** joining (%d Qued)*",
		player.PlayerName,
		message.PlayerDif,
	)
	return s.sendServerContent(ctx, message.ServerName, content)
}

func (s *Service) sendPlayerJoin(ctx context.Context, message packet, player playerJoin) error {
	profileColor := profileColorFromName(player.PlayerName)
	profilePic := s.profilePicture(player.PlayerName)
	countryFlag, isVPN := s.ipAPIResult(player.IP)

	vpnLabel := ""
	if isVPN {
		vpnLabel = " (VPN)"
	}

	var content string
	if isGuest(player.PlayerName) {
		content = fmt.Sprintf(
			"→ %s %s%s\n\n-# %d/%d Players ♦ %d Qued",
			player.PlayerName,
			countryFlag,
			vpnLabel,
			message.PlayerCount,
			message.PlayerMax,
			message.PlayerDif,
		)
	} else {
		content = fmt.Sprintf(
			"→ [%s](https://forum.beammp.com/u/%s) %s%s\n\n-# %d/%d Players ♦ %d Qued",
			player.PlayerName,
			player.PlayerName,
			countryFlag,
			vpnLabel,
			message.PlayerCount,
			message.PlayerMax,
			message.PlayerDif,
		)
	}

	return s.sendWebhook(ctx, webhookPayload{
		Username:  cutServerName(message.ServerName),
		AvatarURL: s.config.AvatarURL,
		Embeds: []webhookEmbed{
			{
				Thumbnail: &webhookThumbnail{URL: profilePic},
				Color:     profileColor,
				Fields: []webhookField{
					{
						Name:  "🧡 New Player Joined!",
						Value: content,
					},
				},
			},
		},
	})
}

func (s *Service) sendPlayerLeft(ctx context.Context, message packet, player playerLeft) error {
	var content string
	if !player.Early {
		content = fmt.Sprintf("> - 🚪 ***%s** left (%d/%d)*", player.PlayerName, message.PlayerCount, message.PlayerMax)
	} else {
		content = fmt.Sprintf("> -# - 🚪 ***%s** left during download (%d/%d)*", player.PlayerName, message.PlayerCount, message.PlayerMax)
	}
	return s.sendServerContent(ctx, message.ServerName, content)
}

func (s *Service) sendChatMessage(ctx context.Context, message packet, chat chatMessage) error {
	content := fmt.Sprintf("> - 💬 **%s:** %s", chat.PlayerName, chat.ChatMessage)
	return s.sendServerContent(ctx, message.ServerName, content)
}

func appendScriptMessage(buffer *strings.Builder, chat scriptMessage) {
	if chat.ScriptRef == "" {
		buffer.WriteString(fmt.Sprintf("%s\n", cleanseString(chat.ChatMessage)))
		return
	}
	buffer.WriteString(fmt.Sprintf("> - ⚙️ **%s:** %s\n", chat.ScriptRef, cleanseString(chat.ChatMessage)))
}

func (s *Service) sendScriptMessageNoBuf(ctx context.Context, message packet, chat scriptMessage) error {
	var content string
	if chat.ScriptRef == "" {
		content = fmt.Sprintf("%s\n", cleanseString(chat.ChatMessage))
	} else {
		content = fmt.Sprintf("> - ⚙️ **%s:** %s\n", chat.ScriptRef, cleanseString(chat.ChatMessage))
	}
	return s.sendServerContent(ctx, message.ServerName, content)
}

func (s *Service) sendServerOnline(ctx context.Context, message packet) error {
	return s.sendServerContent(ctx, message.ServerName, "## ✅ Server has just (re)started!")
}

func (s *Service) sendServerReload(ctx context.Context, message packet) error {
	return s.sendServerContent(ctx, message.ServerName, "## ♻️ Server side script has reloaded")
}

func (s *Service) sendServerContent(ctx context.Context, serverName, content string) error {
	return s.sendWebhook(ctx, webhookPayload{
		Username:  cutServerName(serverName),
		AvatarURL: s.config.AvatarURL,
		Content:   content,
	})
}

func (s *Service) sendWebhook(ctx context.Context, payload webhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send webhook: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook request failed: %s", strings.TrimSpace(string(respBody)))
	}

	return nil
}

func (s *Service) profilePicture(playerName string) string {
	if isGuest(playerName) {
		return ""
	}

	s.mu.Lock()
	if cached, ok := s.profileCache[playerName]; ok {
		s.mu.Unlock()
		return cached
	}
	s.mu.Unlock()

	baseURL, err := url.Parse(s.forumBaseURL)
	if err != nil {
		return ""
	}
	baseURL.Path = path.Join(baseURL.Path, "u", playerName+".json")

	resp, err := s.httpClient.Get(baseURL.String())
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var decoded forumResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ""
	}

	if decoded.User.AvatarTemplate == "" {
		return ""
	}

	url := s.forumBaseURL + decoded.User.AvatarTemplate
	url = strings.ReplaceAll(url, "{size}", "144")

	s.mu.Lock()
	s.profileCache[playerName] = url
	s.mu.Unlock()

	return url
}

func (s *Service) ipAPIResult(ip string) (string, bool) {
	baseURL, err := url.Parse(s.ipAPIBaseURL)
	if err != nil {
		return "", false
	}
	baseURL.Path = path.Join(baseURL.Path, "json", ip)
	query := baseURL.Query()
	query.Set("fields", "status,countryCode,proxy,hosting")
	baseURL.RawQuery = query.Encode()

	resp, err := s.httpClient.Get(baseURL.String())
	if err != nil {
		return "", false
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", false
	}

	var decoded ipAPIResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return "", false
	}
	if decoded.Status != "success" {
		return "", false
	}

	flag := countryFlagEmoji(decoded.CountryCode)
	return flag, decoded.Hosting || decoded.Proxy
}

func countryFlagEmoji(code string) string {
	if code == "" {
		return ""
	}
	country := countries.ByName(strings.ToUpper(code))
	if country == countries.Unknown {
		return ""
	}
	return country.Emoji()
}

func profileColorFromName(name string) uint32 {
	var color uint32
	for _, char := range []byte(name) {
		value := uint32(char) * 10000
		if color+value >= 16777215 {
			break
		}
		color += value
	}
	return color
}

func isGuest(playerName string) bool {
	if len(playerName) != 12 {
		return false
	}
	if !strings.HasPrefix(playerName, "guest") {
		return false
	}
	for _, char := range playerName[5:] {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}

func cutServerName(serverName string) string {
	if utf8.RuneCountInString(serverName) <= maxServerNameRunes {
		return serverName
	}

	var builder strings.Builder
	count := 0
	for _, char := range serverName {
		if count >= maxServerNameRunes-3 {
			break
		}
		builder.WriteRune(char)
		count++
	}
	builder.WriteString("...")
	return builder.String()
}

func cleanseString(input string) string {
	value := input
	for {
		pos := strings.Index(value, "^")
		if pos == -1 {
			return value
		}
		prefix := value[:pos]
		suffix := ""
		if pos+2 < len(value) {
			suffix = value[pos+2:]
		}
		value = prefix + suffix
	}
}
