"""
Query Memory Tool

Semantic search across agent's memory.
"""

import logging
from typing import Any, Dict

logger = logging.getLogger(__name__)


class QueryMemoryTool:
    """Memory query tool"""
    
    def __init__(self, lightrag_client):
        self.lightrag = lightrag_client
    
    async def execute(self, params: Dict[str, Any]) -> Dict[str, Any]:
        """Execute memory query"""
        query = params['query']
        context = params.get('context', '')
        filters = params.get('filters', {})
        max_results = params.get('max_results', 10)
        
        logger.info(f"Querying memory: {query}")
        
        # Query LightRAG
        results = self.lightrag.query_similar(
            query=f"{query}\n{context}" if context else query,
            doc_type=filters.get('type'),
            max_results=max_results,
            filters=filters
        )
        
        # Format results
        formatted_results = []
        for result in results:
            formatted_results.append({
                "id": result['id'],
                "type": result['metadata'].get('type', 'unknown'),
                "content": result['content'],
                "relevance": 1.0 - (result.get('distance', 1.0) / 2.0),  # Normalize
                "metadata": result['metadata'],
                "related_entities": []  # TODO: Get from graph
            })
        
        return {
            "results": formatted_results,
            "total_found": len(formatted_results),
            "query_confidence": 0.8
        }

