"""
Headless PM Client - Lightweight Python client for integrating with Headless PM API
Compatible with Claude Code and other LLM agents.
"""

import requests
import json
import time
from typing import Optional, Dict, List, Any
from dataclasses import dataclass
from enum import Enum


class TaskStatus(Enum):
    """Task status enumeration"""
    CREATED = "created"
    ASSIGNED = "assigned"
    UNDER_WORK = "under_work"
    DEV_DONE = "dev_done"
    TESTING = "testing"
    COMPLETED = "completed"
    BLOCKED = "blocked"


class TaskComplexity(Enum):
    """Task complexity levels"""
    MINOR = "minor"  # Direct commit allowed
    MAJOR = "major"  # Requires PR


class SkillLevel(Enum):
    """Agent skill levels"""
    JUNIOR = "junior"
    SENIOR = "senior"
    PRINCIPAL = "principal"


@dataclass
class Task:
    """Task data structure"""
    id: int
    title: str
    description: str
    status: str
    complexity: str
    role: str
    skill_level: str
    created_at: str
    updated_at: str
    assigned_to: Optional[str] = None
    branch_name: Optional[str] = None


@dataclass
class Agent:
    """Agent data structure"""
    id: str
    role: str
    skill_level: str
    status: str
    last_heartbeat: str


class HeadlessPMClient:
    """
    Lightweight client for Headless PM API integration with Claude Code.

    Usage:
        client = HeadlessPMClient(agent_id="claude_001", role="backend_dev")
        client.register()
        task = client.get_next_task()
        if task:
            client.lock_task(task.id)
            # Do work...
            client.update_task_status(task.id, TaskStatus.DEV_DONE)
    """

    def __init__(self,
                 base_url: str = "http://localhost:6969",
                 agent_id: Optional[str] = None,
                 role: Optional[str] = None,
                 skill_level: SkillLevel = SkillLevel.SENIOR,
                 timeout: int = 30):
        """
        Initialize Headless PM client.

        Args:
            base_url: Base URL of Headless PM API
            agent_id: Unique identifier for this agent
            role: Agent role (frontend_dev, backend_dev, architect, etc.)
            skill_level: Agent skill level
            timeout: Request timeout in seconds
        """
        self.base_url = base_url.rstrip('/')
        self.agent_id = agent_id
        self.role = role
        self.skill_level = skill_level
        self.timeout = timeout
        self.session = requests.Session()
        self.session.headers.update({
            'Content-Type': 'application/json',
            'User-Agent': f'HeadlessPM-Client/1.0 ({agent_id})'
        })

    def _make_request(self, method: str, endpoint: str, **kwargs) -> requests.Response:
        """Make HTTP request with error handling."""
        url = f"{self.base_url}{endpoint}"
        try:
            response = self.session.request(method, url, timeout=self.timeout, **kwargs)
            response.raise_for_status()
            return response
        except requests.exceptions.RequestException as e:
            raise Exception(f"API request failed: {e}")

    def register(self) -> Dict[str, Any]:
        """
        Register agent with Headless PM system.

        Returns:
            Registration response data
        """
        if not self.agent_id or not self.role:
            raise ValueError("agent_id and role must be set before registration")

        data = {
            "agent_id": self.agent_id,
            "role": self.role,
            "skill_level": self.skill_level.value,
            "status": "active"
        }

        response = self._make_request("POST", "/api/v1/register", json=data)
        return response.json()

    def get_context(self) -> Dict[str, Any]:
        """
        Get project configuration and context.

        Returns:
            Project context data
        """
        response = self._make_request("GET", "/api/v1/context")
        return response.json()

    def create_task(self,
                    title: str,
                    description: str,
                    complexity: TaskComplexity,
                    role: Optional[str] = None,
                    skill_level: Optional[SkillLevel] = None) -> Task:
        """
        Create a new task.

        Args:
            title: Task title
            description: Task description
            complexity: Task complexity (minor/major)
            role: Required role (defaults to agent's role)
            skill_level: Required skill level (defaults to agent's level)

        Returns:
            Created task object
        """
        data = {
            "title": title,
            "description": description,
            "complexity": complexity.value,
            "role": role or self.role,
            "skill_level": skill_level.value if skill_level else self.skill_level.value
        }

        response = self._make_request("POST", "/api/v1/tasks/create", json=data)
        task_data = response.json()
        return Task(**task_data)

    def get_next_task(self) -> Optional[Task]:
        """
        Get next available task for this agent.

        Returns:
            Next task or None if no tasks available
        """
        params = {
            "role": self.role,
            "skill_level": self.skill_level.value
        }

        response = self._make_request("GET", "/api/v1/tasks/next", params=params)
        task_data = response.json()

        if not task_data:
            return None

        return Task(**task_data)

    def lock_task(self, task_id: int) -> bool:
        """
        Lock task to prevent other agents from picking it up.

        Args:
            task_id: ID of task to lock

        Returns:
            True if successfully locked
        """
        data = {"agent_id": self.agent_id}

        try:
            self._make_request("POST", f"/api/v1/tasks/{task_id}/lock", json=data)
            return True
        except Exception:
            return False

    def update_task_status(self, task_id: int, status: TaskStatus, notes: Optional[str] = None) -> bool:
        """
        Update task status.

        Args:
            task_id: ID of task to update
            status: New task status
            notes: Optional notes about the update

        Returns:
            True if successfully updated
        """
        data = {
            "status": status.value,
            "agent_id": self.agent_id
        }

        if notes:
            data["notes"] = notes

        try:
            self._make_request("PUT", f"/api/v1/tasks/{task_id}/status", json=data)
            return True
        except Exception:
            return False

    def evaluate_task(self, task_id: int, approved: bool, feedback: Optional[str] = None) -> bool:
        """
        Evaluate/approve a completed task (for architects/PMs).

        Args:
            task_id: ID of task to evaluate
            approved: Whether task is approved
            feedback: Optional feedback

        Returns:
            True if successfully evaluated
        """
        data = {
            "approved": approved,
            "evaluator": self.agent_id
        }

        if feedback:
            data["feedback"] = feedback

        try:
            self._make_request("POST", f"/api/v1/tasks/{task_id}/evaluate", json=data)
            return True
        except Exception:
            return False

    def create_document(self,
                        title: str,
                        content: str,
                        doc_type: str = "note",
                        mentions: Optional[List[str]] = None) -> Dict[str, Any]:
        """
        Create document with optional @mentions.

        Args:
            title: Document title
            content: Document content (supports @mentions)
            doc_type: Document type
            mentions: List of agent IDs to mention

        Returns:
            Created document data
        """
        data = {
            "title": title,
            "content": content,
            "type": doc_type,
            "author": self.agent_id
        }

        if mentions:
            data["mentions"] = mentions

        response = self._make_request("POST", "/api/v1/documents", json=data)
        return response.json()

    def get_documents(self, limit: int = 20, doc_type: Optional[str] = None) -> List[Dict[str, Any]]:
        """
        Get recent documents.

        Args:
            limit: Maximum number of documents to return
            doc_type: Filter by document type

        Returns:
            List of documents
        """
        params = {"limit": limit}
        if doc_type:
            params["type"] = doc_type

        response = self._make_request("GET", "/api/v1/documents", params=params)
        return response.json()

    def get_mentions(self) -> List[Dict[str, Any]]:
        """
        Get notifications/mentions for this agent.

        Returns:
            List of mentions/notifications
        """
        params = {"agent_id": self.agent_id}
        response = self._make_request("GET", "/api/v1/mentions", params=params)
        return response.json()

    def register_service(self, service_name: str, service_url: str, health_check_url: Optional[str] = None) -> bool:
        """
        Register a microservice.

        Args:
            service_name: Name of the service
            service_url: Service URL
            health_check_url: Health check endpoint

        Returns:
            True if successfully registered
        """
        data = {
            "name": service_name,
            "url": service_url,
            "registered_by": self.agent_id
        }

        if health_check_url:
            data["health_check_url"] = health_check_url

        try:
            self._make_request("POST", "/api/v1/services/register", json=data)
            return True
        except Exception:
            return False

    def send_heartbeat(self, service_name: str, status: str = "healthy") -> bool:
        """
        Send service heartbeat.

        Args:
            service_name: Name of the service
            status: Service status

        Returns:
            True if heartbeat sent successfully
        """
        data = {"status": status}

        try:
            self._make_request("POST", f"/api/v1/services/{service_name}/heartbeat", json=data)
            return True
        except Exception:
            return False

    def get_services(self) -> List[Dict[str, Any]]:
        """
        Get all registered services.

        Returns:
            List of services
        """
        response = self._make_request("GET", "/api/v1/services")
        return response.json()

    def poll_changes(self, since_timestamp: Optional[str] = None) -> Dict[str, Any]:
        """
        Poll for changes since timestamp.

        Args:
            since_timestamp: ISO timestamp to poll changes since

        Returns:
            Changes data
        """
        params = {}
        if since_timestamp:
            params["since"] = since_timestamp

        response = self._make_request("GET", "/api/v1/changes", params=params)
        return response.json()

    def get_changelog(self, limit: int = 50) -> List[Dict[str, Any]]:
        """
        Get recent activity changelog.

        Args:
            limit: Maximum number of entries

        Returns:
            List of changelog entries
        """
        params = {"limit": limit}
        response = self._make_request("GET", "/api/v1/changelog", params=params)
        return response.json()


# Convenience functions for Claude Code integration
def claude_register(agent_id: str, role: str, skill_level: str = "senior") -> HeadlessPMClient:
    """
    Quick registration function for Claude Code agents.

    Usage:
        client = claude_register("claude_001", "backend_dev")
    """
    client = HeadlessPMClient(
        agent_id=agent_id,
        role=role,
        skill_level=SkillLevel(skill_level)
    )
    client.register()
    return client


def claude_workflow_example():
    """
    Example workflow for Claude Code integration.
    """
    # Initialize and register agent
    client = claude_register("claude_example", "frontend_dev", "senior")

    # Get project context
    context = client.get_context()
    print(f"Project: {context.get('project_name', 'Unknown')}")

    # Check for next task
    task = client.get_next_task()
    if not task:
        print("No tasks available")
        return

    print(f"Got task: {task.title}")

    # Lock and start work
    if client.lock_task(task.id):
        client.update_task_status(task.id, TaskStatus.UNDER_WORK)

        # Simulate work
        print("Working on task...")
        time.sleep(1)

        # Update status
        client.update_task_status(task.id, TaskStatus.DEV_DONE,
                                  notes="Implemented feature with tests")

        # Create documentation
        client.create_document(
            title=f"Task {task.id} - Implementation Notes",
            content=f"Completed {task.title}\n\nImplementation details:\n- Added new feature\n- Updated tests\n- Ready for review",
            mentions=["architect_001"]  # Notify architect
        )

        print("Task completed and documented!")
    else:
        print("Could not lock task - may be taken by another agent")


if __name__ == "__main__":
    # Run example workflow
    claude_workflow_example()
