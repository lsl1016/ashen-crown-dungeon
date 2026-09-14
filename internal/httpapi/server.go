package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"ashen-crown-dungeon/internal/game"
)

type Server struct {
	engine *game.Engine
	store  *game.Store
	mux    *http.ServeMux
	webDir string
}

func New(engine *game.Engine, store *game.Store, webDir string) *Server {
	s := &Server{engine: engine, store: store, mux: http.NewServeMux(), webDir: webDir}
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return logging(s.mux) }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, 200, map[string]any{"ok": true, "name": "ashen-crown-dungeon"})
	})
	s.mux.HandleFunc("GET /api/world", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, game.World()) })
	s.mux.HandleFunc("GET /api/tools", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, game.ToolDefinitions()) })
	s.mux.HandleFunc("GET /api/saves", s.listSaves)
	s.mux.HandleFunc("POST /api/saves/{id}/load", s.loadSave)
	s.mux.HandleFunc("POST /api/runs", s.createRun)
	s.mux.HandleFunc("GET /api/runs/{id}", s.getRun)
	s.mux.HandleFunc("POST /api/runs/{id}/move", s.move)
	s.mux.HandleFunc("POST /api/runs/{id}/action", s.action)
	s.mux.HandleFunc("POST /api/runs/{id}/combat", s.combat)
	s.mux.HandleFunc("POST /api/runs/{id}/item", s.item)
	s.mux.HandleFunc("POST /api/runs/{id}/talent", s.talent)
	s.mux.HandleFunc("POST /api/runs/{id}/interact", s.interact)
	s.mux.HandleFunc("POST /api/runs/{id}/save", s.manualSave)
	s.mux.HandleFunc("POST /api/runs/{id}/tools/execute", s.executeTool)
	s.mux.Handle("GET /", spaHandler(s.webDir))
}

func (s *Server) createRun(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Name  string `json:"name"`
		Class string `json:"class"`
		Seed  int64  `json:"seed"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errOut(w, 400, "请求格式错误")
		return
	}
	run, err := s.engine.NewRun(in.Name, in.Class, in.Seed)
	if err != nil {
		errOut(w, 400, err.Error())
		return
	}
	_ = s.store.SaveRun(run)
	jsonOut(w, 201, s.snapshot(run))
}
func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.LoadRun(r.PathValue("id"))
	if err != nil {
		errOut(w, 404, "冒险不存在")
		return
	}
	jsonOut(w, 200, s.snapshot(run))
}
func (s *Server) move(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			RoomID string `json:"roomId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.Move(run, in.RoomID)
	})
}
func (s *Server) action(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			ChoiceID string `json:"choiceId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.ResolveChoice(run, in.ChoiceID)
	})
}
func (s *Server) combat(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.CombatAction(run, in.Action)
	})
}
func (s *Server) item(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			ItemID string `json:"itemId"`
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.ItemAction(run, in.ItemID, in.Action)
	})
}

func (s *Server) talent(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			TalentID string `json:"talentId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.UnlockTalent(run, in.TalentID)
	})
}

func (s *Server) interact(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			ElementID string `json:"elementId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.Interact(run, in.ElementID)
	})
}
func (s *Server) manualSave(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.LoadRun(r.PathValue("id"))
	if err != nil {
		errOut(w, 404, "冒险不存在")
		return
	}
	var in struct {
		Name string `json:"name"`
	}
	_ = json.NewDecoder(r.Body).Decode(&in)
	meta, err := s.store.ManualSave(run, in.Name)
	if err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 201, meta)
}
func (s *Server) listSaves(w http.ResponseWriter, r *http.Request) {
	saves, err := s.store.ListSaves()
	if err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, saves)
}
func (s *Server) loadSave(w http.ResponseWriter, r *http.Request) {
	run, err := s.store.LoadSave(r.PathValue("id"))
	if err != nil {
		errOut(w, 404, "存档不存在")
		return
	}
	run.ID = fmt.Sprintf("run_loaded_%d", time.Now().UnixNano())
	_ = s.store.SaveRun(run)
	jsonOut(w, 200, s.snapshot(run))
}
func (s *Server) executeTool(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Name string         `json:"name"`
			Args map[string]any `json:"args"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.ExecuteTool(run, in.Name, in.Args)
	})
}

func (s *Server) mutate(w http.ResponseWriter, r *http.Request, fn func(*game.Run) error) {
	run, err := s.store.LoadRun(r.PathValue("id"))
	if err != nil {
		errOut(w, 404, "冒险不存在")
		return
	}
	if err := fn(run); err != nil {
		errOut(w, 400, err.Error())
		return
	}
	if err := s.store.SaveRun(run); err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, s.snapshot(run))
}
func (s *Server) snapshot(run *game.Run) map[string]any {
	return map[string]any{"run": run, "rooms": game.SortedRooms(run), "items": game.Items, "enemies": game.Enemies}
}

func jsonOut(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
func errOut(w http.ResponseWriter, status int, msg string) {
	jsonOut(w, status, map[string]any{"error": msg})
}
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func spaHandler(root string) http.Handler {
	fs := http.FileServer(http.Dir(root))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clean := path.Clean(r.URL.Path)
		if strings.Contains(clean, "..") {
			http.NotFound(w, r)
			return
		}
		if clean == "/" {
			http.ServeFile(w, r, path.Join(root, "index.html"))
			return
		}
		filePath := path.Join(root, clean)
		if _, err := http.Dir(root).Open(strings.TrimPrefix(clean, "/")); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
		if strings.Contains(path.Base(clean), ".") {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, path.Join(root, "index.html"))
		_ = filePath
	})
}

func ParsePort(v string) int {
	p, err := strconv.Atoi(v)
	if err != nil || p < 1 || p > 65535 {
		return 8080
	}
	return p
}
