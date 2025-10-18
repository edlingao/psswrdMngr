package adapter

import (
	"fmt"
	"io"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/edlingao/psswrdMngr/internal/secured/ports"
	"github.com/sahilm/fuzzy"
)

type fieldViewState int

const (
	fieldListView fieldViewState = iota
	fieldAddFormView
	fieldEditFormView
	fieldDetailModalView
	fieldDeleteConfirmView
	fieldSuccessView
)

type fieldItem struct {
	field core.Field
}

func (i fieldItem) Title() string       { return i.field.Name }
func (i fieldItem) Description() string { return "••••••••" }
func (i fieldItem) FilterValue() string { return i.field.Name }

type fieldItemDelegate struct {
	width int
}

func (d fieldItemDelegate) Height() int                             { return 2 }
func (d fieldItemDelegate) Spacing() int                            { return 1 }
func (d fieldItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d fieldItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(fieldItem)
	if !ok {
		return
	}

	title := i.Title()
	desc := i.Description()

	if index == m.Index() {
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("39")).
			Width(d.width).
			Align(lipgloss.Center)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(d.width).
			Align(lipgloss.Center)

		title = titleStyle.Render(title)
		desc = descStyle.Render(desc)
	} else {
		titleStyle := lipgloss.NewStyle().
			Width(d.width).
			Align(lipgloss.Center)
		descStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Width(d.width).
			Align(lipgloss.Center)

		title = titleStyle.Render(title)
		desc = descStyle.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

type FieldsTUI struct {
	SecuredService ports.SecuredServiceOperations
	FieldsService  ports.FieldServiceOperations
	list           list.Model
	nameInput      textinput.Model
	valueInput     textinput.Model
	searchInput    textinput.Model
	state          fieldViewState
	width          int
	height         int
	message        string
	err            error
	secured        core.Secured
	fields         core.Fields
	allFields      core.Fields
	selectedField  core.Field
	valueRevealed  bool
	searchActive   bool
	searchFocused  bool
	parentTUI      tea.Model
}

func NewFieldsTUI(
	securedService ports.SecuredServiceOperations,
	fieldsService ports.FieldServiceOperations,
) *FieldsTUI {
	const defaultWidth = 80
	const listHeight = 14

	l := list.New([]list.Item{}, fieldItemDelegate{width: defaultWidth}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	nameInput := textinput.New()
	nameInput.Placeholder = "Field name (e.g., username, password)"
	nameInput.CharLimit = 256

	valueInput := textinput.New()
	valueInput.Placeholder = "Field value"
	valueInput.CharLimit = 1024

	searchInput := textinput.New()
	searchInput.Placeholder = "Search fields..."
	searchInput.CharLimit = 256

	return &FieldsTUI{
		SecuredService: securedService,
		FieldsService:  fieldsService,
		list:           l,
		nameInput:      nameInput,
		valueInput:     valueInput,
		searchInput:    searchInput,
		state:          fieldListView,
		width:          defaultWidth,
		height:         listHeight,
	}
}

func (m *FieldsTUI) SetSecured(secured core.Secured) {
	m.secured = secured
}

func (m *FieldsTUI) SetParentTUI(parent tea.Model) {
	m.parentTUI = parent
}

func (m *FieldsTUI) SetWindowSize(width, height int) {
	m.width = width
	m.height = height

	listWidth := width - 10
	listHeight := 10
	if height < 20 {
		listHeight = height - 10
	}

	m.list.SetSize(listWidth, listHeight)
	m.list.SetDelegate(fieldItemDelegate{width: listWidth})
	m.nameInput.Width = listWidth - 4
	m.valueInput.Width = listWidth - 4
	m.searchInput.Width = listWidth - 4
}

func (m *FieldsTUI) returnToParent() tea.Model {
	if m.parentTUI == nil {
		return m
	}

	if sizer, ok := m.parentTUI.(interface {
		Update(tea.Msg) (tea.Model, tea.Cmd)
	}); ok {
		updated, _ := sizer.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
		return updated
	}

	return m.parentTUI
}

func (m *FieldsTUI) loadFields() tea.Cmd {
	return func() tea.Msg {
		fields, err := m.FieldsService.GetFields(m.secured.ID)
		if err != nil {
			return fieldsLoadedMsg{err: err}
		}
		return fieldsLoadedMsg{fields: fields}
	}
}

type fieldsLoadedMsg struct {
	fields core.Fields
	err    error
}

func (m FieldsTUI) Init() tea.Cmd {
	return m.loadFields()
}

func (m FieldsTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case fieldsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.fields = msg.fields
		m.allFields = msg.fields
		items := make([]list.Item, len(msg.fields))
		for i, f := range msg.fields {
			items[i] = fieldItem{field: f}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := msg.Width - 10
		listHeight := 10
		if msg.Height < 20 {
			listHeight = msg.Height - 10
		}

		m.list.SetSize(listWidth, listHeight)
		m.list.SetDelegate(fieldItemDelegate{width: listWidth})
		m.nameInput.Width = listWidth - 4
		m.valueInput.Width = listWidth - 4
		m.searchInput.Width = listWidth - 4
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case fieldListView:
			if m.searchActive {
				if m.searchFocused {
					switch {
					case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
						m.searchFocused = false
						m.searchInput.Blur()
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
						m.searchActive = false
						m.searchFocused = false
						m.searchInput.Blur()
						m.searchInput.SetValue("")
						items := make([]list.Item, len(m.allFields))
						for i, f := range m.allFields {
							items[i] = fieldItem{field: f}
						}
						m.list.SetItems(items)
						m.fields = m.allFields
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
						if m.parentTUI != nil {
							return m.returnToParent(), nil
						}
						return m, tea.Quit
					default:
						var cmd tea.Cmd
						m.searchInput, cmd = m.searchInput.Update(msg)

					query := m.searchInput.Value()
					if query == "" {
						items := make([]list.Item, len(m.allFields))
						for i, f := range m.allFields {
							items[i] = fieldItem{field: f}
						}
						m.list.SetItems(items)
						m.fields = m.allFields
					} else {
						candidates := make([]string, len(m.allFields))
						for i, f := range m.allFields {
							candidates[i] = f.Name
						}

						matches := fuzzy.Find(query, candidates)
						items := make([]list.Item, len(matches))
						filtered := make(core.Fields, len(matches))

						for i, match := range matches {
							for _, f := range m.allFields {
								if f.Name == match.Str {
									items[i] = fieldItem{field: f}
									filtered[i] = f
									break
								}
							}
						}

						m.list.SetItems(items)
						m.fields = filtered
					}

					return m, cmd
					}
				} else {
					switch {
					case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
						m.searchFocused = true
						m.searchInput.Focus()
						return m, textinput.Blink
					case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
						m.searchActive = false
						m.searchFocused = false
						m.searchInput.Blur()
						m.searchInput.SetValue("")
						items := make([]list.Item, len(m.allFields))
						for i, f := range m.allFields {
							items[i] = fieldItem{field: f}
						}
						m.list.SetItems(items)
						m.fields = m.allFields
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
						if m.parentTUI != nil {
							return m.returnToParent(), nil
						}
						return m, tea.Quit
					case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
						if len(m.fields) == 0 {
							return m, nil
						}
						i, ok := m.list.SelectedItem().(fieldItem)
						if ok {
							m.selectedField = i.field
							m.valueRevealed = false
							m.state = fieldDetailModalView
							return m, nil
						}
					}
					var cmd tea.Cmd
					m.list, cmd = m.list.Update(msg)
					return m, cmd
				}
			}

			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
				m.searchActive = true
				m.searchFocused = true
				m.searchInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
				m.state = fieldAddFormView
				m.nameInput.Focus()
				m.valueInput.Blur()
				m.err = nil
				m.message = ""
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("e"))):
				if len(m.fields) == 0 {
					return m, nil
				}
				i, ok := m.list.SelectedItem().(fieldItem)
				if ok {
					m.selectedField = i.field
					m.valueInput.SetValue(i.field.Value)
					m.state = fieldEditFormView
					m.valueInput.Focus()
					m.err = nil
					m.message = ""
					return m, textinput.Blink
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				if len(m.fields) == 0 {
					return m, nil
				}
				i, ok := m.list.SelectedItem().(fieldItem)
				if ok {
					m.selectedField = i.field
					m.state = fieldDeleteConfirmView
					return m, nil
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if len(m.fields) == 0 {
					return m, nil
				}
				i, ok := m.list.SelectedItem().(fieldItem)
				if ok {
					m.selectedField = i.field
					m.valueRevealed = false
					m.state = fieldDetailModalView
					return m, nil
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case fieldAddFormView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				if m.nameInput.Focused() {
					m.state = fieldListView
					m.nameInput.Blur()
					m.valueInput.Blur()
					m.nameInput.SetValue("")
					m.valueInput.SetValue("")
					return m, nil
				}
				m.nameInput.Focus()
				m.valueInput.Blur()
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if m.nameInput.Focused() {
					name := m.nameInput.Value()
					if name == "" {
						m.err = fmt.Errorf("field name cannot be empty")
						return m, nil
					}
					m.nameInput.Blur()
					m.valueInput.Focus()
					m.err = nil
					return m, textinput.Blink
				}

				value := m.valueInput.Value()
				if value == "" {
					m.err = fmt.Errorf("field value cannot be empty")
					return m, nil
				}

				_, err := m.FieldsService.AddField(m.secured.ID, m.nameInput.Value(), value)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Field added successfully!"
				m.state = fieldSuccessView
				m.nameInput.SetValue("")
				m.valueInput.SetValue("")
				m.nameInput.Blur()
				m.valueInput.Blur()
				return m, m.loadFields()
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			if m.nameInput.Focused() {
				m.nameInput, cmd = m.nameInput.Update(msg)
			} else {
				m.valueInput, cmd = m.valueInput.Update(msg)
			}
			return m, cmd

		case fieldEditFormView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = fieldListView
				m.valueInput.Blur()
				m.valueInput.SetValue("")
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				value := m.valueInput.Value()
				if value == "" {
					m.err = fmt.Errorf("field value cannot be empty")
					return m, nil
				}

				_, err := m.FieldsService.UpdateField(m.secured.ID, m.selectedField.ID, value)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Field updated successfully!"
				m.state = fieldSuccessView
				m.valueInput.SetValue("")
				m.valueInput.Blur()
				return m, m.loadFields()
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.valueInput, cmd = m.valueInput.Update(msg)
			return m, cmd

		case fieldDetailModalView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = fieldListView
				m.valueRevealed = false
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("n"))):
				err := clipboard.WriteAll(m.selectedField.Name)
				if err != nil {
					m.err = err
				} else {
					m.message = "Name copied to clipboard!"
				}
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("v"))):
				err := clipboard.WriteAll(m.selectedField.Value)
				if err != nil {
					m.err = err
				} else {
					m.message = "Value copied to clipboard!"
				}
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			default:
				m.valueRevealed = !m.valueRevealed
				m.message = ""
				m.err = nil
				return m, nil
			}

		case fieldDeleteConfirmView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
				err := m.FieldsService.DeleteField(m.secured.ID, m.selectedField.ID)
				if err != nil {
					m.err = err
					m.state = fieldListView
					return m, nil
				}

				m.message = "Field deleted successfully!"
				m.state = fieldSuccessView
				return m, m.loadFields()
			case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc"))):
				m.state = fieldListView
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}

		case fieldSuccessView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc"))):
				m.state = fieldListView
				m.message = ""
				m.err = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m FieldsTUI) View() string {
	listWidth := m.width - 10

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("39")).
		Align(lipgloss.Center).
		Width(listWidth).
		MarginBottom(1)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2)

	inputStyle := lipgloss.NewStyle().
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(2)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	messageStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	switch m.state {
	case fieldListView:
		header := headerStyle.Render(fmt.Sprintf("Fields: %s", m.secured.Title))
		content := header

		if m.searchActive {
			searchBoxStyle := lipgloss.NewStyle().
				Width(listWidth).
				Align(lipgloss.Center).
				MarginTop(1)

			countStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Width(listWidth).
				Align(lipgloss.Center)

			content += "\n" + searchBoxStyle.Render(m.searchInput.View())

			totalCount := len(m.allFields)
			currentCount := len(m.fields)
			if m.searchInput.Value() == "" {
				currentCount = totalCount
			} else {
				currentCount = len(m.list.Items())
			}
			content += "\n" + countStyle.Render(fmt.Sprintf("%d of %d fields", currentCount, totalCount))
		}

		content += "\n" + m.list.View()

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		help := "a: add • e: edit • d: delete • /: search • enter: view • esc: back"
		if m.searchActive {
			if m.searchFocused {
				help = "tab: navigate list • esc: close search"
			} else {
				help = "tab: back to search • enter: select • esc: close search"
			}
		} else if len(m.fields) == 0 {
			help = "a: add new field • esc: back"
		}

		content += "\n" + helpStyle.Render(help)
		return boxStyle.Render(content)

	case fieldAddFormView:
		header := headerStyle.Render("Add Field")

		var content string
		if m.nameInput.Focused() {
			content = header + "\n" + inputStyle.Render(m.nameInput.View())
		} else {
			nameStyle := lipgloss.NewStyle().
				Width(listWidth).
				Align(lipgloss.Center).
				Foreground(lipgloss.Color("241"))
			content = header + "\n" + nameStyle.Render("Name: "+m.nameInput.Value())
			content += "\n" + inputStyle.Render(m.valueInput.View())
		}

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		if m.nameInput.Focused() {
			content += "\n" + helpStyle.Render("enter: continue • esc: cancel")
		} else {
			content += "\n" + helpStyle.Render("enter: save • esc: back")
		}
		return boxStyle.Render(content)

	case fieldEditFormView:
		header := headerStyle.Render("Edit Field")

		nameStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("241")).
			MarginTop(1)

		content := header + "\n" + nameStyle.Render("Name: "+m.selectedField.Name)
		content += "\n" + inputStyle.Render(m.valueInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: cancel")
		return boxStyle.Render(content)

	case fieldDetailModalView:
		header := headerStyle.Render("Field Details")

		fieldNameStyle := lipgloss.NewStyle().
			Bold(true).
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		fieldValueStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1).
			Foreground(lipgloss.Color("39"))

		content := header + "\n" + fieldNameStyle.Render(m.selectedField.Name)

		displayValue := m.selectedField.Value
		if !m.valueRevealed {
			displayValue = strings.Repeat("•", len(m.selectedField.Value))
		}
		content += "\n" + fieldValueStyle.Render(displayValue)

		if m.message != "" {
			content += "\n" + messageStyle.Render(m.message)
		}

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		help := "any key: reveal/hide • n: copy name • v: copy value • esc: close"
		content += "\n" + helpStyle.Render(help)
		return boxStyle.Render(content)

	case fieldDeleteConfirmView:
		header := headerStyle.Render("Delete Field")

		confirmStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		content := header + "\n" + confirmStyle.Render(fmt.Sprintf("Delete field '%s'?", m.selectedField.Name))
		content += "\n" + helpStyle.Render("y: yes • n: no")
		return boxStyle.Render(content)

	case fieldSuccessView:
		header := headerStyle.Render("Success")

		content := header + "\n" + messageStyle.Render(m.message)
		content += "\n" + helpStyle.Render("enter: continue")
		return boxStyle.Render(content)

	default:
		return ""
	}
}
