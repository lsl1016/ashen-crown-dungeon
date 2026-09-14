package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"

	"ashen-crown-dungeon/internal/agentgateway"
	"ashen-crown-dungeon/internal/game"
)

type Server struct {
	engine *game.Engine
	store  *game.Store
	agent  *agentgateway.Gateway
	mux    *http.ServeMux
	webDir string
}

func New(engine *game.Engine, store *game.Store, webDir string) *Server {
	s := &Server{engine: engine, store: store, mux: http.NewServeMux(), webDir: webDir}
	s.agent = newAgentGateway(engine, store)
	s.routes()
	return s
}
func (s *Server) Handler() http.Handler { return logging(s.mux) }

func (s *Server) routes() {
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		jsonOut(w, 200, map[string]any{"ok": true, "name": "ashen-crown-dungeon", "version": "0.7.0", "agentGatewayVersion": "1.0", "mcpProtocolVersion": mcpProtocolVersion})
	})
	s.mux.HandleFunc("GET /api/world", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, game.World()) })
	s.mux.HandleFunc("GET /api/tools", func(w http.ResponseWriter, r *http.Request) { jsonOut(w, 200, game.ToolDefinitions()) })
	s.mux.HandleFunc("GET /api/agent/tools", s.agentTools)
	s.mux.HandleFunc("GET /api/agent/tools/{name}", s.agentToolDefinition)
	s.mux.HandleFunc("POST /api/agent/tools/execute", s.agentExecute)
	s.mux.HandleFunc("GET /api/agent/audit", s.agentAudit)
	s.mux.HandleFunc("POST /mcp", s.mcp)
	s.mux.HandleFunc("GET /api/editor/content", s.editorContent)
	s.mux.HandleFunc("GET /api/editor/overrides", s.editorOverrides)
	s.mux.HandleFunc("GET /api/editor/runs", s.editorRuns)
	s.mux.HandleFunc("POST /api/editor/content", s.editorSaveContent)
	s.mux.HandleFunc("POST /api/editor/runs/{id}/world", s.editorWorld)
	s.mux.HandleFunc("POST /api/editor/runs/{id}/room", s.editorRoom)
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
	s.mux.HandleFunc("POST /api/runs/{id}/dialogue", s.dialogue)
	s.mux.HandleFunc("POST /api/runs/{id}/npc", s.npc)
	s.mux.HandleFunc("POST /api/runs/{id}/shop", s.shop)
	s.mux.HandleFunc("POST /api/runs/{id}/growth", s.growth)
	s.mux.HandleFunc("POST /api/runs/{id}/save", s.manualSave)
	s.mux.HandleFunc("POST /api/runs/{id}/tools/execute", s.executeTool)
	s.mux.Handle("GET /", spaHandler(s.webDir))
}

func (s *Server) editorContent(w http.ResponseWriter, r *http.Request) {
	jsonOut(w, 200, map[string]any{"content": game.EditorContent(), "kinds": game.SortedEditorKinds(), "version": "0.7.0"})
}

func (s *Server) editorRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := s.store.ListRuns()
	if err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, runs)
}

func (s *Server) editorOverrides(w http.ResponseWriter, r *http.Request) {
	items, err := s.store.LoadContentOverrides()
	if err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, items)
}

func (s *Server) editorSaveContent(w http.ResponseWriter, r *http.Request) {
	var in game.ContentOverride
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		errOut(w, 400, "请求格式错误")
		return
	}
	if err := game.ApplyContentOverride(in); err != nil {
		errOut(w, 400, err.Error())
		return
	}
	if err := s.store.UpsertContentOverride(in); err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, map[string]any{"ok": true, "kind": in.Kind, "id": in.ID})
}

func (s *Server) editorWorld(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Action    string `json:"action"`
			NPCID     string `json:"npcId"`
			RoomID    string `json:"roomId"`
			Region    string `json:"region"`
			Value     string `json:"value"`
			Flag      string `json:"flag"`
			BoolValue bool   `json:"boolValue"`
			EventID   string `json:"eventId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		switch in.Action {
		case "move_npc":
			return s.engine.GMMoveNPC(run, in.NPCID, in.RoomID)
		case "set_region":
			return s.engine.GMSetRegion(run, in.Region, in.Value)
		case "set_flag":
			return s.engine.GMSetFlag(run, in.Flag, in.BoolValue)
		case "trigger_event":
			return s.engine.GMTriggerWorldEvent(run, in.EventID)
		case "advance_bell":
			return s.engine.GMAdvanceBell(run)
		default:
			return fmt.Errorf("未知 GM 世界操作")
		}
	})
}

func (s *Server) editorRoom(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Room      game.Room `json:"room"`
			ConnectTo string    `json:"connectTo"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.GMUpsertRoom(run, in.Room, in.ConnectTo)
	})
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

func (s *Server) npc(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			NPCID string `json:"npcId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.StartDialogue(run, in.NPCID)
	})
}

func (s *Server) dialogue(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Action   string `json:"action"`
			ChoiceID string `json:"choiceId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		if in.Action == "close" {
			return s.engine.CloseDialogue(run)
		}
		return s.engine.DialogueChoice(run, in.ChoiceID)
	})
}

func (s *Server) shop(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			ShopID string `json:"shopId"`
			Action string `json:"action"`
			ItemID string `json:"itemId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		return s.engine.ShopAction(run, in.ShopID, in.Action, in.ItemID)
	})
}

func (s *Server) growth(w http.ResponseWriter, r *http.Request) {
	s.mutate(w, r, func(run *game.Run) error {
		var in struct {
			Kind string `json:"kind"`
			ID   string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			return err
		}
		switch in.Kind {
		case "attribute":
			return s.engine.UpgradeAttribute(run, in.ID)
		case "mastery":
			return s.engine.UpgradeGrowth(run, in.ID)
		default:
			return fmt.Errorf("未知成长类型")
		}
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
	s.engine.Prepare(run)
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
	s.engine.Prepare(run)
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
	s.engine.Prepare(run)
	if err := fn(run); err != nil {
		errOut(w, 400, err.Error())
		return
	}
	s.engine.Prepare(run)
	if err := s.store.SaveRun(run); err != nil {
		errOut(w, 500, err.Error())
		return
	}
	jsonOut(w, 200, s.snapshot(run))
}
func (s *Server) snapshot(run *game.Run) map[string]any {
	s.engine.Prepare(run)
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
		if clean == "/editor" {
			http.ServeFile(w, r, path.Join(root, "editor.html"))
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
