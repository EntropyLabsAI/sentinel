from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.tool_call import ToolCall


T = TypeVar("T", bound="AssistantMessage")


@_attrs_define
class AssistantMessage:
    """
    Attributes:
        content (str):
        role (str):
        source (Union[Unset, str]):
        tool_calls (Union[Unset, List['ToolCall']]):
    """

    content: str
    role: str
    source: Union[Unset, str] = UNSET
    tool_calls: Union[Unset, List["ToolCall"]] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        content = self.content

        role = self.role

        source = self.source

        tool_calls: Union[Unset, List[Dict[str, Any]]] = UNSET
        if not isinstance(self.tool_calls, Unset):
            tool_calls = []
            for tool_calls_item_data in self.tool_calls:
                tool_calls_item = tool_calls_item_data.to_dict()
                tool_calls.append(tool_calls_item)

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "content": content,
                "role": role,
            }
        )
        if source is not UNSET:
            field_dict["source"] = source
        if tool_calls is not UNSET:
            field_dict["tool_calls"] = tool_calls

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_call import ToolCall

        d = src_dict.copy()
        content = d.pop("content")

        role = d.pop("role")

        source = d.pop("source", UNSET)

        tool_calls = []
        _tool_calls = d.pop("tool_calls", UNSET)
        for tool_calls_item_data in _tool_calls or []:
            tool_calls_item = ToolCall.from_dict(tool_calls_item_data)

            tool_calls.append(tool_calls_item)

        assistant_message = cls(
            content=content,
            role=role,
            source=source,
            tool_calls=tool_calls,
        )

        assistant_message.additional_properties = d
        return assistant_message

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
