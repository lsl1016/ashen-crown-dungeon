package game

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type SaveMeta struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	RunID       string    `json:"runId"`
	PlayerName  string    `json:"playerName"`
	Level       int       `json:"level"`
	CurrentRoom string    `json:"currentRoom"`
	CreatedAt   time.Time `json:"createdAt"`
}
type SaveFile struct {
	Meta SaveMeta `json:"meta"`
	Run  *Run     `json:"run"`
}

type Store struct {
	root string
	mu   sync.RWMutex
}

func NewStore(root string) *Store {
	_ = os.MkdirAll(filepath.Join(root, "runs"), 0755)
	_ = os.MkdirAll(filepath.Join(root, "saves"), 0755)
	return &Store{root: root}
}

func (s *Store) SaveRun(r *Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.UpdatedAt = time.Now()
	return writeJSON(filepath.Join(s.root, "runs", r.ID+".json"), r)
}
func (s *Store) LoadRun(id string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var r Run
	if err := readJSON(filepath.Join(s.root, "runs", id+".json"), &r); err != nil {
		return nil, err
	}
	return &r, nil
}
func (s *Store) ManualSave(r *Run, name string) (SaveMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if name == "" {
		name = "手动存档"
	}
	id := "save_" + time.Now().Format("20060102_150405.000000000")
	room := r.Rooms[r.CurrentRoomID]
	roomName := r.CurrentRoomID
	if room != nil {
		roomName = room.Name
	}
	meta := SaveMeta{ID: id, Name: name, RunID: r.ID, PlayerName: r.Player.Name, Level: r.Player.Level, CurrentRoom: roomName, CreatedAt: time.Now()}
	sf := SaveFile{Meta: meta, Run: r}
	return meta, writeJSON(filepath.Join(s.root, "saves", id+".json"), sf)
}
func (s *Store) ListSaves() ([]SaveMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	entries, err := os.ReadDir(filepath.Join(s.root, "saves"))
	if err != nil {
		return nil, err
	}
	out := []SaveMeta{}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		var sf SaveFile
		if readJSON(filepath.Join(s.root, "saves", e.Name()), &sf) == nil {
			out = append(out, sf.Meta)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (s *Store) LoadSave(id string) (*Run, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var sf SaveFile
	if err := readJSON(filepath.Join(s.root, "saves", id+".json"), &sf); err != nil {
		return nil, err
	}
	if sf.Run == nil {
		return nil, errors.New("存档损坏")
	}
	return sf.Run, nil
}
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
func readJSON(path string, v any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}
