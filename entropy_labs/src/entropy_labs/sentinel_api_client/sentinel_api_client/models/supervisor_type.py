from enum import Enum


class SupervisorType(str, Enum):
    CLIENT_SUPERVISOR = "client_supervisor"
    HUMAN_SUPERVISOR = "human_supervisor"

    def __str__(self) -> str:
        return str(self.value)
