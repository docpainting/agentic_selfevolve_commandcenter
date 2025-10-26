"""
LightRAG Client Wrapper

Wraps go-light-rag functionality for use in Python MCP server.
Since go-light-rag is a Go library, this client will:
1. Use subprocess to call Go CLI wrapper, OR
2. Use HTTP API if we expose go-light-rag via HTTP, OR
3. Use direct Neo4j + ChromaDB clients (simpler for now)

For now, using direct clients to Neo4j and ChromaDB.
"""

import logging
import uuid
from typing import Any, Dict, List, Optional
from datetime import datetime

from neo4j import GraphDatabase
import chromadb
from chromadb.config import Settings
import ollama

logger = logging.getLogger(__name__)


class LightRAGClient:
    """Client for LightRAG storage and retrieval"""
    
    def __init__(
        self,
        llm_model: str = "gemma3:27b",
        embedding_model: str = "nomic-embed-text:v1.5",
        ollama_base_url: str = "http://localhost:11434",
        neo4j_uri: str = "bolt://localhost:7687",
        neo4j_user: str = "neo4j",
        neo4j_password: str = "password",
        vector_db_path: str = "/var/lib/agent/vectors.db",
        kv_db_path: str = "/var/lib/agent/kv.db"
    ):
        self.llm_model = llm_model
        self.embedding_model = embedding_model
        self.ollama_base_url = ollama_base_url
        
        # Neo4j
        self.neo4j_uri = neo4j_uri
        self.neo4j_user = neo4j_user
        self.neo4j_password = neo4j_password
        self.neo4j_driver = None
        
        # ChromaDB
        self.vector_db_path = vector_db_path
        self.chroma_client = None
        self.collection = None
        
        # Ollama client
        self.ollama_client = ollama.Client(host=ollama_base_url)
    
    async def initialize(self):
        """Initialize connections"""
        logger.info("Initializing LightRAG client...")
        
        # Initialize Neo4j
        self.neo4j_driver = GraphDatabase.driver(
            self.neo4j_uri,
            auth=(self.neo4j_user, self.neo4j_password)
        )
        
        # Test connection
        with self.neo4j_driver.session() as session:
            result = session.run("RETURN 1 as test")
            logger.info(f"Neo4j connection successful: {result.single()['test']}")
        
        # Initialize ChromaDB
        self.chroma_client = chromadb.PersistentClient(
            path=self.vector_db_path,
            settings=Settings(anonymized_telemetry=False)
        )
        
        # Get or create collection
        self.collection = self.chroma_client.get_or_create_collection(
            name="agent_memory",
            metadata={"description": "Agent perception, reasoning, and learning"}
        )
        
        logger.info("LightRAG client initialized successfully")
    
    def create_embedding(self, text: str) -> List[float]:
        """Create embedding using Ollama"""
        try:
            response = self.ollama_client.embeddings(
                model=self.embedding_model,
                prompt=text
            )
            return response['embedding']
        except Exception as e:
            logger.error(f"Error creating embedding: {e}")
            raise
    
    def insert_document(
        self,
        doc_id: str,
        content: str,
        doc_type: str,
        metadata: Dict[str, Any]
    ) -> str:
        """Insert document into LightRAG (Neo4j + ChromaDB)"""
        try:
            # Generate UUID if not provided
            if not doc_id:
                doc_id = str(uuid.uuid4())
            
            # Create embedding
            embedding = self.create_embedding(content)
            
            # Store in ChromaDB
            self.collection.add(
                ids=[doc_id],
                embeddings=[embedding],
                documents=[content],
                metadatas=[{
                    "type": doc_type,
                    "timestamp": datetime.utcnow().isoformat(),
                    **metadata
                }]
            )
            
            # Store in Neo4j
            with self.neo4j_driver.session() as session:
                session.run(
                    f"""
                    MERGE (n:{doc_type} {{id: $id}})
                    SET n.content = $content,
                        n.timestamp = $timestamp,
                        n.metadata = $metadata
                    """,
                    id=doc_id,
                    content=content,
                    timestamp=datetime.utcnow().isoformat(),
                    metadata=metadata
                )
            
            logger.info(f"Inserted document {doc_id} (type: {doc_type})")
            return doc_id
            
        except Exception as e:
            logger.error(f"Error inserting document: {e}")
            raise
    
    def query_similar(
        self,
        query: str,
        doc_type: Optional[str] = None,
        max_results: int = 10,
        filters: Optional[Dict[str, Any]] = None
    ) -> List[Dict[str, Any]]:
        """Query similar documents using vector search"""
        try:
            # Create query embedding
            query_embedding = self.create_embedding(query)
            
            # Build where clause for filters
            where = {}
            if doc_type:
                where["type"] = doc_type
            if filters:
                where.update(filters)
            
            # Query ChromaDB
            results = self.collection.query(
                query_embeddings=[query_embedding],
                n_results=max_results,
                where=where if where else None
            )
            
            # Format results
            formatted_results = []
            if results['ids'] and results['ids'][0]:
                for i, doc_id in enumerate(results['ids'][0]):
                    formatted_results.append({
                        "id": doc_id,
                        "content": results['documents'][0][i],
                        "metadata": results['metadatas'][0][i],
                        "distance": results['distances'][0][i] if 'distances' in results else None
                    })
            
            logger.info(f"Found {len(formatted_results)} similar documents")
            return formatted_results
            
        except Exception as e:
            logger.error(f"Error querying similar documents: {e}")
            raise
    
    def get_by_id(self, doc_id: str) -> Optional[Dict[str, Any]]:
        """Get document by ID"""
        try:
            result = self.collection.get(ids=[doc_id])
            
            if result['ids']:
                return {
                    "id": result['ids'][0],
                    "content": result['documents'][0],
                    "metadata": result['metadatas'][0]
                }
            return None
            
        except Exception as e:
            logger.error(f"Error getting document by ID: {e}")
            raise
    
    def create_relationship(
        self,
        from_id: str,
        to_id: str,
        relationship_type: str,
        properties: Optional[Dict[str, Any]] = None
    ):
        """Create relationship in Neo4j"""
        try:
            with self.neo4j_driver.session() as session:
                query = f"""
                MATCH (a {{id: $from_id}})
                MATCH (b {{id: $to_id}})
                MERGE (a)-[r:{relationship_type}]->(b)
                """
                
                if properties:
                    query += "SET r += $properties"
                
                session.run(
                    query,
                    from_id=from_id,
                    to_id=to_id,
                    properties=properties or {}
                )
            
            logger.info(f"Created relationship: {from_id} -[{relationship_type}]-> {to_id}")
            
        except Exception as e:
            logger.error(f"Error creating relationship: {e}")
            raise
    
    def find_similar_entities(
        self,
        entity_name: str,
        entity_type: str,
        min_similarity: float = 0.7,
        max_results: int = 5
    ) -> List[Dict[str, Any]]:
        """Find similar entities in graph"""
        try:
            # Query for similar entities
            query = f"""
            Find entities similar to {entity_name} (type: {entity_type})
            """
            
            results = self.query_similar(
                query=query,
                doc_type="Entity",
                max_results=max_results
            )
            
            # Filter by similarity threshold
            similar_entities = []
            for result in results:
                # Distance is inverse of similarity (lower distance = higher similarity)
                similarity = 1.0 - (result.get('distance', 1.0) / 2.0)  # Normalize
                
                if similarity >= min_similarity:
                    similar_entities.append({
                        "id": result['id'],
                        "name": result['metadata'].get('name', 'Unknown'),
                        "type": result['metadata'].get('entity_type', entity_type),
                        "similarity": similarity,
                        "metadata": result['metadata']
                    })
            
            return similar_entities
            
        except Exception as e:
            logger.error(f"Error finding similar entities: {e}")
            raise
    
    def get_related_nodes(
        self,
        node_id: str,
        relationship_type: Optional[str] = None,
        max_depth: int = 1
    ) -> List[Dict[str, Any]]:
        """Get related nodes from graph"""
        try:
            with self.neo4j_driver.session() as session:
                if relationship_type:
                    query = f"""
                    MATCH (n {{id: $node_id}})-[r:{relationship_type}*1..{max_depth}]-(related)
                    RETURN related, r
                    LIMIT 20
                    """
                else:
                    query = f"""
                    MATCH (n {{id: $node_id}})-[r*1..{max_depth}]-(related)
                    RETURN related, r
                    LIMIT 20
                    """
                
                result = session.run(query, node_id=node_id)
                
                related_nodes = []
                for record in result:
                    node = record['related']
                    related_nodes.append({
                        "id": node.get('id'),
                        "labels": list(node.labels),
                        "properties": dict(node)
                    })
                
                return related_nodes
                
        except Exception as e:
            logger.error(f"Error getting related nodes: {e}")
            raise
    
    async def close(self):
        """Close connections"""
        if self.neo4j_driver:
            self.neo4j_driver.close()
        logger.info("LightRAG client closed")

