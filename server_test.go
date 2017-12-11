package main

import (
	"os"
	"fmt"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/datatogether/core"
	"github.com/datatogether/sql_datastore"
	"github.com/datatogether/sqlutil"
	"github.com/gchaincl/dotsql"
)

func setUp() {
	var err error
	cfg, err = initConfig( "test" )
	if err != nil {
		// panic if the server is missing a vital configuration detail
		panic( fmt.Errorf( "server configuration error: %s", err.Error() ) )
	}

	sqlutil.ConnectToDb( "postgres", cfg.PostgresDbUrl, appDB )
	sql_datastore.SetDB( appDB )
	sql_datastore.Register(
		&core.Url{},
		&core.Link{},
	)

	// create any tables if they don't exist
	sc, err := sqlutil.LoadSchemaCommands( packagePath( "sql/schema.sql" ) )
	if err != nil {
		fmt.Errorf( "error loading schema file: %s", err )
	} else {
		created, err := sc.Create( appDB, 
						"primers",
						"sources",
						"urls",
						"links",
						"metadata",
						"snapshots",
						"collections",
						"archive_requests",
						"uncrawlables",
						"data_repos" )
		if err != nil {
			fmt.Errorf( "error creating missing tables: %s", err )
		} else if len(created) > 0 {
			fmt.Errorf( "created tables:", created )
		}
	}

	data, err := sqlutil.LoadDataCommands( packagePath( "sql/test_data.sql" ) )
	if err != nil {
		fmt.Errorf( "error loading commands file: %s", err )
	} else {
		err := data.Reset( appDB, "primers", "sources", "urls" )
		if err != nil {
			fmt.Errorf( "error creating missing tables: %s", err )
		}
	}
	
}

func cleanUp() {
	d, err := dotsql.LoadFromFile( packagePath( "sql/test_data.sql" ) )
	if err != nil {
		fmt.Errorf( "error loading schema file: %s", err )
	} else {
		tables := [...]string{ "primers", "sources", "urls" };
		for _, t := range tables {
			if _, err := d.Exec( appDB, fmt.Sprintf( "delete-%s", t ) ); err != nil {
				fmt.Errorf( "error executing 'delete-%s': %s", t, err )
			}
		}
	}
}

func TestMain( m *testing.M ) { 
    setUp()
    ret := m.Run()
    cleanUp()
    os.Exit( ret )
}

func TestServerRoutes(t *testing.T) {
	cases := []struct {
		method, endpoint string
		expectBody       bool
		body             []byte
		resStatus        int
	}{
		//TODO: switch from [A] to [B]: these pass currently but POST, PUT, DELETE should return 404 (http.StatusNotFound ) instead of 200 (http.StatusOK)
		// [A]
		{"GET", "/", false, nil, http.StatusOK},
		{"PUT", "/", false, nil, http.StatusOK},
		{"POST", "/", false, nil, http.StatusOK},
		{"DELETE", "/", false, nil, http.StatusOK},
		// [B]
		// {"GET", "/", false, nil, http.StatusOK},
		// {"PUT", "/", false, nil, http.StatusNotFound},
		// {"POST", "/", false, nil, http.StatusNotFound},
		// {"DELETE", "/", false, nil, http.StatusNotFound},

		//TODO: switch from [A] to [B]: POST, PUT, DELETE should return 404 (http.StatusNotFound ) instead of 200 (http.StatusOK)
		// [A]
		{"GET", "/healthcheck", false, nil, http.StatusOK},
		{"PUT", "/healthcheck", false, nil, http.StatusOK},
		{"POST", "/healthcheck", false, nil, http.StatusOK},
		{"DELETE", "/healthcheck", false, nil, http.StatusOK},
		// [B]
		// {"GET", "/healthcheck", false, nil, http.StatusOK},
		// {"PUT", "/healthcheck", false, nil, http.StatusNotFound},
		// {"POST", "/healthcheck", false, nil, http.StatusNotFound},
		// {"DELETE", "/healthcheck", false, nil, http.StatusNotFound},

		//TODO: these pass currently but GET should return 404 200 (http.StatusOK) instead of 500 (http.StatusInternalServerError)
		// [A]
		{"GET", "/urls", false, nil, http.StatusOK},
		{"PUT", "/urls", false, nil, http.StatusNotFound},
		{"POST", "/urls", false, nil, http.StatusNotFound},
		{"DELETE", "/urls", false, nil, http.StatusNotFound},
		// [B]
		// {"GET", "/urls", false, nil, http.StatusInternalServerError},
		// {"PUT", "/urls", false, nil, http.StatusNotFound},
		// {"POST", "/urls", false, nil, http.StatusNotFound},
		// {"DELETE", "/urls", false, nil, http.StatusNotFound},

		//TODO: these pass currently but POST, PUT, DELETE should return 404 (http.StatusNotFound ) instead of 200 (http.StatusOK)
		// [A]
		{"GET", "/sources", false, nil, http.StatusOK},
		{"PUT", "/sources", false, nil, http.StatusOK},
		{"POST", "/sources", false, nil, http.StatusOK},
		{"DELETE", "/sources", false, nil, http.StatusOK},
		// [B]
		// {"GET", "/sources", false, nil, http.StatusOK},
		// {"PUT", "/sources", false, nil, http.StatusNotFound},
		// {"POST", "/sources", false, nil, http.StatusNotFound},
		// {"DELETE", "/sources", false, nil, http.StatusNotFound},

		//TODO: these pass currently but POST, PUT, DELETE should return 404 (http.StatusNotFound ) instead of 200 (http.StatusOK)
		// [A]
		{"GET", "/mem", false, nil, http.StatusOK},
		{"PUT", "/mem", false, nil, http.StatusOK},
		{"POST", "/mem", false, nil, http.StatusOK},
		{"DELETE", "/mem", false, nil, http.StatusOK},
		// [B]
		// {"GET", "/mem", false, nil, http.StatusOK},
		// {"PUT", "/mem", false, nil, http.StatusNotFound},
		// {"POST", "/mem", false, nil, http.StatusNotFound},
		// {"DELETE", "/mem", false, nil, http.StatusNotFound},

		//TODO: these pass currently but POST should return 200 (http.StatusOK) instead of 400 (http.StatusBadRequest)
		// [A]
		{"GET", "/que", false, nil, http.StatusOK},
		{"PUT", "/que", false, nil, http.StatusNotFound},
		{"POST", "/que", false, nil, http.StatusBadRequest},
		{"DELETE", "/que", false, nil, http.StatusNotFound},
		// [B]
		// {"GET", "/que", false, nil, http.StatusOK},
		// {"PUT", "/que", false, nil, http.StatusNotFound},
		// {"POST", "/que", false, nil, http.StatusOK},
		// {"DELETE", "/que", false, nil, http.StatusNotFound},

		//TODO: these pass currently but POST, PUT, DELETE should return 404 (http.StatusNotFound ) instead of 200 (http.StatusOK)
		// [A]
		{"GET", "/shutodwn", false, nil, http.StatusOK},
		{"PUT", "/shutodwn", false, nil, http.StatusOK},
		{"POST", "/shutodwn", false, nil, http.StatusOK},
		{"DELETE", "/shutodwn", false, nil, http.StatusOK},
		// [B]
		// {"GET", "/shutodwn", false, nil, http.StatusOK},
		// {"PUT", "/shutodwn", false, nil, http.StatusNotFound},
		// {"POST", "/shutodwn", false, nil, http.StatusNotFound},
		// {"DELETE", "/shutodwn", false, nil, http.StatusNotFound},
	}

	client := &http.Client{}
	server := httptest.NewServer(NewServerRoutes())

	for i, c := range cases {
		req, err := http.NewRequest(c.method, server.URL+c.endpoint, bytes.NewReader(c.body))
		if err != nil {
			t.Errorf("case %d error creating request: %s", i, err.Error())
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("case %d error performing request: %s", i, err.Error())
			continue
		}

		if res.StatusCode != c.resStatus {
			t.Errorf("case %d: %s - %s status code mismatch. expected: %d, got: %d", i, c.method, c.endpoint, c.resStatus, res.StatusCode)
			continue
		}

		if c.expectBody {
			env := &struct {
				Meta       map[string]interface{}
				Data       interface{}
				Pagination map[string]interface{}
			}{}

			if err := json.NewDecoder(res.Body).Decode(env); err != nil {
				t.Errorf("case %d: %s - %s error unmarshaling json envelope: %s", i, c.method, c.endpoint, err.Error())
				continue
			}

			if env.Meta == nil {
				t.Errorf("case %d: %s - %s doesn't have a meta field", i, c.method, c.endpoint)
				continue
			}
			if env.Data == nil {
				t.Errorf("case %d: %s - %s doesn't have a data field", i, c.method, c.endpoint)
				continue
			}
		}
	}
}
