package core

type Secured struct {
	ID     string `db:"id"`
	Title  string `db:"title"`
	Fields Fields `db:"-"`
}

func NewSecured(title string) *Secured {
	return &Secured{
		Title: title,
	}
}

func (s *Secured) UpdateTitle(newTitle string) *Secured {
	s.Title = newTitle
	return s
}

func (s *Secured) AddField(field Field) *Secured {
	s.Fields = append(s.Fields, field)
	return s
}
