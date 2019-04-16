package database

import "net/http"

// SourceHandler represents
type SourceHandler interface {
	Create(*Source) (*Source, error)
	Update(*Source) (*Source, error)
	Delete(*Source) error
	Request(http.ResponseWriter, *http.Request)
}

/*
func createSource(db *AdminDB, s *Source) (string, error) {
	if s.Type == nil {
		return "",ErrBadQuery("Must specify a source type")
	}
	stype, ok := db.Assets().GetSourceType(*s.Type)
	if !ok {
		return "",ErrBadQuery("Unrecognized source type")
	}
}
*/
