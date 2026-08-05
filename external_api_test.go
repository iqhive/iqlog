package iqlog_test

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/iqhive/iqlog"
)

func TestPublicAPIIsNameable(t *testing.T) {
	logger, err := iqlog.New(iqlog.Config{})
	if err != nil {
		t.Fatal(err)
	}
	var event *iqlog.Event = logger.InfoEvent()

	logger.Info("message")
	logger.Printf("value=%d", 1)
	logger.Println("value", 2)
	event.Str("key", "value").Msg("event")
	logger.Log(iqlog.LevelInfo, "message")
	logger.LogContext(context.Background(), iqlog.LevelInfo, "message")
}

func TestExportedNamedTypesAreIntentional(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	dir := filepath.Dir(file)
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, dir, func(info os.FileInfo) bool {
		return strings.HasSuffix(info.Name(), ".go") && !strings.HasSuffix(info.Name(), "_test.go")
	}, 0)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, file := range packages["iqlog"].Files {
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.TYPE {
				continue
			}
			for _, spec := range gen.Specs {
				name := spec.(*ast.TypeSpec).Name.Name
				if ast.IsExported(name) {
					names = append(names, name)
				}
			}
		}
	}
	sort.Strings(names)
	want := []string{"Config", "Event", "Format", "Level", "Logger", "OverflowPolicy", "WriterMode"}
	if len(names) != len(want) {
		t.Fatalf("exported types = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("exported types = %v, want %v", names, want)
		}
	}
}
