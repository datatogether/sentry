package archive

import (
	"database/sql"
	"encoding/json"
	"github.com/pborman/uuid"
	"time"
)

// Primer is tracking information about a base URL
type Primer struct {
	Id          string                 `json:"id"`
	Created     time.Time              `json:"created"`
	Updated     time.Time              `json:"updated"`
	ShortTitle  string                 `json:"shortTitle"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Parent      *Primer                `json:"parent"`
	SubPrimers  []*Primer              `json:"subPrimers,omitempty"`
	Meta        map[string]interface{} `json:"meta"`
	Stats       *PrimerStats           `json:"stats"`
	Sources     []*Source              `json:"sources"`
}

// TODO - finish
type PrimerStats struct {
	UrlCount             int `json:"urlCount"`
	ContentUrlCount      int `json:"contentUrlCount"`
	ContentMetadataCount int `json:"contentMetadataCount"`
}

// ReadSubPrimers reads child primers of this primer
func (p *Primer) ReadSubPrimers(db sqlQueryable) error {
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

// ReadSources reads child sources of this primer
func (p *Primer) ReadSources(db sqlQueryable) error {
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

func (p *Primer) Read(db sqlQueryable) error {
	if p.Id != "" {
		row := db.QueryRow(qPrimerById, p.Id)
		return p.UnmarshalSQL(row)
	}
	return ErrNotFound
}

func (p *Primer) Save(db sqlQueryExecable) error {
	prev := &Primer{Id: p.Id}
	if err := prev.Read(db); err != nil {
		if err == ErrNotFound {
			p.Id = uuid.New()
			p.Created = time.Now().Round(time.Second)
			p.Updated = p.Created
			_, err := db.Exec(qPrimerInsert, p.SQLArgs()...)
			return err
		} else {
			return err
		}
	} else {
		p.Updated = time.Now().Round(time.Second)
		_, err := db.Exec(qPrimerUpdate, p.SQLArgs()...)
		return err
	}
	return nil
}

func (p *Primer) Delete(db sqlQueryExecable) error {
	_, err := db.Exec(qPrimerDelete, p.Id)
	return err
}

func (p *Primer) UnmarshalSQL(row sqlScannable) error {
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

func (p *Primer) SQLArgs() []interface{} {

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
