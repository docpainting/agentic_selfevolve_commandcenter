# Neo4j Setup and Configuration

## Installation

### 1. Download Neo4j Community Edition

Visit: https://neo4j.com/deployment-center/

Download: **Neo4j Community Edition** (latest version)

### 2. Install Neo4j

```bash
# Extract and install
tar -xzf neo4j-community-*-unix.tar.gz
sudo mv neo4j-community-* /opt/neo4j
```

### 3. Configure Neo4j

**Location**: `/opt/neo4j/conf/neo4j.conf`

**Key Settings**:

```properties
# Network connector configuration
server.bolt.enabled=true
server.bolt.listen_address=0.0.0.0:7687

# HTTP Connector
server.http.enabled=true
server.http.listen_address=0.0.0.0:7474

# HTTPS Connector  
server.https.enabled=false

# Memory Settings (adjust based on your system)
server.memory.heap.initial_size=512m
server.memory.heap.max_size=2G
server.memory.pagecache.size=512m

# APOC Plugin Configuration
dbms.security.procedures.unrestricted=apoc.*
dbms.security.procedures.allowlist=apoc.*
```

### 4. Install APOC Plugin

**APOC** (Awesome Procedures On Cypher) - Essential for advanced graph operations

```bash
# Download APOC
cd /opt/neo4j/plugins
sudo wget https://github.com/neo4j/apoc/releases/download/5.15.0/apoc-5.15.0-core.jar

# Verify
ls -la /opt/neo4j/plugins/apoc-*.jar
```

**APOC Configuration** (add to neo4j.conf):

```properties
# Enable APOC
dbms.security.procedures.unrestricted=apoc.*
dbms.security.procedures.allowlist=apoc.coll.*,apoc.load.*,apoc.path.*,apoc.algo.*

# Allow APOC to access external URLs
apoc.import.file.enabled=true
apoc.export.file.enabled=true
```

---

## First Startup and Password Configuration

### 1. Start Neo4j

```bash
sudo /opt/neo4j/bin/neo4j start
```

### 2. Access Neo4j Browser

Open: http://localhost:7474

**Default Credentials**:
- Username: `neo4j`
- Password: `neo4j`

### 3. Change Password

**IMPORTANT**: On first login, you MUST change the password.

**New Password**: `BumBleBtuna1011*`

```cypher
// Run this in Neo4j Browser
ALTER USER neo4j SET PASSWORD 'BumBleBtuna1011*';
```

### 4. Verify APOC Installation

```cypher
// Check APOC procedures
CALL apoc.help("apoc");

// Should return list of APOC procedures
```

---

## Environment Configuration

### For Python (LightRAG)

**File**: `/home/ubuntu/agent-workspace/.env`

```bash
# Neo4j Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=BumBleBtuna1011*

# Ollama Configuration
OLLAMA_BASE_URL=http://localhost:11434

# LLM Model
LLM_MODEL=gemma3:27b

# Embedding Model
EMBEDDING_MODEL=nomic-embed-text:v1.5
EMBEDDING_DIM=768
```

### For Go Backend

**File**: `/home/ubuntu/agent-workspace/backend/.env`

```bash
# Neo4j Configuration
NEO4J_URI=bolt://localhost:7687
NEO4J_USERNAME=neo4j
NEO4J_PASSWORD=BumBleBtuna1011*

# Ollama Configuration
OLLAMA_BASE_URL=http://localhost:11434

# LLM Configuration
LLM_MODEL=gemma3:27b
LLM_TEMPERATURE=0.7

# Embedding Configuration
EMBEDDING_MODEL=nomic-embed-text:v1.5
EMBEDDING_DIM=768

# Vector DB Configuration
VECTOR_DB_TYPE=chromem
VECTOR_DB_PATH=./data/chromem

# KV Store Configuration
KV_STORE_TYPE=boltdb
KV_STORE_PATH=./data/boltdb
```

---

## Go Neo4j Driver Configuration

### File: `backend/internal/lightrag/client.go`

```go
package lightrag

import (
    "context"
    "os"
    
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Config struct {
    Neo4jURI      string
    Neo4jUsername string
    Neo4jPassword string
    OllamaBaseURL string
    LLMModel      string
    EmbeddingModel string
    EmbeddingDim  int
}

func LoadConfig() (*Config, error) {
    return &Config{
        Neo4jURI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
        Neo4jUsername: getEnv("NEO4J_USERNAME", "neo4j"),
        Neo4jPassword: getEnv("NEO4J_PASSWORD", "BumBleBtuna1011*"),
        OllamaBaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
        LLMModel:      getEnv("LLM_MODEL", "gemma3:27b"),
        EmbeddingModel: getEnv("EMBEDDING_MODEL", "nomic-embed-text:v1.5"),
        EmbeddingDim:  768,
    }, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

// Create Neo4j driver
func NewNeo4jDriver(config *Config) (neo4j.DriverWithContext, error) {
    driver, err := neo4j.NewDriverWithContext(
        config.Neo4jURI,
        neo4j.BasicAuth(config.Neo4jUsername, config.Neo4jPassword, ""),
    )
    if err != nil {
        return nil, err
    }
    
    // Verify connectivity
    ctx := context.Background()
    err = driver.VerifyConnectivity(ctx)
    if err != nil {
        return nil, err
    }
    
    return driver, nil
}
```

---

## Python LightRAG Configuration

### File: `backend/mcp_servers/dynamic_thinking/config.py`

```python
import os
from lightrag import LightRAG, QueryParam
from lightrag.llm import ollama_model_complete, ollama_embedding
from lightrag.kg.neo4j_impl import Neo4JStorage

def create_lightrag_instance():
    """Create LightRAG instance with Neo4j and Ollama."""
    
    # Configuration
    neo4j_uri = os.getenv('NEO4J_URI', 'bolt://localhost:7687')
    neo4j_username = os.getenv('NEO4J_USERNAME', 'neo4j')
    neo4j_password = os.getenv('NEO4J_PASSWORD', 'BumBleBtuna1011*')
    ollama_base_url = os.getenv('OLLAMA_BASE_URL', 'http://localhost:11434')
    llm_model = os.getenv('LLM_MODEL', 'gemma3:27b')
    embedding_model = os.getenv('EMBEDDING_MODEL', 'nomic-embed-text:v1.5')
    
    # Create LightRAG instance
    rag = LightRAG(
        working_dir="./lightrag_cache",
        
        # LLM Configuration
        llm_model_func=ollama_model_complete,
        llm_model_name=llm_model,
        llm_model_kwargs={
            "host": ollama_base_url,
            "options": {"temperature": 0.7}
        },
        
        # Embedding Configuration
        embedding_func=ollama_embedding,
        embedding_model=embedding_model,
        embedding_dim=768,
        embedding_kwargs={
            "host": ollama_base_url
        },
        
        # Graph Storage (Neo4j)
        graph_storage="Neo4JStorage",
        graph_storage_kwargs={
            "uri": neo4j_uri,
            "username": neo4j_username,
            "password": neo4j_password
        },
        
        # Vector Storage (ChromaDB)
        vector_storage="ChromaVectorDBStorage",
        
        # KV Storage (JSON)
        kv_storage="JsonKVStorage"
    )
    
    return rag
```

---

## Verification Steps

### 1. Check Neo4j is Running

```bash
sudo /opt/neo4j/bin/neo4j status
```

### 2. Test Connection from Python

```python
from neo4j import GraphDatabase

driver = GraphDatabase.driver(
    "bolt://localhost:7687",
    auth=("neo4j", "BumBleBtuna1011*")
)

with driver.session() as session:
    result = session.run("RETURN 1 AS num")
    print(result.single()["num"])  # Should print: 1

driver.close()
```

### 3. Test Connection from Go

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
    driver, err := neo4j.NewDriverWithContext(
        "bolt://localhost:7687",
        neo4j.BasicAuth("neo4j", "BumBleBtuna1011*", ""),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer driver.Close(context.Background())
    
    ctx := context.Background()
    session := driver.NewSession(ctx, neo4j.SessionConfig{})
    defer session.Close(ctx)
    
    result, err := session.Run(ctx, "RETURN 1 AS num", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    if result.Next(ctx) {
        fmt.Println(result.Record().Values[0])  // Should print: 1
    }
}
```

### 4. Test APOC

```cypher
// Test APOC path finding
MATCH (a:Node {name: 'A'}), (b:Node {name: 'B'})
CALL apoc.path.expandConfig(a, {
    relationshipFilter: "CONNECTED_TO>",
    endNodes: [b],
    maxLevel: 5
})
YIELD path
RETURN path;
```

---

## Common Issues

### Issue 1: "Connection refused"

**Solution**: Ensure Neo4j is running
```bash
sudo /opt/neo4j/bin/neo4j start
```

### Issue 2: "Authentication failed"

**Solution**: Verify password is correct
```cypher
ALTER USER neo4j SET PASSWORD 'BumBleBtuna1011*';
```

### Issue 3: "APOC procedures not found"

**Solution**: 
1. Check APOC jar is in plugins directory
2. Verify neo4j.conf has APOC configuration
3. Restart Neo4j

```bash
sudo /opt/neo4j/bin/neo4j restart
```

### Issue 4: "Out of memory"

**Solution**: Increase heap size in neo4j.conf
```properties
server.memory.heap.max_size=4G
```

---

## Summary

### Credentials

- **Username**: `neo4j`
- **Password**: `BumBleBtuna1011*`
- **Bolt URI**: `bolt://localhost:7687`
- **HTTP URI**: `http://localhost:7474`

### Required Components

✅ Neo4j Community Edition  
✅ APOC Plugin (5.15.0 or later)  
✅ Go Neo4j Driver (v5.28.0)  
✅ Python neo4j package (6.0.2)  
✅ LightRAG with Neo4j support  

### Configuration Files

- `/opt/neo4j/conf/neo4j.conf` - Neo4j configuration
- `/home/ubuntu/agent-workspace/.env` - Environment variables
- `/home/ubuntu/agent-workspace/backend/.env` - Go backend config

### Next Steps

1. Install Neo4j Community Edition
2. Configure neo4j.conf
3. Install APOC plugin
4. Start Neo4j
5. Change password to `BumBleBtuna1011*`
6. Verify connections from Python and Go
7. Test APOC procedures

**Neo4j is ready for LightRAG integration!** 🎯

