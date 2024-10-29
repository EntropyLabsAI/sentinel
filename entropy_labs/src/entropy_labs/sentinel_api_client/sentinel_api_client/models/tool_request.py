from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.tool_request_arguments import ToolRequestArguments


T = TypeVar("T", bound="ToolRequest")


@_attrs_define
class ToolRequest:
    """A tool request is a request to use a tool. It must be approved by a supervisor.

    Attributes:
        tool_id (UUID):
        arguments (ToolRequestArguments):
        id (Union[Unset, UUID]):
        supervision_request_id (Union[Unset, UUID]):
        message_id (Union[Unset, UUID]):
    """

    tool_id: UUID
    arguments: "ToolRequestArguments"
    id: Union[Unset, UUID] = UNSET
    supervision_request_id: Union[Unset, UUID] = UNSET
    message_id: Union[Unset, UUID] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        tool_id = str(self.tool_id)

        arguments = self.arguments.to_dict()

        id: Union[Unset, str] = UNSET
        if not isinstance(self.id, Unset):
            id = str(self.id)

        supervision_request_id: Union[Unset, str] = UNSET
        if not isinstance(self.supervision_request_id, Unset):
            supervision_request_id = str(self.supervision_request_id)

        message_id: Union[Unset, str] = UNSET
        if not isinstance(self.message_id, Unset):
            message_id = str(self.message_id)

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "tool_id": tool_id,
                "arguments": arguments,
            }
        )
        if id is not UNSET:
            field_dict["id"] = id
        if supervision_request_id is not UNSET:
            field_dict["supervision_request_id"] = supervision_request_id
        if message_id is not UNSET:
            field_dict["message_id"] = message_id

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_request_arguments import ToolRequestArguments

        d = src_dict.copy()
        tool_id = UUID(d.pop("tool_id"))

        arguments = ToolRequestArguments.from_dict(d.pop("arguments"))

        _id = d.pop("id", UNSET)
        id: Union[Unset, UUID]
        if isinstance(_id, Unset):
            id = UNSET
        else:
            id = UUID(_id)

        _supervision_request_id = d.pop("supervision_request_id", UNSET)
        supervision_request_id: Union[Unset, UUID]
        if isinstance(_supervision_request_id, Unset):
            supervision_request_id = UNSET
        else:
            supervision_request_id = UUID(_supervision_request_id)

        _message_id = d.pop("message_id", UNSET)
        message_id: Union[Unset, UUID]
        if isinstance(_message_id, Unset):
            message_id = UNSET
        else:
            message_id = UUID(_message_id)

        tool_request = cls(
            tool_id=tool_id,
            arguments=arguments,
            id=id,
            supervision_request_id=supervision_request_id,
            message_id=message_id,
        )

        tool_request.additional_properties = d
        return tool_request

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
