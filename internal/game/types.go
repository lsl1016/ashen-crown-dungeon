package game

import "time"

type AttributeSet struct {
	Strength   int `json:"strength"`
	Dexterity  int `json:"dexterity"`
	Perception int `json:"perception"`
	Will       int `json:"will"`
}

type Player struct {
	Name            string            `json:"name"`
	Class           string            `json:"class"`
	Level           int               `json:"level"`
	XP              int               `json:"xp"`
	HP              int               `json:"hp"`
	MaxHP           int               `json:"maxHp"`
	Energy          int               `json:"energy"`
	MaxEnergy       int               `json:"maxEnergy"`
	Defense         int               `json:"defense"`
	Gold            int               `json:"gold"`
	Attributes      AttributeSet      `json:"attributes"`
	Inventory       []string          `json:"inventory"`
	Statuses        []string          `json:"statuses"`
	Equipment       map[string]string `json:"equipment"`
	TalentPoints    int               `json:"talentPoints"`
	Talents         map[string]int    `json:"talents"`
	AttributePoints int               `json:"attributePoints"`
	MasteryPoints   int               `json:"masteryPoints"`
	Growth          map[string]int    `json:"growth"`
}

type SceneElement struct {
	ID               string    `json:"id"`
	Kind             string    `json:"kind"`
	Label            string    `json:"label"`
	Description      string    `json:"description"`
	Icon             string    `json:"icon"`
	X                int       `json:"x"`
	Y                int       `json:"y"`
	Action           string    `json:"action"`
	Target           string    `json:"target,omitempty"`
	Value            int       `json:"value,omitempty"`
	OneShot          bool      `json:"oneShot"`
	RequiresItem     string    `json:"requiresItem,omitempty"`
	RequiresFlag     string    `json:"requiresFlag,omitempty"`
	HiddenUnlessFlag string    `json:"hiddenUnlessFlag,omitempty"`
	ConsumesItem     bool      `json:"consumesItem,omitempty"`
	Check            *CheckDef `json:"check,omitempty"`
	OnSuccess        []Effect  `json:"onSuccess,omitempty"`
	OnFailure        []Effect  `json:"onFailure,omitempty"`
}

type Room struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	Scene       string         `json:"scene"`
	Description string         `json:"description"`
	Zone        string         `json:"zone"`
	X           int            `json:"x"`
	Y           int            `json:"y"`
	Discovered  bool           `json:"discovered"`
	Visited     bool           `json:"visited"`
	Resolved    bool           `json:"resolved"`
	Locked      bool           `json:"locked"`
	EventID     string         `json:"eventId,omitempty"`
	Elements    []SceneElement `json:"elements,omitempty"`
}

type Edge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type ItemDef struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Type        string       `json:"type"`
	Slot        string       `json:"slot,omitempty"`
	Rarity      string       `json:"rarity,omitempty"`
	Description string       `json:"description"`
	Power       int          `json:"power"`
	Value       int          `json:"value"`
	Icon        string       `json:"icon"`
	Defense     int          `json:"defense,omitempty"`
	MaxHP       int          `json:"maxHp,omitempty"`
	MaxEnergy   int          `json:"maxEnergy,omitempty"`
	Attributes  AttributeSet `json:"attributes,omitempty"`
	Special     string       `json:"special,omitempty"`
}

type EnemyDef struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	HP            int    `json:"hp"`
	Attack        int    `json:"attack"`
	Defense       int    `json:"defense"`
	DamageMin     int    `json:"damageMin"`
	DamageMax     int    `json:"damageMax"`
	XP            int    `json:"xp"`
	GoldMin       int    `json:"goldMin"`
	GoldMax       int    `json:"goldMax"`
	Icon          string `json:"icon"`
	Boss          bool   `json:"boss"`
	Archetype     string `json:"archetype,omitempty"`
	Weakness      string `json:"weakness,omitempty"`
	IntentProfile string `json:"intentProfile,omitempty"`
	Portrait      string `json:"portrait,omitempty"`
}

type StatusState struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Rounds      int    `json:"rounds"`
	Stacks      int    `json:"stacks"`
}

type EnemyIntent struct {
	ID          string `json:"id"`
	Kind        string `json:"kind"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Power       int    `json:"power"`
	Telegraph   string `json:"telegraph"`
}

type CombatState struct {
	EnemyID        string         `json:"enemyId"`
	EnemyName      string         `json:"enemyName"`
	EnemyHP        int            `json:"enemyHp"`
	EnemyMaxHP     int            `json:"enemyMaxHp"`
	EnemyDefense   int            `json:"enemyDefense"`
	Round          int            `json:"round"`
	Guarded        bool           `json:"guarded"`
	LastMessage    string         `json:"lastMessage"`
	BossPhase      int            `json:"bossPhase"`
	Intent         EnemyIntent    `json:"intent"`
	EnemyStatuses  []StatusState  `json:"enemyStatuses,omitempty"`
	PlayerStatuses []StatusState  `json:"playerStatuses,omitempty"`
	Cooldowns      map[string]int `json:"cooldowns,omitempty"`
	Counters       map[string]int `json:"counters,omitempty"`
}

type CheckDef struct {
	Attribute string `json:"attribute"`
	DC        int    `json:"dc"`
}

type Effect struct {
	Type   string `json:"type"`
	Value  int    `json:"value,omitempty"`
	Target string `json:"target,omitempty"`
	Text   string `json:"text,omitempty"`
}

type ChoiceDef struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Check     *CheckDef `json:"check,omitempty"`
	OnSuccess []Effect  `json:"onSuccess,omitempty"`
	OnFailure []Effect  `json:"onFailure,omitempty"`
}

type EventDef struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Choices     []ChoiceDef `json:"choices"`
}

type ActiveEvent struct {
	ID          string      `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Choices     []ChoiceDef `json:"choices"`
	LastResult  string      `json:"lastResult,omitempty"`
}

type QuestState struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	Goal        int    `json:"goal"`
}

type LoreEntry struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type RollResult struct {
	Kind     string `json:"kind"`
	Label    string `json:"label"`
	Roll     int    `json:"roll"`
	Bonus    int    `json:"bonus"`
	Total    int    `json:"total"`
	DC       int    `json:"dc,omitempty"`
	Success  bool   `json:"success"`
	Critical bool   `json:"critical,omitempty"`
}

type LogEntry struct {
	Turn      int       `json:"turn"`
	Type      string    `json:"type"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

type WorldClock struct {
	Bell       int    `json:"bell"`
	Steps      int    `json:"steps"`
	Threat     int    `json:"threat"`
	LastChange string `json:"lastChange,omitempty"`
}

type DialogueChoiceDef struct {
	ID           string   `json:"id"`
	Text         string   `json:"text"`
	NextNode     string   `json:"nextNode,omitempty"`
	End          bool     `json:"end,omitempty"`
	OpenShop     string   `json:"openShop,omitempty"`
	RequiresFlag string   `json:"requiresFlag,omitempty"`
	RequiresItem string   `json:"requiresItem,omitempty"`
	Effects      []Effect `json:"effects,omitempty"`
}

type DialogueNodeDef struct {
	ID      string              `json:"id"`
	Text    string              `json:"text"`
	Choices []DialogueChoiceDef `json:"choices"`
}

type DialogueDef struct {
	ID        string                     `json:"id"`
	NPCID     string                     `json:"npcId"`
	StartNode string                     `json:"startNode"`
	Nodes     map[string]DialogueNodeDef `json:"nodes"`
}

type ActiveDialogue struct {
	NPCID      string `json:"npcId"`
	DialogueID string `json:"dialogueId"`
	NodeID     string `json:"nodeId"`
}

type NPCDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Title       string `json:"title"`
	Faction     string `json:"faction"`
	Description string `json:"description"`
	Portrait    string `json:"portrait"`
	DialogueID  string `json:"dialogueId"`
	ShopID      string `json:"shopId,omitempty"`
}

type ShopItemDef struct {
	ItemID string `json:"itemId"`
	Price  int    `json:"price"`
	Stock  int    `json:"stock"`
}

type ShopDef struct {
	ID      string        `json:"id"`
	Name    string        `json:"name"`
	NPCID   string        `json:"npcId"`
	Buyback float64       `json:"buyback"`
	Items   []ShopItemDef `json:"items"`
}

type GrowthDef struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	MaxRank     int    `json:"maxRank"`
}

type Run struct {
	ID             string                 `json:"id"`
	Seed           int64                  `json:"seed"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
	Turn           int                    `json:"turn"`
	Player         Player                 `json:"player"`
	Rooms          map[string]*Room       `json:"rooms"`
	Edges          []Edge                 `json:"edges"`
	CurrentRoomID  string                 `json:"currentRoomId"`
	ActiveEvent    *ActiveEvent           `json:"activeEvent,omitempty"`
	ActiveDialogue *ActiveDialogue        `json:"activeDialogue,omitempty"`
	ActiveShop     string                 `json:"activeShop,omitempty"`
	Combat         *CombatState           `json:"combat,omitempty"`
	Flags          map[string]bool        `json:"flags"`
	Quests         map[string]*QuestState `json:"quests"`
	Lore           []LoreEntry            `json:"lore"`
	LastRoll       *RollResult            `json:"lastRoll,omitempty"`
	Log            []LogEntry             `json:"log"`
	Clock          WorldClock             `json:"clock"`
	NPCRelations   map[string]int         `json:"npcRelations,omitempty"`
	ShopStock      map[string]int         `json:"shopStock,omitempty"`
	GameOver       bool                   `json:"gameOver"`
	Victory        bool                   `json:"victory"`
}

type TalentDef struct {
	ID          string `json:"id"`
	Class       string `json:"class"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	MaxRank     int    `json:"maxRank"`
}

type WorldMeta struct {
	Title     string                 `json:"title"`
	Subtitle  string                 `json:"subtitle"`
	Premise   string                 `json:"premise"`
	Era       string                 `json:"era"`
	Factions  []string               `json:"factions"`
	Truths    []string               `json:"truths"`
	Classes   []ClassDef             `json:"classes"`
	Talents   []TalentDef            `json:"talents"`
	Items     map[string]ItemDef     `json:"items"`
	NPCs      map[string]NPCDef      `json:"npcs"`
	Shops     map[string]ShopDef     `json:"shops"`
	Dialogues map[string]DialogueDef `json:"dialogues"`
	Growth    []GrowthDef            `json:"growth"`
	ToolHint  string                 `json:"toolHint"`
}

type ClassDef struct {
	ID               string       `json:"id"`
	Name             string       `json:"name"`
	Description      string       `json:"description"`
	HP               int          `json:"hp"`
	Energy           int          `json:"energy"`
	Defense          int          `json:"defense"`
	Attributes       AttributeSet `json:"attributes"`
	StarterItem      string       `json:"starterItem"`
	SkillName        string       `json:"skillName"`
	SkillDescription string       `json:"skillDescription"`
	Icon             string       `json:"icon"`
}
