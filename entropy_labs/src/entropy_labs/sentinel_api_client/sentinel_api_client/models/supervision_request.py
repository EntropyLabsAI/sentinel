from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.llm_message import LLMMessage
    from ..models.supervision_status import SupervisionStatus
    from ..models.task_state import TaskState
    from ..models.tool_request import ToolRequest


T = TypeVar("T", bound="SupervisionRequest")


@_attrs_define
class SupervisionRequest:
    """
    Attributes:
        run_id (UUID):
        execution_id (UUID):
        supervisor_id (UUID):
        task_state (TaskState):
        tool_requests (List['ToolRequest']):
        messages (List['LLMMessage']):
        id (Union[Unset, UUID]):
        status (Union[Unset, SupervisionStatus]):
    """

    run_id: UUID
    execution_id: UUID
    supervisor_id: UUID
    task_state: "TaskState"
    tool_requests: List["ToolRequest"]
    messages: List["LLMMessage"]
    id: Union[Unset, UUID] = UNSET
    status: Union[Unset, "SupervisionStatus"] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        run_id = str(self.run_id)

        execution_id = str(self.execution_id)

        supervisor_id = str(self.supervisor_id)

        task_state = self.task_state.to_dict()

        tool_requests = []
        for tool_requests_item_data in self.tool_requests:
            tool_requests_item = tool_requests_item_data.to_dict()
            tool_requests.append(tool_requests_item)

        messages = []
        for messages_item_data in self.messages:
            messages_item = messages_item_data.to_dict()
            messages.append(messages_item)

        id: Union[Unset, str] = UNSET
        if not isinstance(self.id, Unset):
            id = str(self.id)

        status: Union[Unset, Dict[str, Any]] = UNSET
        if not isinstance(self.status, Unset):
            status = self.status.to_dict()

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "run_id": run_id,
                "execution_id": execution_id,
                "supervisor_id": supervisor_id,
                "task_state": task_state,
                "tool_requests": tool_requests,
                "messages": messages,
            }
        )
        if id is not UNSET:
            field_dict["id"] = id
        if status is not UNSET:
            field_dict["status"] = status

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.llm_message import LLMMessage
        from ..models.supervision_status import SupervisionStatus
        from ..models.task_state import TaskState
        from ..models.tool_request import ToolRequest

        d = src_dict.copy()
        run_id = UUID(d.pop("run_id"))

        execution_id = UUID(d.pop("execution_id"))

        supervisor_id = UUID(d.pop("supervisor_id"))

        task_state = TaskState.from_dict(d.pop("task_state"))

        tool_requests = []
        _tool_requests = d.pop("tool_requests")
        for tool_requests_item_data in _tool_requests:
            tool_requests_item = ToolRequest.from_dict(tool_requests_item_data)

            tool_requests.append(tool_requests_item)

        messages = []
        _messages = d.pop("messages")
        for messages_item_data in _messages:
            messages_item = LLMMessage.from_dict(messages_item_data)

            messages.append(messages_item)

        _id = d.pop("id", UNSET)
        id: Union[Unset, UUID]
        if isinstance(_id, Unset):
            id = UNSET
        else:
            id = UUID(_id)

        _status = d.pop("status", UNSET)
        status: Union[Unset, SupervisionStatus]
        if isinstance(_status, Unset):
            status = UNSET
        else:
            status = SupervisionStatus.from_dict(_status)

        supervision_request = cls(
            run_id=run_id,
            execution_id=execution_id,
            supervisor_id=supervisor_id,
            task_state=task_state,
            tool_requests=tool_requests,
            messages=messages,
            id=id,
            status=status,
        )

        supervision_request.additional_properties = d
        return supervision_request

    @property
    def additional_keys(self) -> List[str]:
        return list(self.additional_properties.keys())

    def __getitem__(self, key: str) -> Any:
        return self.additional_properties[key]

    def __setitem__(self, key: str, value: Any) -> None:
        self.additional_properties[key] = value

    def __delitem__(self, key: str) -> None:
        del self.additional_properties[key]

    def __contains__(self, key: str) -> bool:
        return key in self.additional_properties
