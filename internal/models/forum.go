package models

type Forum struct {
	Title   string `json:"title,omitempty"`
	User    string `json:"user,omitempty"`
	Slug    string `json:"slug,omitempty"`
	Posts   int64  `json:"posts,omitempty"`
	Threads int32  `json:"threads,omitempty"`
}
