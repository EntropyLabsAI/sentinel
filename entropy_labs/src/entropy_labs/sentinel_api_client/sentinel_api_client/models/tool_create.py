from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union

from attrs import define as _attrs_define
from attrs import field as _attrs_field

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.tool_create_attributes import ToolCreateAttributes


T = TypeVar("T", bound="ToolCreate")


@_attrs_define
class ToolCreate:
    """
    Attributes:
        name (str):
        description (str):
        attributes (Union[Unset, ToolCreateAttributes]):
    """

    name: str
    description: str
    attributes: Union[Unset, "ToolCreateAttributes"] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        name = self.name

        description = self.description

        attributes: Union[Unset, Dict[str, Any]] = UNSET
        if not isinstance(self.attributes, Unset):
            attributes = self.attributes.to_dict()

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "name": name,
                "description": description,
            }
        )
        if attributes is not UNSET:
            field_dict["attributes"] = attributes

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_create_attributes import ToolCreateAttributes

        d = src_dict.copy()
        name = d.pop("name")

        description = d.pop("description")

        _attributes = d.pop("attributes", UNSET)
        attributes: Union[Unset, ToolCreateAttributes]
        if isinstance(_attributes, Unset):
            attributes = UNSET
        else:
            attributes = ToolCreateAttributes.from_dict(_attributes)

        tool_create = cls(
            name=name,
            description=description,
            attributes=attributes,
        )

        tool_create.additional_properties = d
        return tool_create

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
