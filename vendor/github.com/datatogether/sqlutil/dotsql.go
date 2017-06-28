// utils for working with dotsql structs
package sqlutil

import (
	"github.com/gchaincl/dotsql"
	"strings"
)

// commandsWithPrefix returns all commands who's name matches a given prefix
func commandsWithPrefix(ds *dotsql.DotSql, prefix string) (cmds []string) {
	for name, _ := range ds.QueryMap() {
		if strings.HasPrefix(name, prefix) {
			cmds = append(cmds, name)
		}
	}
	return
}
