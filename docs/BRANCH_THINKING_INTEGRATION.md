# Branch-Thinking MCP Integration
## Multi-Branch Reasoning for Dynamic Thinking

## Overview

The **branch-thinking-mcp** repository provides a production-ready TypeScript MCP server for managing parallel branches of thought with semantic search, cross-references, and AI insights. This is **exactly** what we need for the multi-branch reasoning component of Dynamic Thinking!

---

## What branch-thinking-mcp Provides

### Core Features

1. **Branch Management**
   - Create, focus, and navigate multiple lines of thought
   - Session state management (INIT → BRANCH_CREATED → BRANCH_FOCUSED → THOUGHT_ADDED → ACTIVE)
   - Command safety validation (only allowed commands in each state)

2. **Semantic Search**
   - Embeddings using @xenova/transformers
   - Cosine similarity for finding related thoughts
   - LRU+TTL caching for performance (30min TTL, 1000 max)

3. **Cross-References**
   - Link related thoughts across branches
   - Typed relationships (supports, contradicts, extends, etc.)
   - Scored connections

4. **AI Insights**
   - Automatic summarization (distilbart-cnn-6-6)
   - Insight generation
   - Summary caching (5min TTL, 100 max)

5. **Visualization**
   - Knowledge graph with @dagrejs/graphlib
   - Node clustering (k-means/degree)
   - Centrality overlays (closeness, betweenness)
   - Edge bundling
   - Task overlays

6. **Task Management**
   - Extract actionable items from thoughts
   - Track status, assignee, due dates
   - Task visualization

---

## Integration with Dynamic Thinking

### Current Dynamic Thinking Architecture

```
Dynamic Thinking MCP Server (Python)
├── perceive (Enhanced perception)
├── reason (Multi-branch reasoning) ← NEEDS IMPLEMENTATION
├── act (Dynamic execution)
├── reflect (Learning)
├── query_memory (Semantic search)
└── evolve_prompt (Prompt evolution)
```

### Integration Strategy

**Option 1: Use branch-thinking as a Sub-MCP** (Recommended)

```
Dynamic Thinking MCP (Python)
    ↓
    Calls branch-thinking MCP (TypeScript)
    ↓
    Returns reasoning branches
```

**Benefits**:
- Reuse production-ready code
- TypeScript performance for graph operations
- Proven semantic search and caching
- No need to reimplement in Python

**How**:
```python
# In reason tool (Dynamic Thinking MCP)
import subprocess
import json

def reason(perception_data):
    # 1. Create reasoning branch
    result = subprocess.run([
        'node', '/path/to/branch-thinking-mcp/dist/index.js',
        'create-branch', 'reasoning-session-123'
    ], capture_output=True)
    
    branch_id = json.loads(result.stdout)['branchId']
    
    # 2. Add reasoning branches as thoughts
    for i, branch in enumerate(generate_branches(perception_data)):
        subprocess.run([
            'node', '/path/to/branch-thinking-mcp/dist/index.js',
            'add-thought', branch_id, branch['content'], 'analysis'
        ])
    
    # 3. Get semantic search for similar past reasoning
    result = subprocess.run([
        'node', '/path/to/branch-thinking-mcp/dist/index.js',
        'semantic-search', perception_data['query']
    ], capture_output=True)
    
    similar = json.loads(result.stdout)
    
    # 4. Get insights
    result = subprocess.run([
        'node', '/path/to/branch-thinking-mcp/dist/index.js',
        'insights', branch_id
    ], capture_output=True)
    
    insights = json.loads(result.stdout)
    
    # 5. Select best branch based on insights
    best_branch = select_best(insights)
    
    return {
        'branch_id': branch_id,
        'selected': best_branch,
        'insights': insights
    }
```

**Option 2: Reimplement in Python**

Reimplement the key features in Python:
- Use sentence-transformers for embeddings
- Use networkx for graph operations
- Use transformers for summarization

**Downside**: More work, need to maintain parallel implementation

---

## Key Components to Integrate

### 1. BranchManager Class

**Location**: `src/branchManager.ts`

**Key Methods**:
```typescript
class BranchManager {
  // Branch operations
  createBranch(name: string, description?: string): ThoughtBranch
  focusBranch(branchId: string): void
  addThought(branchId: string, content: string, type: string): ThoughtData
  
  // Semantic search
  semanticSearch(query: string, limit: number): ThoughtData[]
  
  // Cross-references
  linkThoughts(fromId: string, toId: string, type: CrossRefType, reason?: string): void
  getCrossReferences(branchId: string): CrossReference[]
  
  // AI insights
  generateInsights(branchId: string): Insight[]
  summarizeBranchThoughts(branchId: string): Promise<string>
  
  // Visualization
  visualizeBranch(branchId: string, options?: VisualizationOptions): VisualizationData
  
  // Task management
  extractTasks(branchId: string): TaskItem[]
  listTasks(branchId: string, filters?: TaskFilters): TaskItem[]
}
```

### 2. Caching Strategy

**Embedding Cache**:
```typescript
private embeddingCache = new LRUCache<string, number[]>({
  max: 1000,
  ttl: 1000 * 60 * 30  // 30 minutes
});
```

**Summary Cache**:
```typescript
private summaryCache = new LRUCache<string, string>({
  max: 100,
  ttl: 1000 * 60 * 5  // 5 minutes
});
```

**Benefits**:
- Fast repeated queries
- Memory efficient (LRU eviction)
- Fresh data (TTL expiration)

### 3. Session State Management

**States**:
```typescript
enum SessionState {
  INIT = 'INIT',
  BRANCH_CREATED = 'BRANCH_CREATED',
  BRANCH_FOCUSED = 'BRANCH_FOCUSED',
  THOUGHT_ADDED = 'THOUGHT_ADDED',
  ACTIVE = 'ACTIVE',
  RESET = 'RESET'
}
```

**Command Validation**:
```typescript
private allowedCommands: Record<SessionState, string[]> = {
  [SessionState.INIT]: ['create-branch', 'list'],
  [SessionState.BRANCH_CREATED]: ['focus', 'list', 'create-branch'],
  [SessionState.BRANCH_FOCUSED]: [
    'add-thought', 'insights', 'crossrefs', 'hub-thoughts',
    'semantic-search', 'link-thoughts', 'add-snippet',
    'snippet-search', 'summarize-branch', 'doc-thought',
    'extract-tasks', 'review-branch', 'visualize', 'ask'
  ],
  [SessionState.THOUGHT_ADDED]: [...],
  [SessionState.ACTIVE]: [...],
  [SessionState.RESET]: ['create-branch', 'list']
};
```

**Why This Matters**:
- Prevents invalid command sequences
- Ensures proper context
- Clear error messages

---

## Integration Implementation

### Phase 1: Setup (1 day)

1. Install branch-thinking-mcp in agent workspace
```bash
cd /home/ubuntu/agent-workspace/backend/mcp_servers
git clone https://github.com/ssdeanx/branch-thinking-mcp.git
cd branch-thinking-mcp
pnpm install
pnpm build
```

2. Test standalone
```bash
node dist/index.js create-branch "test"
node dist/index.js focus [branchId]
node dist/index.js add-thought [branchId] "Test thought" note
node dist/index.js insights [branchId]
```

### Phase 2: Python Wrapper (2 days)

Create Python wrapper in Dynamic Thinking MCP:

**File**: `backend/mcp_servers/dynamic_thinking/reasoning/branch_thinking_client.py`

```python
import subprocess
import json
from typing import List, Dict, Any, Optional
from pathlib import Path

class BranchThinkingClient:
    """Client for branch-thinking-mcp TypeScript server."""
    
    def __init__(self, mcp_path: str):
        self.mcp_path = Path(mcp_path)
        self.node_cmd = ['node', str(self.mcp_path / 'dist' / 'index.js')]
    
    def _run_command(self, *args) -> Dict[str, Any]:
        """Run branch-thinking command and return JSON result."""
        result = subprocess.run(
            self.node_cmd + list(args),
            capture_output=True,
            text=True,
            check=True
        )
        return json.loads(result.stdout)
    
    def create_branch(self, name: str, description: Optional[str] = None) -> Dict[str, Any]:
        """Create a new reasoning branch."""
        args = ['create-branch', name]
        if description:
            args.append(description)
        return self._run_command(*args)
    
    def focus_branch(self, branch_id: str) -> Dict[str, Any]:
        """Focus on a specific branch."""
        return self._run_command('focus', branch_id)
    
    def add_thought(self, branch_id: str, content: str, thought_type: str = 'analysis') -> Dict[str, Any]:
        """Add a thought to a branch."""
        return self._run_command('add-thought', branch_id, content, thought_type)
    
    def semantic_search(self, query: str, limit: int = 10) -> List[Dict[str, Any]]:
        """Search for similar thoughts."""
        result = self._run_command('semantic-search', query, str(limit))
        return result.get('thoughts', [])
    
    def get_insights(self, branch_id: str) -> List[Dict[str, Any]]:
        """Get AI-generated insights for a branch."""
        result = self._run_command('insights', branch_id)
        return result.get('insights', [])
    
    def summarize_branch(self, branch_id: str) -> str:
        """Get AI summary of branch."""
        result = self._run_command('summarize-branch', branch_id)
        return result.get('summary', '')
    
    def link_thoughts(self, from_id: str, to_id: str, link_type: str, reason: Optional[str] = None) -> Dict[str, Any]:
        """Link two thoughts with a relationship."""
        args = ['link-thoughts', from_id, to_id, link_type]
        if reason:
            args.append(reason)
        return self._run_command(*args)
    
    def visualize(self, branch_id: str) -> Dict[str, Any]:
        """Get visualization data for a branch."""
        return self._run_command('visualize', branch_id)
    
    def extract_tasks(self, branch_id: str) -> List[Dict[str, Any]]:
        """Extract actionable tasks from thoughts."""
        result = self._run_command('extract-tasks', branch_id)
        return result.get('tasks', [])
```

### Phase 3: Integration with Reason Tool (3 days)

Update `backend/mcp_servers/dynamic_thinking/tools/reason.py`:

```python
from ..reasoning.branch_thinking_client import BranchThinkingClient
from typing import Dict, Any, List
import uuid

class ReasonTool:
    def __init__(self, lightrag_client, branch_thinking_path: str):
        self.lightrag = lightrag_client
        self.branch_client = BranchThinkingClient(branch_thinking_path)
    
    async def execute(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """
        Multi-branch reasoning with branch-thinking-mcp.
        
        Args:
            params: {
                'perception_uuid': UUID from perceive tool,
                'task': Task description,
                'num_branches': Number of reasoning branches (default: 3)
            }
        
        Returns:
            {
                'reasoning_uuid': UUID for this reasoning,
                'branches': List of reasoning branches,
                'selected': Best branch selected,
                'confidence': Confidence score,
                'insights': AI-generated insights
            }
        """
        perception_uuid = params['perception_uuid']
        task = params['task']
        num_branches = params.get('num_branches', 3)
        
        # 1. Retrieve perception from LightRAG
        perception = await self.lightrag.query_by_uuid(perception_uuid)
        
        # 2. Create reasoning branch in branch-thinking
        session_id = str(uuid.uuid4())
        branch_result = self.branch_client.create_branch(
            name=f"reasoning-{session_id}",
            description=f"Reasoning for: {task}"
        )
        branch_id = branch_result['branchId']
        
        # 3. Focus on branch
        self.branch_client.focus_branch(branch_id)
        
        # 4. Generate reasoning branches
        branches = []
        for i in range(num_branches):
            # Generate branch content based on perception
            branch_content = await self._generate_branch(
                perception=perception,
                task=task,
                branch_num=i
            )
            
            # Add as thought to branch-thinking
            thought_result = self.branch_client.add_thought(
                branch_id=branch_id,
                content=branch_content['reasoning'],
                thought_type='analysis'
            )
            
            branches.append({
                'id': thought_result['thoughtId'],
                'content': branch_content['reasoning'],
                'approach': branch_content['approach'],
                'confidence': branch_content['confidence']
            })
        
        # 5. Query for similar past reasoning
        similar = self.branch_client.semantic_search(
            query=task,
            limit=5
        )
        
        # 6. Link to similar thoughts
        for sim in similar:
            if sim['similarity'] > 0.8:
                self.branch_client.link_thoughts(
                    from_id=branches[0]['id'],
                    to_id=sim['thoughtId'],
                    link_type='similar_to',
                    reason=f"Similar reasoning pattern (similarity: {sim['similarity']:.2f})"
                )
        
        # 7. Get AI insights
        insights = self.branch_client.get_insights(branch_id)
        
        # 8. Select best branch based on insights and confidence
        best_branch = self._select_best_branch(branches, insights)
        
        # 9. Get summary
        summary = self.branch_client.summarize_branch(branch_id)
        
        # 10. Store in LightRAG
        reasoning_uuid = await self.lightrag.insert_reasoning(
            id=session_id,
            branches=[b['content'] for b in branches],
            selected=best_branch['content'],
            perception_uuid=perception_uuid,
            metadata={
                'branch_thinking_id': branch_id,
                'insights': insights,
                'summary': summary,
                'confidence': best_branch['confidence']
            }
        )
        
        return {
            'reasoning_uuid': reasoning_uuid,
            'branch_thinking_id': branch_id,
            'branches': branches,
            'selected': best_branch,
            'insights': insights,
            'summary': summary,
            'confidence': best_branch['confidence']
        }
    
    async def _generate_branch(self, perception: Dict, task: str, branch_num: int) -> Dict[str, Any]:
        """Generate a reasoning branch using LLM."""
        # Use Gemma 3 to generate reasoning
        # Different approaches for each branch
        approaches = ['deductive', 'inductive', 'abductive']
        approach = approaches[branch_num % len(approaches)]
        
        prompt = f"""
        Task: {task}
        
        Perception: {perception['content']}
        
        Use {approach} reasoning to solve this task.
        Provide step-by-step reasoning and a confidence score (0-1).
        """
        
        # Call Gemma 3 via Ollama
        response = await self._call_llm(prompt)
        
        return {
            'reasoning': response['reasoning'],
            'approach': approach,
            'confidence': response['confidence']
        }
    
    def _select_best_branch(self, branches: List[Dict], insights: List[Dict]) -> Dict:
        """Select best branch based on confidence and insights."""
        # Combine confidence scores with insight recommendations
        scores = []
        for branch in branches:
            score = branch['confidence']
            
            # Boost score if insights mention this branch positively
            for insight in insights:
                if branch['id'] in insight.get('relatedThoughts', []):
                    if insight['type'] == 'strength':
                        score += 0.1
                    elif insight['type'] == 'weakness':
                        score -= 0.1
            
            scores.append(score)
        
        best_idx = scores.index(max(scores))
        return branches[best_idx]
```

### Phase 4: Testing (1 day)

Create test cases:

```python
# test_branch_thinking_integration.py
import pytest
from dynamic_thinking.tools.reason import ReasonTool

@pytest.mark.asyncio
async def test_multi_branch_reasoning():
    # Setup
    reason_tool = ReasonTool(lightrag_client, '/path/to/branch-thinking-mcp')
    
    # Test
    result = await reason_tool.execute({
        'perception_uuid': 'test-perception-uuid',
        'task': 'Optimize database queries',
        'num_branches': 3
    })
    
    # Assertions
    assert len(result['branches']) == 3
    assert result['selected'] is not None
    assert result['confidence'] > 0.5
    assert len(result['insights']) > 0
```

---

## Benefits of Integration

### 1. Production-Ready Code

- ✅ Already tested and working
- ✅ TypeScript performance for graph operations
- ✅ Proven caching strategy
- ✅ Robust error handling

### 2. Rich Features

- ✅ Semantic search with embeddings
- ✅ AI summarization
- ✅ Insight generation
- ✅ Visualization
- ✅ Task extraction
- ✅ Cross-references

### 3. Performance

- ✅ LRU+TTL caching (30min embeddings, 5min summaries)
- ✅ Efficient graph algorithms (@dagrejs/graphlib)
- ✅ Fast k-means clustering
- ✅ Optimized cosine similarity

### 4. Extensibility

- ✅ Easy to add new thought types
- ✅ Pluggable embedding models
- ✅ Customizable visualization
- ✅ Extensible insight generation

---

## Alternative: Pure Python Implementation

If you prefer pure Python, reimplement key features:

```python
# backend/mcp_servers/dynamic_thinking/reasoning/branch_manager.py
from sentence_transformers import SentenceTransformer
from transformers import pipeline
import networkx as nx
from sklearn.cluster import KMeans
from cachetools import TTLCache
import numpy as np

class BranchManager:
    def __init__(self):
        # Embedding model
        self.embedder = SentenceTransformer('all-MiniLM-L6-v2')
        
        # Summarization model
        self.summarizer = pipeline('summarization', model='facebook/bart-large-cnn')
        
        # Caches
        self.embedding_cache = TTLCache(maxsize=1000, ttl=1800)  # 30 min
        self.summary_cache = TTLCache(maxsize=100, ttl=300)  # 5 min
        
        # Graph
        self.graph = nx.DiGraph()
        
        # Storage
        self.branches = {}
        self.thoughts = {}
    
    def create_branch(self, name: str, description: str = None):
        branch_id = str(uuid.uuid4())
        self.branches[branch_id] = {
            'id': branch_id,
            'name': name,
            'description': description,
            'thoughts': []
        }
        self.graph.add_node(branch_id, type='branch', name=name)
        return branch_id
    
    def add_thought(self, branch_id: str, content: str, thought_type: str):
        thought_id = str(uuid.uuid4())
        
        # Generate embedding
        if thought_id not in self.embedding_cache:
            embedding = self.embedder.encode(content)
            self.embedding_cache[thought_id] = embedding
        
        # Store thought
        self.thoughts[thought_id] = {
            'id': thought_id,
            'content': content,
            'type': thought_type,
            'branch_id': branch_id
        }
        
        # Add to graph
        self.graph.add_node(thought_id, type='thought', content=content)
        self.graph.add_edge(branch_id, thought_id, relation='contains')
        
        # Add to branch
        self.branches[branch_id]['thoughts'].append(thought_id)
        
        return thought_id
    
    def semantic_search(self, query: str, limit: int = 10):
        # Generate query embedding
        query_embedding = self.embedder.encode(query)
        
        # Calculate similarities
        similarities = []
        for thought_id, thought in self.thoughts.items():
            if thought_id in self.embedding_cache:
                thought_embedding = self.embedding_cache[thought_id]
                similarity = np.dot(query_embedding, thought_embedding) / (
                    np.linalg.norm(query_embedding) * np.linalg.norm(thought_embedding)
                )
                similarities.append((thought_id, similarity))
        
        # Sort and return top results
        similarities.sort(key=lambda x: x[1], reverse=True)
        return [
            {
                'thoughtId': tid,
                'similarity': sim,
                'content': self.thoughts[tid]['content']
            }
            for tid, sim in similarities[:limit]
        ]
    
    def summarize_branch(self, branch_id: str):
        # Check cache
        if branch_id in self.summary_cache:
            return self.summary_cache[branch_id]
        
        # Get all thoughts
        thought_ids = self.branches[branch_id]['thoughts']
        text = '\n'.join([self.thoughts[tid]['content'] for tid in thought_ids])
        
        # Summarize
        summary = self.summarizer(text, max_length=120, min_length=20, do_sample=False)
        result = summary[0]['summary_text']
        
        # Cache
        self.summary_cache[branch_id] = result
        
        return result
```

**Pros**:
- Pure Python (no subprocess calls)
- Full control over implementation
- Easy to customize

**Cons**:
- More work to implement
- Need to maintain parallel implementation
- May miss features from branch-thinking-mcp

---

## Recommendation

**Use branch-thinking-mcp via subprocess** (Option 1):

**Reasons**:
1. Production-ready code (already tested)
2. Rich features (semantic search, insights, visualization, tasks)
3. TypeScript performance for graph operations
4. Proven caching strategy
5. Easy to integrate (Python wrapper)
6. Can always reimplement later if needed

**Timeline**:
- Phase 1 (Setup): 1 day
- Phase 2 (Python Wrapper): 2 days
- Phase 3 (Integration): 3 days
- Phase 4 (Testing): 1 day
- **Total: 1 week**

---

## Summary

### What We Get

✅ **Multi-branch reasoning** - Create and evaluate multiple reasoning paths  
✅ **Semantic search** - Find similar past reasoning  
✅ **AI insights** - Automatic insight generation  
✅ **Visualization** - Knowledge graph with clustering  
✅ **Task extraction** - Identify actionable items  
✅ **Caching** - Fast repeated queries  
✅ **Production-ready** - Already tested and working  

### Integration Approach

1. Install branch-thinking-mcp in agent workspace
2. Create Python wrapper (BranchThinkingClient)
3. Integrate with ReasonTool in Dynamic Thinking MCP
4. Test with real reasoning tasks

### Result

A complete multi-branch reasoning system that:
- Generates multiple reasoning paths
- Searches for similar past reasoning
- Gets AI insights and summaries
- Selects best branch based on confidence and insights
- Stores results in LightRAG for future learning

**This completes the missing piece of Dynamic Thinking!** 🎯

