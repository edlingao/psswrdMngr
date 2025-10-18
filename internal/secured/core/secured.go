package core

type Secured struct {
	ID      string  `db:"id"`
	Title   string  `db:"title"`
	GroupID *string `db:"group_id"`
	Fields  Fields  `db:"-"`
}

func NewSecured(title string, groupID *string) *Secured {
	return &Secured{
		Title:   title,
		GroupID: groupID,
	}
}

func (s *Secured) UpdateTitle(newTitle string) *Secured {
	s.Title = newTitle
	return s
}

func (s *Secured) SetGroup(groupID *string) *Secured {
	s.GroupID = groupID
	return s
}

func (s *Secured) AddField(field Field) *Secured {
	s.Fields = append(s.Fields, field)
	return s
}
