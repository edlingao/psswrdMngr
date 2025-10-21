package adapter

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/edlingao/psswrdMngr/internal/secured/core"
	"github.com/edlingao/psswrdMngr/internal/secured/ports"
	"github.com/sahilm/fuzzy"
)

type securedViewState int

const (
	securedListView securedViewState = iota
	securedAddFormView
	securedDeleteConfirmView
	securedSuccessView
)

type securedItem struct {
	secured core.Secured
}

func (i securedItem) Title() string       { return i.secured.Title }
func (i securedItem) Description() string { return fmt.Sprintf("ID: %s", i.secured.ID) }
func (i securedItem) FilterValue() string { return i.secured.Title }

type securedItemDelegate struct {
	width int
}

func (d securedItemDelegate) Height() int                             { return 2 }
func (d securedItemDelegate) Spacing() int                            { return 1 }
func (d securedItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d securedItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(securedItem)
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

type SecuredTUI struct {
	SecuredService ports.SecuredServiceOperations
	FieldsService  ports.FieldServiceOperations
	list           list.Model
	titleInput     textinput.Model
	searchInput    textinput.Model
	state          securedViewState
	width          int
	height         int
	message        string
	err            error
	secureds       []core.Secured
	allSecureds    []core.Secured
	selectedID     string
	groupID        *string
	groupName      string
	searchActive   bool
	searchFocused  bool
	parentTUI      tea.Model
	fieldsTUI      *FieldsTUI
}

func NewSecuredTUI(
	securedService ports.SecuredServiceOperations,
	fieldsService ports.FieldServiceOperations,
) *SecuredTUI {
	const defaultWidth = 80
	const listHeight = 14

	l := list.New([]list.Item{}, securedItemDelegate{width: defaultWidth}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	titleInput := textinput.New()
	titleInput.Placeholder = "Enter secured item title"
	titleInput.CharLimit = 256

	searchInput := textinput.New()
	searchInput.Placeholder = "Search secureds..."
	searchInput.CharLimit = 256

	return &SecuredTUI{
		SecuredService: securedService,
		FieldsService:  fieldsService,
		list:           l,
		titleInput:     titleInput,
		searchInput:    searchInput,
		state:          securedListView,
		width:          defaultWidth,
		height:         listHeight,
	}
}

func (m *SecuredTUI) SetParentMenu(menu tea.Model) {
	m.parentTUI = menu
}

func (m *SecuredTUI) SetParentTUI(parent tea.Model) {
	m.parentTUI = parent
}

func (m *SecuredTUI) SetGroup(groupID *string, groupName string) {
	m.groupID = groupID
	m.groupName = groupName
}

func (m *SecuredTUI) SetWindowSize(width, height int) {
	m.width = width
	m.height = height

	listWidth := width - 10
	overhead := 10
	if m.searchActive {
		overhead += 4
	}
	listHeight := height - overhead
	if listHeight < 5 {
		listHeight = 5
	}

	m.list.SetSize(listWidth, listHeight)
	m.list.SetDelegate(securedItemDelegate{width: listWidth})
	m.titleInput.Width = listWidth - 4
	m.searchInput.Width = listWidth - 4
}

func (m *SecuredTUI) returnToParent() tea.Model {
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

func (m *SecuredTUI) loadSecureds() tea.Cmd {
	return func() tea.Msg {
		var secureds []core.Secured
		var err error

		if m.groupID == nil {
			secureds, err = m.SecuredService.GetUngroupedSecureds()
		} else {
			secureds, err = m.SecuredService.GetSecuredsByGroup(*m.groupID)
		}

		if err != nil {
			return securedsLoadedMsg{err: err}
		}
		return securedsLoadedMsg{secureds: secureds}
	}
}

type securedsLoadedMsg struct {
	secureds []core.Secured
	err      error
}

func (m SecuredTUI) Init() tea.Cmd {
	return m.loadSecureds()
}

func (m SecuredTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case securedsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.secureds = msg.secureds
		m.allSecureds = msg.secureds
		items := make([]list.Item, len(msg.secureds))
		for i, s := range msg.secureds {
			items[i] = securedItem{secured: s}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

		listWidth := msg.Width - 10
		overhead := 10
		if m.searchActive {
			overhead += 4
		}
		listHeight := msg.Height - overhead
		if listHeight < 5 {
			listHeight = 5
		}

		m.list.SetSize(listWidth, listHeight)
		m.list.SetDelegate(securedItemDelegate{width: listWidth})
		m.titleInput.Width = listWidth - 4
		m.searchInput.Width = listWidth - 4
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case securedListView:
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
						items := make([]list.Item, len(m.allSecureds))
						for i, s := range m.allSecureds {
							items[i] = securedItem{secured: s}
						}
						m.list.SetItems(items)
						m.secureds = m.allSecureds
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
						items := make([]list.Item, len(m.allSecureds))
						for i, s := range m.allSecureds {
							items[i] = securedItem{secured: s}
						}
						m.list.SetItems(items)
						m.secureds = m.allSecureds
					} else {
						candidates := make([]string, len(m.allSecureds))
						for i, s := range m.allSecureds {
							candidates[i] = s.Title
						}

						matches := fuzzy.Find(query, candidates)
						items := make([]list.Item, len(matches))
						filtered := make([]core.Secured, len(matches))

						for i, match := range matches {
							for _, s := range m.allSecureds {
								if s.Title == match.Str {
									items[i] = securedItem{secured: s}
									filtered[i] = s
									break
								}
							}
						}

						m.list.SetItems(items)
						m.secureds = filtered
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
						items := make([]list.Item, len(m.allSecureds))
						for i, s := range m.allSecureds {
							items[i] = securedItem{secured: s}
						}
						m.list.SetItems(items)
						m.secureds = m.allSecureds
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
						if m.parentTUI != nil {
							return m.returnToParent(), nil
						}
						return m, tea.Quit
					case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
						if len(m.secureds) == 0 {
							return m, nil
						}
						i, ok := m.list.SelectedItem().(securedItem)
						if ok {
							if m.fieldsTUI == nil {
								m.fieldsTUI = NewFieldsTUI(m.SecuredService, m.FieldsService)
							}
							m.fieldsTUI.SetSecured(i.secured)
							m.fieldsTUI.SetParentTUI(m)
							m.fieldsTUI.SetWindowSize(m.width, m.height)
							return m.fieldsTUI, m.fieldsTUI.Init()
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
				m.state = securedAddFormView
				m.titleInput.Focus()
				m.err = nil
				m.message = ""
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				if len(m.secureds) == 0 {
					return m, nil
				}
				i, ok := m.list.SelectedItem().(securedItem)
				if ok {
					m.selectedID = i.secured.ID
					m.state = securedDeleteConfirmView
					return m, nil
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if len(m.secureds) == 0 {
					return m, nil
				}
				i, ok := m.list.SelectedItem().(securedItem)
				if ok {
					if m.fieldsTUI == nil {
						m.fieldsTUI = NewFieldsTUI(m.SecuredService, m.FieldsService)
					}
					m.fieldsTUI.SetSecured(i.secured)
					m.fieldsTUI.SetParentTUI(m)
					m.fieldsTUI.SetWindowSize(m.width, m.height)
					return m.fieldsTUI, m.fieldsTUI.Init()
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case securedAddFormView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = securedListView
				m.titleInput.Blur()
				m.titleInput.SetValue("")
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				title := m.titleInput.Value()
				if title == "" {
					m.err = fmt.Errorf("title cannot be empty")
					return m, nil
				}

				_, err := m.SecuredService.AddSecured(title, m.groupID)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Secured item added successfully!"
				m.state = securedSuccessView
				m.titleInput.SetValue("")
				m.titleInput.Blur()
				return m, m.loadSecureds()
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.titleInput, cmd = m.titleInput.Update(msg)
			return m, cmd

		case securedDeleteConfirmView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
				err := m.SecuredService.DeleteSecured(m.selectedID)
				if err != nil {
					m.err = err
					m.state = securedListView
					return m, nil
				}

				m.message = "Secured item deleted successfully!"
				m.state = securedSuccessView
				m.selectedID = ""
				return m, m.loadSecureds()
			case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc"))):
				m.state = securedListView
				m.selectedID = ""
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentTUI != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}

		case securedSuccessView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc"))):
				m.state = securedListView
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

func (m SecuredTUI) View() string {
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
		Align(lipgloss.Left).
		MarginTop(2)

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("196")).
		Width(listWidth).
		Align(lipgloss.Center).
		MarginTop(1)

	switch m.state {
	case securedListView:
		headerText := "Secured Items"
		if m.groupName != "" {
			headerText = fmt.Sprintf("Secured Items: %s", m.groupName)
		}
		header := headerStyle.Render(headerText)
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

			totalCount := len(m.allSecureds)
			currentCount := len(m.secureds)
			if m.searchInput.Value() == "" {
				currentCount = totalCount
			} else {
				currentCount = len(m.list.Items())
			}
			content += "\n" + countStyle.Render(fmt.Sprintf("%d of %d items", currentCount, totalCount))
		}

		content += "\n" + m.list.View()

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		help := "a: add • d: delete • /: search • enter: view fields • esc: back"
		if m.searchActive {
			if m.searchFocused {
				help = "tab: navigate list • esc: close search"
			} else {
				help = "tab: back to search • enter: select • esc: close search"
			}
		} else if len(m.secureds) == 0 {
			help = "a: add new secured item • esc: back"
		}

		content += "\n" + helpStyle.Render(help)
		return boxStyle.Render(content)

	case securedAddFormView:
		header := headerStyle.Render("Add Secured Item")
		content := header + "\n" + inputStyle.Render(m.titleInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: cancel")
		return boxStyle.Render(content)

	case securedDeleteConfirmView:
		header := headerStyle.Render("Delete Secured Item")

		confirmStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		content := header + "\n" + confirmStyle.Render("Are you sure you want to delete this item?")
		content += "\n" + helpStyle.Render("y: yes • n: no")
		return boxStyle.Render(content)

	case securedSuccessView:
		header := headerStyle.Render("Success")

		messageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("46")).
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		content := header + "\n" + messageStyle.Render(m.message)
		content += "\n" + helpStyle.Render("enter: continue")
		return boxStyle.Render(content)

	default:
		return ""
	}
}
