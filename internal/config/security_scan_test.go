package config

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

var hardcodedSnowflakePattern = regexp.MustCompile(`^[0-9]{17,20}$`)

func TestProductionSourceContainsNoHardcodedWebhookOrOperatorID(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	for _, directory := range []string{"internal", "cmd"} {
		err := filepath.WalkDir(filepath.Join(root, directory), func(path string, entry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Errorf("parse %s: %v", path, err)
				return nil
			}
			ast.Inspect(file, func(node ast.Node) bool {
				switch typed := node.(type) {
				case *ast.BasicLit:
					if value, ok := stringLiteral(typed); ok && containsDiscordWebhook(value) {
						t.Errorf("hardcoded Discord webhook in %s", path)
					}
				case *ast.ValueSpec:
					for index, name := range typed.Names {
						if index < len(typed.Values) {
							checkPrivilegedLiteral(t, path, name.Name, typed.Values[index])
						}
					}
				case *ast.AssignStmt:
					for index, left := range typed.Lhs {
						if index < len(typed.Rhs) {
							checkPrivilegedLiteral(t, path, expressionName(left), typed.Rhs[index])
						}
					}
				case *ast.KeyValueExpr:
					checkPrivilegedLiteral(t, path, expressionName(typed.Key), typed.Value)
				case *ast.BinaryExpr:
					checkPrivilegedLiteral(t, path, expressionName(typed.X), typed.Y)
					checkPrivilegedLiteral(t, path, expressionName(typed.Y), typed.X)
				}
				return true
			})
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", directory, err)
		}
	}
}

func checkPrivilegedLiteral(t *testing.T, path string, name string, expression ast.Expr) {
	t.Helper()
	if !isPrivilegedIdentifier(name) {
		return
	}
	literal, ok := expression.(*ast.BasicLit)
	if !ok {
		return
	}
	value, ok := stringLiteral(literal)
	if ok && hardcodedSnowflakePattern.MatchString(value) {
		t.Errorf("hardcoded privileged Discord ID assigned to %s in %s", name, path)
	}
}

func isPrivilegedIdentifier(name string) bool {
	name = strings.ToLower(strings.ReplaceAll(name, "_", ""))
	return strings.Contains(name, "ownerid") || strings.Contains(name, "adminid") || strings.Contains(name, "operatorid")
}

func expressionName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	default:
		return ""
	}
}

func stringLiteral(literal *ast.BasicLit) (string, bool) {
	if literal == nil || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func containsDiscordWebhook(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "discord.com/api/webhooks/") ||
		strings.Contains(value, "discordapp.com/api/webhooks/")
}
