package core

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/ipfs/go-datastore"
	"github.com/pborman/uuid"
	"time"
)

// Primer is tracking information about an abstract group of content.
// For example a government agency is a primer
type Primer struct {
	// version 4 uuid
	Id string `json:"id"`
	// Created timestamp rounded to seconds in UTC
	Created time.Time `json:"created"`
	// Updated timestamp rounded to seconds in UTC
	Updated time.Time `json:"updated"`
	// shortest possible expression of this primer's name, usually an acronym
	// called shortTitle b/c acronyms collide often & users should feel free to
	// expand on acronyms
	ShortTitle string `json:"shortTitle"`
	// human-readable title of this primer.
	Title string `json:"title"`
	// long-form description of this primer.
	// TODO - Maybe we should store this in markdown format?
	Description string `json:"description"`
	// parent primer (if any)
	Parent *Primer `json:"parent"`
	// child-primers list
	SubPrimers []*Primer `json:"subPrimers,omitempty"`
	// metadata to associate with this primer
	Meta map[string]interface{} `json:"meta"`
	// statistics about this primer
	Stats *PrimerStats `json:"stats"`
	// collection of child sources
	Sources []*Source `json:"sources,omitempty"`
}

// TODO - finish
type PrimerStats struct {
	UrlCount                int `json:"urlCount"`
	ArchivedUrlCount        int `json:"archivedUrlCount"`
	ContentUrlCount         int `json:"contentUrlCount"`
	ContentMetadataCount    int `json:"contentMetadataCount"`
	SourcesUrlCount         int `json:"sourcesUrlCount"`
	SourcesArchivedUrlCount int `json:"sourcesArchivedUrlCount"`
}

func (p Primer) DatastoreType() string {
	return "Primer"
}

func (p Primer) GetId() string {
	return p.Id
}

func (p Primer) Key() datastore.Key {
	return datastore.NewKey(fmt.Sprintf("%s:%s", p.DatastoreType(), p.GetId()))
}

// ReadSubPrimers reads child primers of this primer
func (p *Primer) ReadSubPrimers(db sqlutil.Queryable) error {
	rows, err := db.Query(qPrimerSubPrimers, p.Id)
	if err != nil {
		return err
	}

	defer rows.Close()
	sps := make([]*Primer, 0)
	for rows.Next() {
		c := &Primer{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return err
		}
		sps = append(sps, c)
	}

	p.SubPrimers = sps
	return nil
}

func (p *Primer) CalcStats(db *sql.DB) error {
	p.Stats = &PrimerStats{}
	if err := p.ReadSources(db); err != nil {
		return err
	}
	for _, s := range p.Sources {
		if err := s.CalcStats(db); err != nil {
			return err
		}
		p.Stats.UrlCount += s.Stats.UrlCount
		p.Stats.ContentMetadataCount += s.Stats.ContentMetadataCount
		p.Stats.ContentUrlCount += s.Stats.ContentUrlCount
	}

	if err := p.ReadSubPrimers(db); err != nil {
		return err
	}
	for _, sp := range p.SubPrimers {
		if err := sp.CalcStats(db); err != nil {
			return err
		}
		p.Stats.UrlCount += sp.Stats.UrlCount
		p.Stats.ContentMetadataCount += sp.Stats.ContentMetadataCount
		p.Stats.ContentUrlCount += sp.Stats.ContentUrlCount
	}

	store := sql_datastore.NewDatastore(db)
	if err := store.Register(&Primer{}); err != nil {
		return err
	}

	return p.Save(store)
}

// ReadSources reads child sources of this primer
func (p *Primer) ReadSources(db sqlutil.Queryable) error {
	rows, err := db.Query(qPrimerSources, p.Id)
	if err != nil {
		return err
	}

	defer rows.Close()
	s := make([]*Source, 0)
	for rows.Next() {
		c := &Source{}
		if err := c.UnmarshalSQL(rows); err != nil {
			return err
		}
		s = append(s, c)
	}

	p.Sources = s
	return nil
}

func (p *Primer) Read(store datastore.Datastore) error {
	pi, err := store.Get(p.Key())
	if err != nil {
		if err == datastore.ErrNotFound {
			return ErrNotFound
		}
		return err
	}

	got, ok := pi.(*Primer)
	if !ok {
		return ErrInvalidResponse
	}
	*p = *got
	return nil
}

func (p *Primer) Save(store datastore.Datastore) (err error) {
	var exists bool

	if p.Id != "" {
		exists, err = store.Has(p.Key())
		if err != nil {
			return err
		}
	}

	if !exists {
		p.Id = uuid.New()
		p.Created = time.Now().Round(time.Second)
		p.Updated = p.Created
	} else {
		p.Updated = time.Now().Round(time.Second)
	}

	return store.Put(p.Key(), p)
}

func (p *Primer) Delete(store datastore.Datastore) error {
	return store.Delete(p.Key())
}

func (p *Primer) NewSQLModel(key datastore.Key) sql_datastore.Model {
	return &Primer{Id: key.Name()}
}

func (p *Primer) SQLQuery(cmd sql_datastore.Cmd) string {
	switch cmd {
	case sql_datastore.CmdCreateTable:
		return qPrimerCreateTable
	case sql_datastore.CmdExistsOne:
		return qPrimerExists
	case sql_datastore.CmdSelectOne:
		return qPrimerById
	case sql_datastore.CmdInsertOne:
		return qPrimerInsert
	case sql_datastore.CmdUpdateOne:
		return qPrimerUpdate
	case sql_datastore.CmdDeleteOne:
		return qPrimerDelete
	case sql_datastore.CmdList:
		return qPrimersList
	default:
		return ""
	}
}

func (p *Primer) SQLParams(cmd sql_datastore.Cmd) []interface{} {
	switch cmd {
	case sql_datastore.CmdSelectOne, sql_datastore.CmdExistsOne, sql_datastore.CmdDeleteOne:
		return []interface{}{p.Id}
	case sql_datastore.CmdList:
		return []interface{}{}
	default:
		parentId := ""
		if p.Parent != nil {
			parentId = p.Parent.Id
		}

		metaBytes, err := json.Marshal(p.Meta)
		if err != nil {
			panic(err)
		}

		statBytes, err := json.Marshal(p.Stats)
		if err != nil {
			panic(err)
		}

		return []interface{}{
			p.Id,
			p.Created.In(time.UTC),
			p.Updated.In(time.UTC),
			p.ShortTitle,
			p.Title,
			p.Description,
			parentId,
			statBytes,
			metaBytes,
		}
	}
}

func (p *Primer) UnmarshalSQL(row sqlutil.Scannable) error {
	var (
		parent                                  *Primer
		id, title, description, short, parentId string
		created, updated                        time.Time
		statsBytes, metaBytes                   []byte
		meta                                    map[string]interface{}
		stats                                   *PrimerStats
	)

	if err := row.Scan(&id, &created, &updated, &short, &title, &description, &parentId, &statsBytes, &metaBytes); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	if parentId != "" {
		parent = &Primer{Id: parentId}
	}

	if metaBytes != nil {
		if err := json.Unmarshal(metaBytes, &meta); err != nil {
			return err
		}
	}

	if statsBytes != nil {
		stats = &PrimerStats{}
		if err := json.Unmarshal(statsBytes, stats); err != nil {
			return err
		}
	}

	*p = Primer{
		Id:          id,
		Created:     created.In(time.UTC),
		Updated:     updated.In(time.UTC),
		ShortTitle:  short,
		Title:       title,
		Description: description,
		Parent:      parent,
		Meta:        meta,
		Stats:       stats,
	}

	return nil
}
