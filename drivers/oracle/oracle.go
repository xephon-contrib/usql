package oracle

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	// DRIVER: ora
	_ "gopkg.in/rana/ora.v4"

	"github.com/knq/usql/drivers"
)

var allCapsRE = regexp.MustCompile(`^[A-Z][A-Z0-9_]+$`)
var endRE = regexp.MustCompile(`;?\s*$`)

func init() {
	drivers.Register("ora", drivers.Driver{
		V: func(db drivers.DB) (string, error) {
			var ver string
			err := db.QueryRow(`SELECT version FROM V$INSTANCE`).Scan(&ver)
			if err != nil {
				return ver, err
			}
			return "Oracle " + ver, nil
		},
		U: func(db drivers.DB) (string, error) {
			var user string
			err := db.QueryRow(`select user from dual`).Scan(&user)
			return user, err
		},
		ChPw: func(db drivers.DB, user, new, _ string) error {
			_, err := db.Exec(`alter user ` + user + ` IDENTIFIED BY ` + new)
			return err
		},
		E: func(err error) (string, string) {
			code, msg := "", err.Error()

			if e, ok := err.(interface {
				Code() int
			}); ok {
				code = fmt.Sprintf("ORA-%05d", e.Code())
			}

			if i := strings.LastIndex(msg, "ORA-"); i != -1 {
				msg = msg[i:]
				j := strings.Index(msg, ":")
				if j != -1 {
					msg = msg[j+1:]
					if code == "" {
						code = msg[i:j]
					}
				}
			}

			return code, strings.TrimSpace(msg)
		},
		PwErr: func(err error) bool {
			if e, ok := err.(interface {
				Code() int
			}); ok {
				return e.Code() == 1017
			}
			return false
		},
		Cols: func(rows *sql.Rows) ([]string, error) {
			cols, err := rows.Columns()
			if err != nil {
				return nil, err
			}

			for i, c := range cols {
				if allCapsRE.MatchString(c) {
					cols[i] = strings.ToLower(c)
				}
			}

			return cols, nil
		},
		P: func(prefix string, sqlstr string) (string, string, bool, error) {
			sqlstr = endRE.ReplaceAllString(sqlstr, "")
			typ, q := drivers.QueryExecType(prefix, sqlstr)
			return typ, sqlstr, q, nil
		},
	})
}
