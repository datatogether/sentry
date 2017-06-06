// TODO - turn "Metadata" into github.com/archivers-space/metablocks.Metablock
package archive

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/archivers-space/sqlutil"
	"github.com/multiformats/go-multihash"
	"time"
)

// CalcHash calculates the multihash key for a given slice of bytes
// TODO - find a proper home for this
func CalcHash(data []byte) (string, error) {
	h := sha256.New()
	h.Write(data)
	mhBuf, err := multihash.EncodeName(h.Sum(nil), "sha2-256")
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(mhBuf), nil
}

// A snapshot is a record of a GET request to a url
// There can be many metadata of a given url
type Metadata struct {
	// Hash is the sha256 multihash of all other fields in metadata
	// as expressed by Metadata.HashableBytes()
	Hash string `json:"hash"`
	// Creation timestamp
	Timestamp time.Time `json:"timestamp"`
	// Sha256 multihash of the public key that signed this metadata
	KeyId string `json:"keyId"`
	// Sha256 multihash of the content this metadata is describing
	Subject string `json:"subject"`
	// Hash value of the metadata that came before this, if any
	Prev string `json:"prev"`
	// Acutal metadata, a valid json Object
	Meta map[string]interface{} `json:"meta"`
}

// String is metadata's abbreviated string representation
func (m Metadata) String() string {
	return fmt.Sprintf("%s : %s.%s", m.Hash, m.KeyId, m.Subject)
}

// MetadatasBySubject returns all metadata for a given subject hash
func MetadataBySubject(db sqlutil.Queryable, subject string) ([]*Metadata, error) {
	res, err := db.Query(qMetadataForSubject, subject)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	metadata := make([]*Metadata, 0)
	for res.Next() {
		m := &Metadata{}
		if err := m.UnmarshalSQL(res); err != nil {
			return nil, err
		}
		metadata = append(metadata, m)
	}

	return metadata, nil
}

func MetadataCountByKey(db sqlutil.Queryable, keyId string) (count int, err error) {
	err = db.QueryRow(qMetadataCountForKey, keyId).Scan(&count)
	return
}

func MetadataByKey(db sqlutil.Queryable, keyId string, limit, offset int) ([]*Metadata, error) {
	rows, err := db.Query(qMetadataLatestForKey, keyId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]*Metadata, limit)
	i := 0
	for rows.Next() {
		m := &Metadata{}
		if err := m.UnmarshalSQL(rows); err != nil {
			return nil, err
		}
		results[i] = m
		i++
	}

	return results[:i], nil
}

// LatestMetadata gives the most recent metadata timestamp for a given keyId & subject
// combination if one exists
func LatestMetadata(db sqlutil.Queryable, keyId, subject string) (m *Metadata, err error) {
	row := db.QueryRow(qMetadataLatest, keyId, subject)
	m = &Metadata{}
	err = m.UnmarshalSQL(row)
	return
}

// NextMetadata returns the next metadata block for a given subject. If no metablock
// exists a new one is created
func NextMetadata(db sqlutil.Queryable, keyId, subject string) (*Metadata, error) {
	m, err := LatestMetadata(db, keyId, subject)
	if err != nil {
		if err == ErrNotFound {
			return &Metadata{
				KeyId:   keyId,
				Subject: subject,
				Meta:    map[string]interface{}{},
			}, nil
		} else {
			return nil, err
		}
	}

	return &Metadata{
		KeyId:   m.KeyId,
		Subject: m.Subject,
		Prev:    m.Hash,
		Meta:    m.Meta,
	}, nil
}

// HashableBytes returns the exact structure to be used for hash
func (m *Metadata) HashableBytes() ([]byte, error) {
	hash := struct {
		Timestamp time.Time              `json:"timestamp"`
		KeyId     string                 `json:"keyId"`
		Subject   string                 `json:"subject"`
		Prev      string                 `json:"prev"`
		Meta      map[string]interface{} `json:"meta"`
	}{
		Timestamp: m.Timestamp,
		KeyId:     m.KeyId,
		Subject:   m.Subject,
		Prev:      m.Prev,
		Meta:      m.Meta,
	}
	return json.Marshal(&hash)
}

func (m *Metadata) calcHash() error {
	data, err := m.HashableBytes()
	if err != nil {
		return err
	}

	h := sha256.New()
	h.Write(data)

	mhBuf, err := multihash.EncodeName(h.Sum(nil), "sha2-256")
	if err != nil {
		return err
	}

	m.Hash = hex.EncodeToString(mhBuf)
	return nil
}

// WriteMetadata creates a snapshot record in the DB from a given Url struct
func (m *Metadata) Write(db sqlutil.Execable) error {
	// TODO - check for valid subject hash

	m.Timestamp = time.Now().Round(time.Second)
	if err := m.calcHash(); err != nil {
		return err
	}
	metaBytes, err := json.Marshal(m.Meta)
	if err != nil {
		return err
	}

	_, err = db.Exec(qMetadataInsert, m.Hash, m.Timestamp.In(time.UTC).Round(time.Second), m.KeyId, m.Subject, m.Prev, metaBytes)

	if str, ok := m.Meta["title"].(string); ok && str != "" {
		go func() {
			u := &Url{Hash: m.Subject}
			if err := u.Read(db); err != nil {
				return
			}

			// TODO - this is a straight set, should be derived from consensus calculation
			u.Title = str
			if err := u.Update(db); err != nil {
				return
			}
		}()
	}

	return err
}

// UnmarshalSQL reads an SQL result into the snapshot receiver
func (m *Metadata) UnmarshalSQL(row sqlutil.Scannable) error {
	var (
		hash, keyId, subject, prev string
		timestamp                  time.Time
		metaBytes                  []byte
	)

	if err := row.Scan(&hash, &timestamp, &keyId, &subject, &prev, &metaBytes); err != nil {
		if err == sql.ErrNoRows {
			return ErrNotFound
		}
		return err
	}

	var meta map[string]interface{}
	if metaBytes != nil {
		if err := json.Unmarshal(metaBytes, &meta); err != nil {
			return err
		}
	}

	*m = Metadata{
		Hash:      hash,
		Timestamp: timestamp,
		KeyId:     keyId,
		Subject:   subject,
		Prev:      prev,
		Meta:      meta,
	}

	return nil
}

// TODO - this is ripped from metablocks
func (m *Metadata) HashMaps() (keyMap map[string]string, valueMap map[string]interface{}, err error) {
	var (
		value []byte
		hash  string
	)

	keyMap = map[string]string{}
	valueMap = map[string]interface{}{}

	if m.Meta == nil {
		err = fmt.Errorf("metablock has no metadata calculate hashmaps from")
		return
	}

	for k, v := range m.Meta {
		value, err = json.Marshal(v)
		if err != nil {
			return
		}

		hash, err = CalcHash(value)
		if err != nil {
			return
		}

		keyMap[k] = hash
		valueMap[hash] = v
	}

	return
}
