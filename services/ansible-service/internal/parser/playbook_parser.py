"""Parser YAML pour analyser les playbooks Ansible."""
import logging
from typing import Optional, List, Dict, Any
import yaml
from pathlib import Path

logger = logging.getLogger(__name__)


class PlaybookParser:
    """Parser pour analyser les playbooks Ansible en profondeur."""

    @staticmethod
    def parse_playbook(content: str) -> Optional[Dict[str, Any]]:
        """Parse le contenu YAML d'un playbook."""
        try:
            data = yaml.safe_load(content)
            if not data:
                return None

            # Un playbook peut être une liste de plays ou un seul play
            plays = data if isinstance(data, list) else [data]

            parsed_plays = []
            for play in plays:
                parsed_play = PlaybookParser._parse_play(play)
                if parsed_play:
                    parsed_plays.append(parsed_play)

            return {
                "plays": parsed_plays,
                "play_count": len(parsed_plays),
            }
        except yaml.YAMLError as e:
            logger.error(f"Erreur lors du parsing YAML: {e}")
            return None
        except Exception as e:
            logger.error(f"Erreur lors de l'analyse du playbook: {e}")
            return None

    @staticmethod
    def _parse_play(play: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Parse un play individuel."""
        if not isinstance(play, dict):
            return None

        parsed = {
            "name": play.get("name", "Unnamed play"),
            "hosts": play.get("hosts", "all"),
            "gather_facts": play.get("gather_facts", True),
            "vars": play.get("vars", {}),
            "vars_files": play.get("vars_files", []),
            "vars_prompt": play.get("vars_prompt", []),
            "pre_tasks": [],
            "tasks": [],
            "handlers": [],
            "post_tasks": [],
            "roles": [],
            "collections": play.get("collections", []),
            "become": play.get("become", False),
            "become_user": play.get("become_user"),
            "become_method": play.get("become_method"),
            "environment": play.get("environment", {}),
        }

        # Parser les pre_tasks
        if "pre_tasks" in play:
            parsed["pre_tasks"] = PlaybookParser._parse_tasks(play["pre_tasks"])

        # Parser les tasks
        if "tasks" in play:
            parsed["tasks"] = PlaybookParser._parse_tasks(play["tasks"])

        # Parser les handlers
        if "handlers" in play:
            parsed["handlers"] = PlaybookParser._parse_tasks(play["handlers"])

        # Parser les post_tasks
        if "post_tasks" in play:
            parsed["post_tasks"] = PlaybookParser._parse_tasks(play["post_tasks"])

        # Parser les roles
        if "roles" in play:
            parsed["roles"] = PlaybookParser._parse_roles(play["roles"])

        return parsed

    @staticmethod
    def _parse_tasks(tasks: List[Any]) -> List[Dict[str, Any]]:
        """Parse une liste de tâches."""
        parsed_tasks = []
        for task in tasks:
            if isinstance(task, dict):
                parsed_task = PlaybookParser._parse_task(task)
                if parsed_task:
                    parsed_tasks.append(parsed_task)
            elif isinstance(task, str):
                # Tâche inline simple
                parsed_tasks.append({
                    "name": task,
                    "type": "inline",
                })
        return parsed_tasks

    @staticmethod
    def _parse_task(task: Dict[str, Any]) -> Optional[Dict[str, Any]]:
        """Parse une tâche individuelle."""
        parsed = {
            "name": task.get("name", "Unnamed task"),
            "module": None,
            "args": {},
            "when": task.get("when"),
            "loop": task.get("loop"),
            "until": task.get("until"),
            "retries": task.get("retries"),
            "delay": task.get("delay"),
            "register": task.get("register"),
            "changed_when": task.get("changed_when"),
            "failed_when": task.get("failed_when"),
            "ignore_errors": task.get("ignore_errors", False),
            "async_val": task.get("async"),
            "poll": task.get("poll"),
            "delegate_to": task.get("delegate_to"),
            "delegate_facts": task.get("delegate_facts", False),
            "become": task.get("become"),
            "become_user": task.get("become_user"),
            "environment": task.get("environment", {}),
            "tags": task.get("tags", []),
            "notify": task.get("notify", []),
        }

        # Identifier le module utilisé
        # Les modules peuvent être définis comme clés directes ou via "action"
        module_keys = [k for k in task.keys() if k not in [
            "name", "when", "loop", "until", "retries", "delay", "register",
            "changed_when", "failed_when", "ignore_errors", "async", "poll",
            "delegate_to", "delegate_facts", "become", "become_user",
            "environment", "tags", "notify", "action", "local_action",
            "with_", "vars", "include", "include_tasks", "import_tasks",
            "block", "rescue", "always", "role", "include_role", "import_role",
        ]]

        if module_keys:
            parsed["module"] = module_keys[0]
            parsed["args"] = task.get(module_keys[0], {})
        elif "action" in task:
            action = task["action"]
            if isinstance(action, dict):
                parsed["module"] = list(action.keys())[0] if action else None
                parsed["args"] = list(action.values())[0] if action else {}
            elif isinstance(action, str):
                parsed["module"] = action.split()[0] if action else None

        # Gérer les includes et imports
        if "include" in task:
            parsed["type"] = "include"
            parsed["include"] = task["include"]
        elif "include_tasks" in task:
            parsed["type"] = "include_tasks"
            parsed["include_tasks"] = task["include_tasks"]
        elif "import_tasks" in task:
            parsed["type"] = "import_tasks"
            parsed["import_tasks"] = task["import_tasks"]
        elif "include_role" in task:
            parsed["type"] = "include_role"
            parsed["include_role"] = task["include_role"]
        elif "import_role" in task:
            parsed["type"] = "import_role"
            parsed["import_role"] = task["import_role"]
        elif "role" in task:
            parsed["type"] = "role"
            parsed["role"] = task["role"]
        elif "block" in task:
            parsed["type"] = "block"
            parsed["block"] = PlaybookParser._parse_tasks(task.get("block", []))
            parsed["rescue"] = PlaybookParser._parse_tasks(task.get("rescue", []))
            parsed["always"] = PlaybookParser._parse_tasks(task.get("always", []))
        else:
            parsed["type"] = "task"

        return parsed

    @staticmethod
    def _parse_roles(roles: List[Any]) -> List[Dict[str, Any]]:
        """Parse une liste de rôles."""
        parsed_roles = []
        for role in roles:
            if isinstance(role, str):
                parsed_roles.append({
                    "name": role,
                    "vars": {},
                })
            elif isinstance(role, dict):
                parsed_roles.append({
                    "name": role.get("role", role.get("name", "unknown")),
                    "vars": {k: v for k, v in role.items() if k != "role" and k != "name"},
                })
        return parsed_roles

    @staticmethod
    def analyze_playbook(content: str) -> Dict[str, Any]:
        """Analyse complète d'un playbook avec statistiques."""
        parsed = PlaybookParser.parse_playbook(content)
        if not parsed:
            return {"error": "Impossible de parser le playbook"}

        # Calculer les statistiques
        stats = {
            "total_plays": parsed["play_count"],
            "total_tasks": 0,
            "total_handlers": 0,
            "total_pre_tasks": 0,
            "total_post_tasks": 0,
            "total_roles": 0,
            "modules_used": set(),
            "hosts_targeted": set(),
            "become_used": False,
            "collections_used": set(),
        }

        for play in parsed["plays"]:
            stats["total_tasks"] += len(play.get("tasks", []))
            stats["total_handlers"] += len(play.get("handlers", []))
            stats["total_pre_tasks"] += len(play.get("pre_tasks", []))
            stats["total_post_tasks"] += len(play.get("post_tasks", []))
            stats["total_roles"] += len(play.get("roles", []))

            hosts = play.get("hosts")
            if isinstance(hosts, str):
                stats["hosts_targeted"].add(hosts)
            elif isinstance(hosts, list):
                stats["hosts_targeted"].update(hosts)

            if play.get("become"):
                stats["become_used"] = True

            stats["collections_used"].update(play.get("collections", []))

            # Extraire les modules utilisés
            for task_list in [play.get("tasks", []), play.get("handlers", []), 
                            play.get("pre_tasks", []), play.get("post_tasks", [])]:
                for task in task_list:
                    if task.get("module"):
                        stats["modules_used"].add(task["module"])

        return {
            "parsed": parsed,
            "statistics": {
                "total_plays": stats["total_plays"],
                "total_tasks": stats["total_tasks"],
                "total_handlers": stats["total_handlers"],
                "total_pre_tasks": stats["total_pre_tasks"],
                "total_post_tasks": stats["total_post_tasks"],
                "total_roles": stats["total_roles"],
                "modules_used": sorted(list(stats["modules_used"])),
                "hosts_targeted": sorted(list(stats["hosts_targeted"])),
                "become_used": stats["become_used"],
                "collections_used": sorted(list(stats["collections_used"])),
            },
        }
