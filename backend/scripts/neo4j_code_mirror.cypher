// ============================================================================
// Neo4j Code Mirroring Schema for LLM Understanding
// ============================================================================
// This schema represents programming concepts optimally for LLM comprehension
// Uses MERGE to ensure idempotency and prevent duplicates
// ============================================================================

// ----------------------------------------------------------------------------
// 1. CONSTRAINTS & INDEXES (Run these first)
// ----------------------------------------------------------------------------

// Unique constraints
CREATE CONSTRAINT project_name_unique IF NOT EXISTS FOR (p:Project) REQUIRE p.name IS UNIQUE;
CREATE CONSTRAINT file_path_unique IF NOT EXISTS FOR (f:File) REQUIRE f.path IS UNIQUE;
CREATE CONSTRAINT function_signature_unique IF NOT EXISTS FOR (fn:Function) REQUIRE fn.signature IS UNIQUE;
CREATE CONSTRAINT class_fqn_unique IF NOT EXISTS FOR (c:Class) REQUIRE c.fully_qualified_name IS UNIQUE;
CREATE CONSTRAINT concept_name_unique IF NOT EXISTS FOR (con:Concept) REQUIRE con.name IS UNIQUE;
CREATE CONSTRAINT pattern_signature_unique IF NOT EXISTS FOR (pat:Pattern) REQUIRE pat.signature IS UNIQUE;

// Indexes for performance
CREATE INDEX file_language_idx IF NOT EXISTS FOR (f:File) ON (f.language);
CREATE INDEX function_name_idx IF NOT EXISTS FOR (fn:Function) ON (fn.name);
CREATE INDEX class_name_idx IF NOT EXISTS FOR (c:Class) ON (c.name);
CREATE INDEX concept_category_idx IF NOT EXISTS FOR (con:Concept) ON (con.category);
CREATE INDEX execution_timestamp_idx IF NOT EXISTS FOR (e:Execution) ON (e.timestamp);

// Full-text search indexes
CREATE FULLTEXT INDEX file_content_fulltext IF NOT EXISTS FOR (f:File) ON EACH [f.content];
CREATE FULLTEXT INDEX function_doc_fulltext IF NOT EXISTS FOR (fn:Function) ON EACH [fn.documentation];
CREATE FULLTEXT INDEX concept_description_fulltext IF NOT EXISTS FOR (con:Concept) ON EACH [con.description];

// ----------------------------------------------------------------------------
// 2. NODE TYPES (Core Entities)
// ----------------------------------------------------------------------------

// Project Node - Top-level container
// Properties: name, description, language, version, created_at, updated_at
// Example:
MERGE (p:Project {name: $project_name})
ON CREATE SET 
  p.description = $description,
  p.language = $language,
  p.version = $version,
  p.created_at = datetime(),
  p.updated_at = datetime()
ON MATCH SET
  p.updated_at = datetime();

// File Node - Source code files
// Properties: path, name, extension, language, content, lines, size, hash, created_at, updated_at
// Example:
MERGE (f:File {path: $file_path})
ON CREATE SET
  f.name = $file_name,
  f.extension = $extension,
  f.language = $language,
  f.content = $content,
  f.lines = $line_count,
  f.size = $size_bytes,
  f.hash = $content_hash,
  f.created_at = datetime(),
  f.updated_at = datetime()
ON MATCH SET
  f.content = $content,
  f.lines = $line_count,
  f.size = $size_bytes,
  f.hash = $content_hash,
  f.updated_at = datetime();

// Function Node - Functions/Methods
// Properties: name, signature, parameters, return_type, documentation, complexity, lines, is_async, is_exported
// Example:
MERGE (fn:Function {signature: $function_signature})
ON CREATE SET
  fn.name = $function_name,
  fn.parameters = $parameters,
  fn.return_type = $return_type,
  fn.documentation = $doc_string,
  fn.complexity = $cyclomatic_complexity,
  fn.lines = $line_count,
  fn.is_async = $is_async,
  fn.is_exported = $is_exported,
  fn.created_at = datetime()
ON MATCH SET
  fn.documentation = $doc_string,
  fn.complexity = $cyclomatic_complexity,
  fn.lines = $line_count,
  fn.updated_at = datetime();

// Class/Struct Node - Classes, Structs, Interfaces
// Properties: name, fully_qualified_name, type (class/struct/interface), documentation, is_exported
// Example:
MERGE (c:Class {fully_qualified_name: $fqn})
ON CREATE SET
  c.name = $class_name,
  c.type = $class_type,
  c.documentation = $doc_string,
  c.is_exported = $is_exported,
  c.created_at = datetime()
ON MATCH SET
  c.documentation = $doc_string,
  c.updated_at = datetime();

// Variable Node - Variables, Constants, Fields
// Properties: name, type, scope (global/local/field), value, is_constant, is_exported
// Example:
MERGE (v:Variable {name: $var_name, scope: $scope, file_path: $file_path})
ON CREATE SET
  v.type = $var_type,
  v.value = $initial_value,
  v.is_constant = $is_const,
  v.is_exported = $is_exported,
  v.created_at = datetime();

// Import Node - Dependencies/Imports
// Properties: module, alias, items (what's imported), source
// Example:
MERGE (i:Import {module: $module_name, file_path: $file_path})
ON CREATE SET
  i.alias = $alias,
  i.items = $imported_items,
  i.source = $import_source,
  i.created_at = datetime();

// Concept Node - Abstract concepts (patterns, algorithms, architectures)
// Properties: name, category, description, examples, related_terms
// Example:
MERGE (con:Concept {name: $concept_name})
ON CREATE SET
  con.category = $category,
  con.description = $description,
  con.examples = $examples,
  con.related_terms = $related_terms,
  con.created_at = datetime()
ON MATCH SET
  con.description = $description,
  con.examples = $examples,
  con.updated_at = datetime();

// Pattern Node - Design patterns, code patterns
// Properties: signature, pattern_type, description, use_cases, frequency
// Example:
MERGE (pat:Pattern {signature: $pattern_signature})
ON CREATE SET
  pat.pattern_type = $type,
  pat.description = $description,
  pat.use_cases = $use_cases,
  pat.frequency = 1,
  pat.created_at = datetime()
ON MATCH SET
  pat.frequency = pat.frequency + 1,
  pat.last_seen = datetime();

// Execution Node - Runtime executions
// Properties: execution_id, command, status, timestamp, duration, result
// Example:
CREATE (e:Execution {execution_id: $exec_id})
SET
  e.command = $command,
  e.status = $status,
  e.timestamp = datetime(),
  e.duration = $duration_ms,
  e.result = $result;

// Conversation Node - User-Agent conversations
// Properties: conversation_id, user_message, agent_response, timestamp, context
// Example:
CREATE (conv:Conversation {conversation_id: $conv_id})
SET
  conv.user_message = $user_msg,
  conv.agent_response = $agent_msg,
  conv.timestamp = datetime(),
  conv.context = $context;

// ----------------------------------------------------------------------------
// 3. RELATIONSHIP TYPES (Semantic Connections)
// ----------------------------------------------------------------------------

// File Relationships
MERGE (p:Project {name: $project_name})
MERGE (f:File {path: $file_path})
MERGE (p)-[:CONTAINS_FILE]->(f);

MERGE (f1:File {path: $file1_path})
MERGE (f2:File {path: $file2_path})
MERGE (f1)-[:IMPORTS]->(f2);

// Function Relationships
MERGE (f:File {path: $file_path})
MERGE (fn:Function {signature: $function_signature})
MERGE (f)-[:DEFINES_FUNCTION]->(fn);

MERGE (fn1:Function {signature: $caller_signature})
MERGE (fn2:Function {signature: $callee_signature})
MERGE (fn1)-[:CALLS {count: $call_count, line: $line_number}]->(fn2);

MERGE (fn:Function {signature: $function_signature})
MERGE (v:Variable {name: $var_name})
MERGE (fn)-[:USES_VARIABLE]->(v);

MERGE (fn:Function {signature: $function_signature})
MERGE (v:Variable {name: $var_name})
MERGE (fn)-[:MODIFIES_VARIABLE]->(v);

// Class Relationships
MERGE (f:File {path: $file_path})
MERGE (c:Class {fully_qualified_name: $fqn})
MERGE (f)-[:DEFINES_CLASS]->(c);

MERGE (c:Class {fully_qualified_name: $fqn})
MERGE (fn:Function {signature: $method_signature})
MERGE (c)-[:HAS_METHOD]->(fn);

MERGE (c:Class {fully_qualified_name: $fqn})
MERGE (v:Variable {name: $field_name})
MERGE (c)-[:HAS_FIELD]->(v);

MERGE (c1:Class {fully_qualified_name: $child_fqn})
MERGE (c2:Class {fully_qualified_name: $parent_fqn})
MERGE (c1)-[:INHERITS_FROM]->(c2);

MERGE (c1:Class {fully_qualified_name: $impl_fqn})
MERGE (c2:Class {fully_qualified_name: $interface_fqn})
MERGE (c1)-[:IMPLEMENTS]->(c2);

// Dependency Relationships
MERGE (f:File {path: $file_path})
MERGE (i:Import {module: $module_name})
MERGE (f)-[:HAS_IMPORT]->(i);

MERGE (fn:Function {signature: $function_signature})
MERGE (i:Import {module: $module_name})
MERGE (fn)-[:DEPENDS_ON]->(i);

// Concept Relationships
MERGE (fn:Function {signature: $function_signature})
MERGE (con:Concept {name: $concept_name})
MERGE (fn)-[:IMPLEMENTS_CONCEPT {confidence: $confidence}]->(con);

MERGE (pat:Pattern {signature: $pattern_signature})
MERGE (con:Concept {name: $concept_name})
MERGE (pat)-[:RELATES_TO]->(con);

MERGE (con1:Concept {name: $concept1})
MERGE (con2:Concept {name: $concept2})
MERGE (con1)-[:RELATED_TO {strength: $strength}]->(con2);

// Execution Relationships
MERGE (e:Execution {execution_id: $exec_id})
MERGE (fn:Function {signature: $function_signature})
MERGE (e)-[:EXECUTED]->(fn);

MERGE (e:Execution {execution_id: $exec_id})
MERGE (f:File {path: $file_path})
MERGE (e)-[:MODIFIED_FILE]->(f);

// Conversation Relationships
MERGE (conv:Conversation {conversation_id: $conv_id})
MERGE (f:File {path: $file_path})
MERGE (conv)-[:MENTIONS_FILE]->(f);

MERGE (conv:Conversation {conversation_id: $conv_id})
MERGE (con:Concept {name: $concept_name})
MERGE (conv)-[:DISCUSSES_CONCEPT]->(con);

MERGE (conv:Conversation {conversation_id: $conv_id})
MERGE (e:Execution {execution_id: $exec_id})
MERGE (conv)-[:TRIGGERED_EXECUTION]->(e);

// ----------------------------------------------------------------------------
// 4. COMPLETE EXAMPLE: Mirroring a Go Function
// ----------------------------------------------------------------------------

// Example: Mirror a Go HTTP handler function
MERGE (p:Project {name: "agent-workspace"})
ON CREATE SET 
  p.description = "AI Agent Workspace",
  p.language = "Go",
  p.version = "1.0.0",
  p.created_at = datetime();

MERGE (f:File {path: "/backend/internal/websocket/chat.go"})
ON CREATE SET
  f.name = "chat.go",
  f.extension = "go",
  f.language = "Go",
  f.lines = 150,
  f.created_at = datetime();

MERGE (fn:Function {signature: "HandleChatWebSocket(c *fiber.Ctx, agentController *agent.Controller)"})
ON CREATE SET
  fn.name = "HandleChatWebSocket",
  fn.parameters = ["c *fiber.Ctx", "agentController *agent.Controller"],
  fn.return_type = "error",
  fn.documentation = "Handles WebSocket connections for real-time chat with the agent",
  fn.complexity = 8,
  fn.lines = 45,
  fn.is_async = false,
  fn.is_exported = true,
  fn.created_at = datetime();

MERGE (p)-[:CONTAINS_FILE]->(f);
MERGE (f)-[:DEFINES_FUNCTION]->(fn);

// Link to concepts
MERGE (con:Concept {name: "WebSocket"})
ON CREATE SET
  con.category = "Communication Protocol",
  con.description = "Full-duplex communication channel over TCP",
  con.created_at = datetime();

MERGE (fn)-[:IMPLEMENTS_CONCEPT {confidence: 0.95}]->(con);

// Link to pattern
MERGE (pat:Pattern {signature: "WebSocket Handler Pattern"})
ON CREATE SET
  pat.pattern_type = "Communication",
  pat.description = "Upgrade HTTP connection to WebSocket and handle messages",
  pat.frequency = 1,
  pat.created_at = datetime();

MERGE (fn)-[:USES_PATTERN]->(pat);

// ----------------------------------------------------------------------------
// 5. QUERY EXAMPLES FOR LLM
// ----------------------------------------------------------------------------

// Find all functions that implement a concept
// MATCH (fn:Function)-[:IMPLEMENTS_CONCEPT]->(con:Concept {name: "Authentication"})
// RETURN fn.name, fn.signature, fn.documentation;

// Find call chain for a function
// MATCH path = (fn1:Function {name: "ExecuteCommand"})-[:CALLS*]->(fn2:Function)
// RETURN path;

// Find files related to a conversation
// MATCH (conv:Conversation)-[:MENTIONS_FILE]->(f:File)
// WHERE conv.user_message CONTAINS "authentication"
// RETURN f.path, f.language;

// Find patterns used in a project
// MATCH (p:Project {name: "agent-workspace"})-[:CONTAINS_FILE]->(f:File)-[:DEFINES_FUNCTION]->(fn:Function)-[:USES_PATTERN]->(pat:Pattern)
// RETURN pat.pattern_type, pat.description, count(fn) as usage_count
// ORDER BY usage_count DESC;

// Find concepts related to a file
// MATCH (f:File {path: "/backend/internal/agent/controller.go"})-[:DEFINES_FUNCTION]->(fn:Function)-[:IMPLEMENTS_CONCEPT]->(con:Concept)
// RETURN DISTINCT con.name, con.category, con.description;

