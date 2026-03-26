package app

import (
	"regexp"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/inrd/nocte/internal/config"
)

const largeNoteWarningThreshold int64 = 10 * 1024 * 1024

const (
	defaultDialogWidth     = 64
	defaultListDialogWidth = 84
	maxListDialogWidth     = 100
	listUpdatedAtWidth     = 17
	listMetaWidth          = 20
	listColumnGap          = 2
	editorPaneGap          = 2
	minEditorPaneWidth     = 24
	minPreviewPaneWidth    = 24
)

var (
	docStyle = lipgloss.NewStyle().
			Padding(1, 2)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	metaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10"))

	dialogStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("12")).
			Padding(1, 2).
			Width(defaultDialogWidth)

	dialogTitleStyle = lipgloss.NewStyle().
				Bold(true)

	commandPaletteStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1).
				Width(48)

	commandSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("0")).
				Background(lipgloss.Color("12"))

	listNameStyle = lipgloss.NewStyle()

	listUpdatedStyle = lipgloss.NewStyle()

	linkLabelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	linkURLStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	previewHeadingStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("12"))

	previewQuoteStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("11"))

	previewCodeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))

	previewMutedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	previewInlineCodeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("10"))

	previewLinkLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("12")).
				Underline(true)

	previewLinkURLStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8"))

	keyHintStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true)

	invalidFileChars      = regexp.MustCompile(`[^a-z0-9._-]+`)
	openPathWithSystemApp = openPath
	openURLWithSystemApp  = openURL
	commands              = []command{
		{name: ":help", description: "Show available commands"},
		{name: ":files", description: "Open the notes folder"},
		{name: ":info", description: "Show app and path info"},
		{name: ":list", description: "List existing notes"},
		{name: ":quit", description: "Exit the app"},
	}
)

type Model struct {
	input          textinput.Model
	editor         textarea.Model
	width          int
	height         int
	status         string
	isError        bool
	config         config.Config
	configPath     string
	version        string
	activeDialog   string
	commandIndex   int
	noteIndex      int
	noteMatches    []noteMatch
	dialogNotes    []noteMatch
	dialogLinks    []noteLink
	dialogIndex    int
	dialogOffset   int
	editorPath     string
	editorName     string
	lastSaved      string
	editorCreated  bool
	editorAction   string
	previewEnabled bool
}

type command struct {
	name        string
	description string
}

type noteMatch struct {
	name      string
	path      string
	score     int
	wordCount int
	sizeBytes int64
	modTime   time.Time
}

type noteLink struct {
	label string
	url   string
}

func New(cfg config.Config, configPath string, version string) Model {
	input := textinput.New()
	input.Placeholder = "Search or create a note..."
	input.Prompt = ""
	input.Focus()
	input.Width = 48

	editor := textarea.New()
	editor.Placeholder = "Start writing..."
	editor.Prompt = ""
	editor.ShowLineNumbers = false
	editor.CharLimit = 0
	editor.SetHeight(12)
	editor.SetWidth(64)

	return Model{
		input:          input,
		editor:         editor,
		noteIndex:      -1,
		dialogIndex:    -1,
		dialogOffset:   0,
		config:         cfg,
		configPath:     configPath,
		version:        version,
		previewEnabled: true,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
