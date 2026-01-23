package system

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

var (
	ErrUserExists    = errors.New("user already exists")
	ErrUserNotFound  = errors.New("user not found")
	ErrBadCredential = errors.New("invalid credentials")
)

type userStore struct {
	Users map[string]string `json:"users"` // username -> bcrypt hash
}

func usersFilePath() (string, error) {
	base, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "users.json"), nil
}

func loadUsers() (*userStore, error) {
	path, err := usersFilePath()
	if err != nil {
		return nil, err
	}

	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return &userStore{Users: map[string]string{}}, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s userStore
	if err := json.NewDecoder(f).Decode(&s); err != nil {
		return nil, err
	}

	if s.Users == nil {
		s.Users = map[string]string{}
	}

	return &s, nil
}

func saveUsers(s *userStore) error {
	path, err := usersFilePath()
	if err != nil {
		return err
	}

	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return err
	}

	if err := json.NewEncoder(f).Encode(s); err != nil {
		f.Close()
		return err
	}
	f.Close()

	return os.Rename(tmp, path)
}
