package main

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// CodeMirror mirrors code structure into Neo4j
type CodeMirror struct {
	driver  neo4j.DriverWithContext
	project string
}

// NewCodeMirror creates a new code mirror
func NewCodeMirror(uri, username, password, projectName string) (*CodeMirror, error) {
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return nil, err
	}

	return &CodeMirror{
		driver:  driver,
		project: projectName,
	}, nil
}

// Close closes the driver
func (cm *CodeMirror) Close() error {
	return cm.driver.Close(context.Background())
}

// MirrorDirectory mirrors an entire directory
func (cm *CodeMirror) MirrorDirectory(ctx context.Context, rootPath string) error {
	// Create project node
	if err := cm.createProject(ctx); err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Walk directory tree
	return filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip vendor and node_modules
		if strings.Contains(path, "/vendor/") || strings.Contains(path, "/node_modules/") {
			return nil
		}

		log.Printf("Mirroring file: %s", path)
		return cm.MirrorFile(ctx, path)
	})
}

// MirrorFile mirrors a single file
func (cm *CodeMirror) MirrorFile(ctx context.Context, filePath string) error {
	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse Go file
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		log.Printf("Warning: failed to parse %s: %v", filePath, err)
		// Continue with basic file mirroring
		return cm.mirrorFileBasic(ctx, filePath, string(content))
	}

	// Mirror file node
	if err := cm.mirrorFileNode(ctx, filePath, string(content), fset.File(node.Pos()).LineCount()); err != nil {
		return err
	}

	// Mirror imports
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}
		if err := cm.mirrorImport(ctx, filePath, importPath, alias); err != nil {
			log.Printf("Warning: failed to mirror import: %v", err)
		}
	}

	// Mirror declarations
	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if err := cm.mirrorFunction(ctx, filePath, d, fset); err != nil {
				log.Printf("Warning: failed to mirror function: %v", err)
			}
		case *ast.GenDecl:
			if err := cm.mirrorGenDecl(ctx, filePath, d, fset); err != nil {
				log.Printf("Warning: failed to mirror declaration: %v", err)
			}
		}
	}

	return nil
}

// createProject creates the project node
func (cm *CodeMirror) createProject(ctx context.Context) error {
	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MERGE (p:Project {name: $name})
			ON CREATE SET 
				p.description = $description,
				p.language = $language,
				p.version = $version,
				p.created_at = datetime()
			ON MATCH SET
				p.updated_at = datetime()
		`
		params := map[string]interface{}{
			"name":        cm.project,
			"description": "AI Agent Workspace",
			"language":    "Go",
			"version":     "1.0.0",
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// mirrorFileNode creates/updates file node
func (cm *CodeMirror) mirrorFileNode(ctx context.Context, path, content string, lines int) error {
	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Create file node
		fileQuery := `
			MERGE (f:File {path: $path})
			ON CREATE SET
				f.name = $name,
				f.extension = $extension,
				f.language = $language,
				f.content = $content,
				f.lines = $lines,
				f.created_at = datetime()
			ON MATCH SET
				f.content = $content,
				f.lines = $lines,
				f.updated_at = datetime()
		`
		fileParams := map[string]interface{}{
			"path":      path,
			"name":      filepath.Base(path),
			"extension": filepath.Ext(path),
			"language":  "Go",
			"content":   content,
			"lines":     lines,
		}
		if _, err := tx.Run(ctx, fileQuery, fileParams); err != nil {
			return nil, err
		}

		// Link to project
		linkQuery := `
			MATCH (p:Project {name: $project})
			MATCH (f:File {path: $path})
			MERGE (p)-[:CONTAINS_FILE]->(f)
		`
		linkParams := map[string]interface{}{
			"project": cm.project,
			"path":    path,
		}
		_, err := tx.Run(ctx, linkQuery, linkParams)
		return nil, err
	})

	return err
}

// mirrorFileBasic creates basic file node without parsing
func (cm *CodeMirror) mirrorFileBasic(ctx context.Context, path, content string) error {
	lines := strings.Count(content, "\n") + 1
	return cm.mirrorFileNode(ctx, path, content, lines)
}

// mirrorImport creates import node
func (cm *CodeMirror) mirrorImport(ctx context.Context, filePath, module, alias string) error {
	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (f:File {path: $file_path})
			MERGE (i:Import {module: $module, file_path: $file_path})
			ON CREATE SET
				i.alias = $alias,
				i.created_at = datetime()
			MERGE (f)-[:HAS_IMPORT]->(i)
		`
		params := map[string]interface{}{
			"file_path": filePath,
			"module":    module,
			"alias":     alias,
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// mirrorFunction creates function node
func (cm *CodeMirror) mirrorFunction(ctx context.Context, filePath string, fn *ast.FuncDecl, fset *token.FileSet) error {
	// Build signature
	signature := cm.buildFunctionSignature(fn)

	// Extract documentation
	doc := ""
	if fn.Doc != nil {
		doc = fn.Doc.Text()
	}

	// Calculate lines
	startLine := fset.Position(fn.Pos()).Line
	endLine := fset.Position(fn.End()).Line
	lines := endLine - startLine + 1

	// Check if exported
	isExported := fn.Name.IsExported()

	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// Create function node
		funcQuery := `
			MERGE (fn:Function {signature: $signature})
			ON CREATE SET
				fn.name = $name,
				fn.documentation = $documentation,
				fn.lines = $lines,
				fn.is_exported = $is_exported,
				fn.created_at = datetime()
			ON MATCH SET
				fn.documentation = $documentation,
				fn.lines = $lines,
				fn.updated_at = datetime()
		`
		funcParams := map[string]interface{}{
			"signature":     signature,
			"name":          fn.Name.Name,
			"documentation": doc,
			"lines":         lines,
			"is_exported":   isExported,
		}
		if _, err := tx.Run(ctx, funcQuery, funcParams); err != nil {
			return nil, err
		}

		// Link to file
		linkQuery := `
			MATCH (f:File {path: $file_path})
			MATCH (fn:Function {signature: $signature})
			MERGE (f)-[:DEFINES_FUNCTION]->(fn)
		`
		linkParams := map[string]interface{}{
			"file_path": filePath,
			"signature": signature,
		}
		_, err := tx.Run(ctx, linkQuery, linkParams)
		return nil, err
	})

	return err
}

// mirrorGenDecl mirrors general declarations (types, vars, consts)
func (cm *CodeMirror) mirrorGenDecl(ctx context.Context, filePath string, decl *ast.GenDecl, fset *token.FileSet) error {
	for _, spec := range decl.Specs {
		switch s := spec.(type) {
		case *ast.TypeSpec:
			// Mirror type (struct, interface, etc.)
			if err := cm.mirrorType(ctx, filePath, s, decl.Doc); err != nil {
				return err
			}
		case *ast.ValueSpec:
			// Mirror variable or constant
			for _, name := range s.Names {
				if err := cm.mirrorVariable(ctx, filePath, name.Name, decl.Tok.String()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// mirrorType creates type/class node
func (cm *CodeMirror) mirrorType(ctx context.Context, filePath string, spec *ast.TypeSpec, doc *ast.CommentGroup) error {
	fqn := fmt.Sprintf("%s.%s", filepath.Base(filepath.Dir(filePath)), spec.Name.Name)
	
	docText := ""
	if doc != nil {
		docText = doc.Text()
	}

	typeKind := "type"
	switch spec.Type.(type) {
	case *ast.StructType:
		typeKind = "struct"
	case *ast.InterfaceType:
		typeKind = "interface"
	}

	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (f:File {path: $file_path})
			MERGE (c:Class {fully_qualified_name: $fqn})
			ON CREATE SET
				c.name = $name,
				c.type = $type,
				c.documentation = $documentation,
				c.is_exported = $is_exported,
				c.created_at = datetime()
			ON MATCH SET
				c.documentation = $documentation,
				c.updated_at = datetime()
			MERGE (f)-[:DEFINES_CLASS]->(c)
		`
		params := map[string]interface{}{
			"file_path":     filePath,
			"fqn":           fqn,
			"name":          spec.Name.Name,
			"type":          typeKind,
			"documentation": docText,
			"is_exported":   spec.Name.IsExported(),
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// mirrorVariable creates variable node
func (cm *CodeMirror) mirrorVariable(ctx context.Context, filePath, name, scope string) error {
	session := cm.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
			MATCH (f:File {path: $file_path})
			MERGE (v:Variable {name: $name, scope: $scope, file_path: $file_path})
			ON CREATE SET
				v.created_at = datetime()
			MERGE (f)-[:DEFINES_VARIABLE]->(v)
		`
		params := map[string]interface{}{
			"file_path": filePath,
			"name":      name,
			"scope":     scope,
		}
		_, err := tx.Run(ctx, query, params)
		return nil, err
	})

	return err
}

// buildFunctionSignature builds a function signature string
func (cm *CodeMirror) buildFunctionSignature(fn *ast.FuncDecl) string {
	sig := fn.Name.Name + "("

	if fn.Type.Params != nil {
		params := []string{}
		for _, param := range fn.Type.Params.List {
			paramType := cm.exprToString(param.Type)
			for _, name := range param.Names {
				params = append(params, name.Name+" "+paramType)
			}
		}
		sig += strings.Join(params, ", ")
	}

	sig += ")"

	if fn.Type.Results != nil && len(fn.Type.Results.List) > 0 {
		results := []string{}
		for _, result := range fn.Type.Results.List {
			results = append(results, cm.exprToString(result.Type))
		}
		if len(results) == 1 {
			sig += " " + results[0]
		} else {
			sig += " (" + strings.Join(results, ", ") + ")"
		}
	}

	return sig
}

// exprToString converts an expression to string
func (cm *CodeMirror) exprToString(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.StarExpr:
		return "*" + cm.exprToString(e.X)
	case *ast.SelectorExpr:
		return cm.exprToString(e.X) + "." + e.Sel.Name
	case *ast.ArrayType:
		return "[]" + cm.exprToString(e.Elt)
	default:
		return "interface{}"
	}
}

func main() {
	// Configuration
	neo4jURI := os.Getenv("NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "bolt://localhost:7687"
	}

	neo4jUser := os.Getenv("NEO4J_USERNAME")
	if neo4jUser == "" {
		neo4jUser = "neo4j"
	}

	neo4jPass := os.Getenv("NEO4J_PASSWORD")
	if neo4jPass == "" {
		log.Fatal("NEO4J_PASSWORD environment variable is required")
	}

	projectName := "agent-workspace"
	rootPath := "./backend"

	// Create mirror
	mirror, err := NewCodeMirror(neo4jURI, neo4jUser, neo4jPass, projectName)
	if err != nil {
		log.Fatalf("Failed to create code mirror: %v", err)
	}
	defer mirror.Close()

	// Mirror directory
	ctx := context.Background()
	log.Printf("Starting code mirror for project: %s", projectName)
	log.Printf("Root path: %s", rootPath)

	if err := mirror.MirrorDirectory(ctx, rootPath); err != nil {
		log.Fatalf("Failed to mirror directory: %v", err)
	}

	log.Println("Code mirroring completed successfully!")
}

