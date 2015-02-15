package main

type Repository struct {
	Dir string
}

func (r *Repository) IndexPath() string {
	return r.Dir + "/_index"
}
