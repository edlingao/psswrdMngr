package adapter

import (
	"fmt"
	"io"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/edlingao/psswrdMngr/internal/group/core"
	"github.com/edlingao/psswrdMngr/internal/group/ports"
	"github.com/sahilm/fuzzy"
)

type groupViewState int

const (
	groupListView groupViewState = iota
	groupAddFormView
	groupRenameFormView
	groupDeleteConfirmView
	groupSuccessView
)

type groupItem struct {
	group      *core.Group
	isUngrouped bool
}

func (i groupItem) Title() string {
	if i.isUngrouped {
		return "Ungrouped"
	}
	return i.group.Name
}

func (i groupItem) Description() string {
	if i.isUngrouped {
		return "Items without a group"
	}
	return fmt.Sprintf("ID: %s", i.group.ID)
}

func (i groupItem) FilterValue() string {
	if i.isUngrouped {
		return "Ungrouped"
	}
	return i.group.Name
}

type groupItemDelegate struct {
	width int
}

func (d groupItemDelegate) Height() int                             { return 2 }
func (d groupItemDelegate) Spacing() int                            { return 1 }
func (d groupItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d groupItemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(groupItem)
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

type GroupTUI struct {
	GroupService  ports.GroupServiceOperations
	list          list.Model
	nameInput     textinput.Model
	searchInput   textinput.Model
	state         groupViewState
	width         int
	height        int
	message       string
	err           error
	groups        []core.Group
	allGroups     []core.Group
	selectedGroup *core.Group
	searchActive  bool
	searchFocused bool
	parentMenu    tea.Model
	securedTUI    tea.Model
}

func NewGroupTUI(
	groupService ports.GroupServiceOperations,
) *GroupTUI {
	const defaultWidth = 80
	const listHeight = 14

	l := list.New([]list.Item{}, groupItemDelegate{width: defaultWidth}, defaultWidth, listHeight)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)

	nameInput := textinput.New()
	nameInput.Placeholder = "Enter group name"
	nameInput.CharLimit = 256

	searchInput := textinput.New()
	searchInput.Placeholder = "Search groups..."
	searchInput.CharLimit = 256

	return &GroupTUI{
		GroupService: groupService,
		list:         l,
		nameInput:    nameInput,
		searchInput:  searchInput,
		state:        groupListView,
		width:        defaultWidth,
		height:       listHeight,
	}
}

func (m *GroupTUI) SetParentMenu(menu tea.Model) {
	m.parentMenu = menu
}

func (m *GroupTUI) SetSecuredTUI(securedTUI tea.Model) {
	m.securedTUI = securedTUI
}

func (m *GroupTUI) SetWindowSize(width, height int) {
	m.width = width
	m.height = height

	listWidth := width - 10
	listHeight := 10
	if height < 20 {
		listHeight = height - 10
	}

	m.list.SetSize(listWidth, listHeight)
	m.list.SetDelegate(groupItemDelegate{width: listWidth})
	m.nameInput.Width = listWidth - 4
	m.searchInput.Width = listWidth - 4
}

func (m *GroupTUI) returnToParent() tea.Model {
	if m.parentMenu == nil {
		return m
	}

	if sizer, ok := m.parentMenu.(interface {
		Update(tea.Msg) (tea.Model, tea.Cmd)
	}); ok {
		updated, _ := sizer.Update(tea.WindowSizeMsg{
			Width:  m.width,
			Height: m.height,
		})
		return updated
	}

	return m.parentMenu
}

func (m *GroupTUI) loadGroups() tea.Cmd {
	return func() tea.Msg {
		groups, err := m.GroupService.GetAllGroups()
		if err != nil {
			return groupsLoadedMsg{err: err}
		}
		return groupsLoadedMsg{groups: groups}
	}
}

type groupsLoadedMsg struct {
	groups []core.Group
	err    error
}

func (m GroupTUI) Init() tea.Cmd {
	return m.loadGroups()
}

func (m GroupTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case groupsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.groups = msg.groups
		m.allGroups = msg.groups
		items := make([]list.Item, 0, len(msg.groups)+1)
		items = append(items, groupItem{isUngrouped: true})
		for _, g := range msg.groups {
			group := g
			items = append(items, groupItem{group: &group})
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
		m.list.SetDelegate(groupItemDelegate{width: listWidth})
		m.nameInput.Width = listWidth - 4
		m.searchInput.Width = listWidth - 4
		return m, nil

	case tea.KeyMsg:
		switch m.state {
		case groupListView:
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
						items := make([]list.Item, 0, len(m.allGroups)+1)
						items = append(items, groupItem{isUngrouped: true})
						for _, g := range m.allGroups {
							group := g
							items = append(items, groupItem{group: &group})
						}
						m.list.SetItems(items)
						m.groups = m.allGroups
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
						if m.parentMenu != nil {
							return m.returnToParent(), nil
						}
						return m, tea.Quit
					default:
						var cmd tea.Cmd
						m.searchInput, cmd = m.searchInput.Update(msg)

					query := m.searchInput.Value()
					if query == "" {
						items := make([]list.Item, 0, len(m.allGroups)+1)
						items = append(items, groupItem{isUngrouped: true})
						for _, g := range m.allGroups {
							group := g
							items = append(items, groupItem{group: &group})
						}
						m.list.SetItems(items)
						m.groups = m.allGroups
					} else {
						candidates := make([]string, 0, len(m.allGroups)+1)
						candidates = append(candidates, "Ungrouped")
						for _, g := range m.allGroups {
							candidates = append(candidates, g.Name)
						}

						matches := fuzzy.Find(query, candidates)
						items := make([]list.Item, 0, len(matches))
						filtered := make([]core.Group, 0)

						for _, match := range matches {
							if match.Str == "Ungrouped" {
								items = append(items, groupItem{isUngrouped: true})
							} else {
								for _, g := range m.allGroups {
									if g.Name == match.Str {
										group := g
										items = append(items, groupItem{group: &group})
										filtered = append(filtered, g)
										break
									}
								}
							}
						}

						m.list.SetItems(items)
						m.groups = filtered
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
						items := make([]list.Item, 0, len(m.allGroups)+1)
						items = append(items, groupItem{isUngrouped: true})
						for _, g := range m.allGroups {
							group := g
							items = append(items, groupItem{group: &group})
						}
						m.list.SetItems(items)
						m.groups = m.allGroups
						return m, nil
					case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
						if m.parentMenu != nil {
							return m.returnToParent(), nil
						}
						return m, tea.Quit
					case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
						i, ok := m.list.SelectedItem().(groupItem)
						if ok && m.securedTUI != nil {
							var groupID *string
							var groupName string
							if i.isUngrouped {
								groupID = nil
								groupName = "Ungrouped"
							} else {
								groupID = &i.group.ID
								groupName = i.group.Name
							}

							if setter, ok := m.securedTUI.(interface{ SetGroup(groupID *string, groupName string) }); ok {
								setter.SetGroup(groupID, groupName)
							}
							if setter, ok := m.securedTUI.(interface{ SetWindowSize(int, int) }); ok {
								setter.SetWindowSize(m.width, m.height)
							}
							if setter, ok := m.securedTUI.(interface{ SetParentTUI(tea.Model) }); ok {
								setter.SetParentTUI(m)
							}
							if initer, ok := m.securedTUI.(interface{ Init() tea.Cmd }); ok {
								return m.securedTUI, initer.Init()
							}
							return m.securedTUI, nil
						}
					}
					var cmd tea.Cmd
					m.list, cmd = m.list.Update(msg)
					return m, cmd
				}
			}

			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc", "q", "ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			case key.Matches(msg, key.NewBinding(key.WithKeys("/"))):
				m.searchActive = true
				m.searchFocused = true
				m.searchInput.Focus()
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("a"))):
				m.state = groupAddFormView
				m.nameInput.Focus()
				m.err = nil
				m.message = ""
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("r"))):
				i, ok := m.list.SelectedItem().(groupItem)
				if ok && !i.isUngrouped && i.group != nil {
					m.selectedGroup = i.group
					m.nameInput.SetValue(i.group.Name)
					m.state = groupRenameFormView
					m.nameInput.Focus()
					m.err = nil
					m.message = ""
					return m, textinput.Blink
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("d"))):
				i, ok := m.list.SelectedItem().(groupItem)
				if ok && !i.isUngrouped && i.group != nil {
					m.selectedGroup = i.group
					m.state = groupDeleteConfirmView
					return m, nil
				}
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				i, ok := m.list.SelectedItem().(groupItem)
				if ok && m.securedTUI != nil {
					var groupID *string
					var groupName string
					if i.isUngrouped {
						groupID = nil
						groupName = "Ungrouped"
					} else {
						groupID = &i.group.ID
						groupName = i.group.Name
					}

					if setter, ok := m.securedTUI.(interface{ SetGroup(groupID *string, groupName string) }); ok {
						setter.SetGroup(groupID, groupName)
					}
					if setter, ok := m.securedTUI.(interface{ SetWindowSize(int, int) }); ok {
						setter.SetWindowSize(m.width, m.height)
					}
					if setter, ok := m.securedTUI.(interface{ SetParentTUI(tea.Model) }); ok {
						setter.SetParentTUI(m)
					}
					if initer, ok := m.securedTUI.(interface{ Init() tea.Cmd }); ok {
						return m.securedTUI, initer.Init()
					}
					return m.securedTUI, nil
				}
			}
			var cmd tea.Cmd
			m.list, cmd = m.list.Update(msg)
			return m, cmd

		case groupAddFormView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = groupListView
				m.nameInput.Blur()
				m.nameInput.SetValue("")
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				name := m.nameInput.Value()
				if name == "" {
					m.err = fmt.Errorf("group name cannot be empty")
					return m, nil
				}

				_, err := m.GroupService.AddGroup(name)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Group added successfully!"
				m.state = groupSuccessView
				m.nameInput.SetValue("")
				m.nameInput.Blur()
				return m, m.loadGroups()
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd

		case groupRenameFormView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("esc"))):
				m.state = groupListView
				m.nameInput.Blur()
				m.nameInput.SetValue("")
				m.selectedGroup = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				name := m.nameInput.Value()
				if name == "" {
					m.err = fmt.Errorf("group name cannot be empty")
					return m, nil
				}

				_, err := m.GroupService.UpdateGroupName(m.selectedGroup.ID, name)
				if err != nil {
					m.err = err
					return m, nil
				}

				m.message = "Group renamed successfully!"
				m.state = groupSuccessView
				m.nameInput.SetValue("")
				m.nameInput.Blur()
				m.selectedGroup = nil
				return m, m.loadGroups()
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd

		case groupDeleteConfirmView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("y"))):
				err := m.GroupService.DeleteGroup(m.selectedGroup.ID)
				if err != nil {
					m.err = err
					m.state = groupListView
					return m, nil
				}

				m.message = "Group deleted successfully!"
				m.state = groupSuccessView
				m.selectedGroup = nil
				return m, m.loadGroups()
			case key.Matches(msg, key.NewBinding(key.WithKeys("n", "esc"))):
				m.state = groupListView
				m.selectedGroup = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}

		case groupSuccessView:
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter", "esc"))):
				m.state = groupListView
				m.message = ""
				m.err = nil
				return m, nil
			case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))):
				if m.parentMenu != nil {
					return m.returnToParent(), nil
				}
				return m, tea.Quit
			}
		}
	}

	return m, nil
}

func (m GroupTUI) View() string {
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

	switch m.state {
	case groupListView:
		header := headerStyle.Render("Groups")
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

			totalCount := len(m.allGroups) + 1
			currentCount := len(m.groups)
			if m.searchInput.Value() == "" {
				currentCount = totalCount
			} else {
				currentCount = len(m.list.Items())
			}
			content += "\n" + countStyle.Render(fmt.Sprintf("%d of %d groups", currentCount, totalCount))
		}

		content += "\n" + m.list.View()

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		help := "a: add • r: rename • d: delete • /: search • enter: view items • esc: back"
		if m.searchActive {
			if m.searchFocused {
				help = "tab: navigate list • esc: close search"
			} else {
				help = "tab: back to search • enter: select • esc: close search"
			}
		}
		content += "\n" + helpStyle.Render(help)
		return boxStyle.Render(content)

	case groupAddFormView:
		header := headerStyle.Render("Add Group")
		content := header + "\n" + inputStyle.Render(m.nameInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: cancel")
		return boxStyle.Render(content)

	case groupRenameFormView:
		header := headerStyle.Render("Rename Group")
		content := header + "\n" + inputStyle.Render(m.nameInput.View())

		if m.err != nil {
			content += "\n" + errorStyle.Render(m.err.Error())
		}

		content += "\n" + helpStyle.Render("enter: save • esc: cancel")
		return boxStyle.Render(content)

	case groupDeleteConfirmView:
		header := headerStyle.Render("Delete Group")

		confirmStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			MarginTop(1)

		warningStyle := lipgloss.NewStyle().
			Width(listWidth).
			Align(lipgloss.Center).
			Foreground(lipgloss.Color("196")).
			MarginTop(1)

		content := header + "\n" + confirmStyle.Render(fmt.Sprintf("Delete group '%s'?", m.selectedGroup.Name))
		content += "\n" + warningStyle.Render("All items in this group will be deleted!")
		content += "\n" + helpStyle.Render("y: yes • n: no")
		return boxStyle.Render(content)

	case groupSuccessView:
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
