package migrations

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

const LockPrefix = "passnow_schema_migrations"

var filePattern = regexp.MustCompile("^[0-9]+_.+\\.up\\.sql$")

type Migration struct {
	Name string
	Path string
	SQL string
	Checksum string
}

func EnsureTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations (name VARCHAR(255) NOT NULL PRIMARY KEY, checksum CHAR(64) NOT NULL, applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci")
	return err
}

func RunUp(ctx context.Context, db *sql.DB, dir, lockName string) error {
	if lockName == "" { lockName = LockPrefix }
	if err := EnsureTable(ctx, db); err != nil { return err }
	if err := acquireLock(ctx, db, lockName); err != nil { return err }
	defer releaseLock(db, lockName)

	items, err := Load(dir)
	if err != nil { return err }
	for _, m := range items {
		var checksum string
		err := db.QueryRowContext(ctx, "SELECT checksum FROM schema_migrations WHERE name = ?", m.Name).Scan(&checksum)
		if err == nil {
			if checksum != m.Checksum { return fmt.Errorf("%s was already applied but its checksum changed; never edit an applied migration", m.Name) }
			continue
		}
		if err != sql.ErrNoRows { return err }

		tx, err := db.BeginTx(ctx, nil)
		if err != nil { return err }
		for _, statement := range SplitSQL(m.SQL) {
			if strings.TrimSpace(statement) == "" { continue }
			if _, err := tx.ExecContext(ctx, statement); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("%s: %w", m.Name, err)
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (name, checksum, applied_at) VALUES (?, ?, NOW())", m.Name, m.Checksum); err != nil {
			_ = tx.Rollback(); return err
		}
		if err := tx.Commit(); err != nil { return err }
	}
	return nil
}

func Status(ctx context.Context, db *sql.DB, dir string) ([]string, error) {
	if err := EnsureTable(ctx, db); err != nil { return nil, err }
	items, err := Load(dir)
	if err != nil { return nil, err }
	rows, err := db.QueryContext(ctx, "SELECT name, checksum FROM schema_migrations")
	if err != nil { return nil, err }
	defer rows.Close()
	applied := map[string]string{}
	for rows.Next() { var n,c string; if err:=rows.Scan(&n,&c); err!=nil{return nil,err}; applied[n]=c }
	if err:=rows.Err(); err!=nil{return nil,err}
	out:=make([]string,0,len(items))
	for _, m:=range items {
		state:="pending"
		if checksum,ok:=applied[m.Name];ok { state="applied"; if checksum!=m.Checksum { state="checksum-mismatch" } }
		out=append(out, fmt.Sprintf("%-20s %s", state,m.Name))
	}
	return out,nil
}

func Load(dir string) ([]Migration,error) {
	entries,err:=os.ReadDir(dir); if err!=nil{return nil,err}
	items:=make([]Migration,0)
	for _,entry:=range entries {
		if entry.IsDir() || !filePattern.MatchString(entry.Name()) { continue }
		path:=filepath.Join(dir,entry.Name()); body,err:=os.ReadFile(path); if err!=nil{return nil,err}
		sum:=sha256.Sum256(body)
		items=append(items,Migration{Name:entry.Name(),Path:path,SQL:string(body),Checksum:fmt.Sprintf("%x",sum[:])})
	}
	sort.Slice(items,func(i,j int)bool{return items[i].Name<items[j].Name})
	if len(items)==0{return nil,fmt.Errorf("no *.up.sql migrations found in %s",dir)}
	return items,nil
}

func acquireLock(ctx context.Context, db *sql.DB, name string) error {
	var ok int
	if err:=db.QueryRowContext(ctx,"SELECT GET_LOCK(?, 30)",name).Scan(&ok);err!=nil{return err}
	if ok!=1{return fmt.Errorf("could not acquire migration lock")}
	return nil
}
func releaseLock(db *sql.DB,name string){_,_=db.Exec("SELECT RELEASE_LOCK(?)",name)}

func SplitSQL(input string) []string {
	var out []string; var b strings.Builder; var quote rune; inLine:=false; inBlock:=false
	runes:=[]rune(input)
	for i,r:=range runes {
		next:=rune(0);if i+1<len(runes){next=runes[i+1]}
		if inLine { b.WriteRune(r);if r=='\n'{inLine=false};continue }
		if inBlock { b.WriteRune(r);if r=='/'&&i>0&&runes[i-1]=='*'{inBlock=false};continue }
		if quote==0 {
			if r=='-'&&next=='-' {inLine=true;b.WriteRune(r);continue}
			if r=='#' {inLine=true;b.WriteRune(r);continue}
			if r=='/'&&next=='*' {inBlock=true;b.WriteRune(r);continue}
			if r=='\''||r=='"'||r=='\x60' {quote=r;b.WriteRune(r);continue}
			if r==';' {out=append(out,b.String());b.Reset();continue}
			b.WriteRune(r);continue
		}
		b.WriteRune(r);if r==quote&&(i==0||runes[i-1]!='\\'){quote=0}
	}
	if strings.TrimSpace(b.String())!=""{out=append(out,b.String())}
	return out
}
