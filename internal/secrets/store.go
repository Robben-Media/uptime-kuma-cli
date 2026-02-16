package secrets

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/99designs/keyring"
	"golang.org/x/term"

	"github.com/builtbyrobben/uptime-kuma-cli/internal/config"
)

type Store interface {
	GetAPIURL() (string, error)
	SetAPIURL(url string) error
	GetUsername() (string, error)
	SetUsername(username string) error
	GetPassword() (string, error)
	SetPassword(password string) error
	DeleteAll() error
	HasCredentials() (bool, error)
}

type KeyringStore struct {
	ring keyring.Keyring
}

const (
	apiURLKey   = "api_url"
	usernameKey = "username"
	passwordKey = "password"

	keyringPasswordEnv = "UPTIME_KUMA_CLI_KEYRING_PASS" //nolint:gosec // env var name, not a credential
	keyringBackendEnv  = "UPTIME_KUMA_CLI_KEYRING_BACKEND"
	keyringOpenTimeout = 5 * time.Second
)

var (
	errMissingValue          = errors.New("value cannot be empty")
	errMissingSecretKey      = errors.New("missing secret key")
	errNoTTY                 = errors.New("no TTY available for keyring file backend password prompt")
	errInvalidKeyringBackend = errors.New("invalid keyring backend")
	errKeyringTimeout        = errors.New("keyring connection timed out")
)

type KeyringBackendInfo struct {
	Value  string
	Source string
}

const (
	keyringBackendSourceEnv     = "env"
	keyringBackendSourceDefault = "default"
	keyringBackendAuto          = "auto"
)

func ResolveKeyringBackendInfo() (KeyringBackendInfo, error) {
	if v := normalizeKeyringBackend(os.Getenv(keyringBackendEnv)); v != "" {
		return KeyringBackendInfo{Value: v, Source: keyringBackendSourceEnv}, nil
	}

	return KeyringBackendInfo{Value: keyringBackendAuto, Source: keyringBackendSourceDefault}, nil
}

func allowedBackends(info KeyringBackendInfo) ([]keyring.BackendType, error) {
	switch info.Value {
	case "", keyringBackendAuto:
		return nil, nil
	case "keychain":
		return []keyring.BackendType{keyring.KeychainBackend}, nil
	case "file":
		return []keyring.BackendType{keyring.FileBackend}, nil
	default:
		return nil, fmt.Errorf("%w: %q (expected %s, keychain, or file)", errInvalidKeyringBackend, info.Value, keyringBackendAuto)
	}
}

func fileKeyringPasswordFunc() keyring.PromptFunc {
	return fileKeyringPasswordFuncFrom(os.Getenv(keyringPasswordEnv), term.IsTerminal(int(os.Stdin.Fd())))
}

func fileKeyringPasswordFuncFrom(password string, isTTY bool) keyring.PromptFunc {
	if password != "" {
		return keyring.FixedStringPrompt(password)
	}

	if isTTY {
		return keyring.TerminalPrompt
	}

	return func(_ string) (string, error) {
		return "", fmt.Errorf("%w; set %s", errNoTTY, keyringPasswordEnv)
	}
}

func normalizeKeyringBackend(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func shouldForceFileBackend(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == keyringBackendAuto && dbusAddr == ""
}

func shouldUseKeyringTimeout(goos string, backendInfo KeyringBackendInfo, dbusAddr string) bool {
	return goos == "linux" && backendInfo.Value == "auto" && dbusAddr != ""
}

func openKeyring() (keyring.Keyring, error) {
	keyringDir, err := config.EnsureKeyringDir()
	if err != nil {
		return nil, fmt.Errorf("ensure keyring dir: %w", err)
	}

	backendInfo, err := ResolveKeyringBackendInfo()
	if err != nil {
		return nil, err
	}

	backends, err := allowedBackends(backendInfo)
	if err != nil {
		return nil, err
	}

	dbusAddr := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if shouldForceFileBackend(runtime.GOOS, backendInfo, dbusAddr) {
		backends = []keyring.BackendType{keyring.FileBackend}
	}

	cfg := keyring.Config{
		ServiceName:              config.AppName,
		KeychainTrustApplication: true,
		AllowedBackends:          backends,
		FileDir:                  keyringDir,
		FilePasswordFunc:         fileKeyringPasswordFunc(),
	}

	if shouldUseKeyringTimeout(runtime.GOOS, backendInfo, dbusAddr) {
		return openKeyringWithTimeout(cfg, keyringOpenTimeout)
	}

	ring, err := keyring.Open(cfg)
	if err != nil {
		return nil, fmt.Errorf("open keyring: %w", err)
	}

	return ring, nil
}

type keyringResult struct {
	ring keyring.Keyring
	err  error
}

func openKeyringWithTimeout(cfg keyring.Config, timeout time.Duration) (keyring.Keyring, error) {
	ch := make(chan keyringResult, 1)

	go func() {
		ring, err := keyring.Open(cfg)
		ch <- keyringResult{ring, err}
	}()

	select {
	case res := <-ch:
		if res.err != nil {
			return nil, fmt.Errorf("open keyring: %w", res.err)
		}

		return res.ring, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("%w after %v (D-Bus SecretService may be unresponsive); "+
			"set UPTIME_KUMA_CLI_KEYRING_BACKEND=file and UPTIME_KUMA_CLI_KEYRING_PASS=<password> to use encrypted file storage instead",
			errKeyringTimeout, timeout)
	}
}

func OpenDefault() (Store, error) {
	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}

	return &KeyringStore{ring: ring}, nil
}

func (s *KeyringStore) GetAPIURL() (string, error) {
	item, err := s.ring.Get(apiURLKey)
	if err != nil {
		return "", fmt.Errorf("read API URL: %w", err)
	}

	return string(item.Data), nil
}

func (s *KeyringStore) SetAPIURL(u string) error {
	u = strings.TrimSpace(u)
	if u == "" {
		return fmt.Errorf("API URL: %w", errMissingValue)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  apiURLKey,
		Data: []byte(u),
	}); err != nil {
		return fmt.Errorf("store API URL: %w", err)
	}

	return nil
}

func (s *KeyringStore) GetUsername() (string, error) {
	item, err := s.ring.Get(usernameKey)
	if err != nil {
		return "", fmt.Errorf("read username: %w", err)
	}

	return string(item.Data), nil
}

func (s *KeyringStore) SetUsername(username string) error {
	username = strings.TrimSpace(username)
	if username == "" {
		return fmt.Errorf("username: %w", errMissingValue)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  usernameKey,
		Data: []byte(username),
	}); err != nil {
		return fmt.Errorf("store username: %w", err)
	}

	return nil
}

func (s *KeyringStore) GetPassword() (string, error) {
	item, err := s.ring.Get(passwordKey)
	if err != nil {
		return "", fmt.Errorf("read password: %w", err)
	}

	return string(item.Data), nil
}

func (s *KeyringStore) SetPassword(pw string) error {
	pw = strings.TrimSpace(pw)
	if pw == "" {
		return fmt.Errorf("password: %w", errMissingValue)
	}

	if err := s.ring.Set(keyring.Item{
		Key:  passwordKey,
		Data: []byte(pw),
	}); err != nil {
		return fmt.Errorf("store password: %w", err)
	}

	return nil
}

func (s *KeyringStore) DeleteAll() error {
	for _, key := range []string{apiURLKey, usernameKey, passwordKey} {
		if err := s.ring.Remove(key); err != nil && !errors.Is(err, keyring.ErrKeyNotFound) {
			return fmt.Errorf("delete %s: %w", key, err)
		}
	}

	return nil
}

func (s *KeyringStore) HasCredentials() (bool, error) {
	for _, key := range []string{apiURLKey, usernameKey, passwordKey} {
		_, err := s.ring.Get(key)
		if err != nil {
			if errors.Is(err, keyring.ErrKeyNotFound) {
				return false, nil
			}

			return false, fmt.Errorf("check %s: %w", key, err)
		}
	}

	return true, nil
}

func GetSecret(key string) ([]byte, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, errMissingSecretKey
	}

	ring, err := openKeyring()
	if err != nil {
		return nil, err
	}

	item, err := ring.Get(key)
	if err != nil {
		return nil, fmt.Errorf("read secret: %w", err)
	}

	return item.Data, nil
}

func SetSecret(key string, value []byte) error {
	key = strings.TrimSpace(key)
	if key == "" {
		return errMissingSecretKey
	}

	ring, err := openKeyring()
	if err != nil {
		return err
	}

	if err := ring.Set(keyring.Item{
		Key:  key,
		Data: value,
	}); err != nil {
		return fmt.Errorf("store secret: %w", err)
	}

	return nil
}
