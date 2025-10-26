"""
Session Manager

Tracks PRAR loop sessions across multiple tool calls.
"""

import logging
from typing import Any, Dict, Optional
from datetime import datetime
from collections import defaultdict

logger = logging.getLogger(__name__)


class SessionManager:
    """Manages PRAR loop sessions"""
    
    def __init__(self):
        self.sessions = defaultdict(dict)
    
    def create_session(self, task_id: str) -> Dict[str, Any]:
        """Create a new session"""
        session = {
            "task_id": task_id,
            "created_at": datetime.utcnow().isoformat(),
            "phases": {},
            "status": "active"
        }
        self.sessions[task_id] = session
        logger.info(f"Created session: {task_id}")
        return session
    
    def update_session(self, task_id: str, phase: str, data: Dict[str, Any]):
        """Update session with phase data"""
        if task_id not in self.sessions:
            self.create_session(task_id)
        
        self.sessions[task_id]["phases"][phase] = {
            "data": data,
            "timestamp": datetime.utcnow().isoformat()
        }
        
        logger.info(f"Updated session {task_id} - phase: {phase}")
    
    def get_session(self, task_id: str) -> Optional[Dict[str, Any]]:
        """Get session data"""
        return self.sessions.get(task_id)
    
    def complete_session(self, task_id: str):
        """Mark session as complete"""
        if task_id in self.sessions:
            self.sessions[task_id]["status"] = "complete"
            self.sessions[task_id]["completed_at"] = datetime.utcnow().isoformat()
            logger.info(f"Completed session: {task_id}")
    
    def get_active_sessions(self) -> Dict[str, Dict[str, Any]]:
        """Get all active sessions"""
        return {
            task_id: session
            for task_id, session in self.sessions.items()
            if session.get("status") == "active"
        }

