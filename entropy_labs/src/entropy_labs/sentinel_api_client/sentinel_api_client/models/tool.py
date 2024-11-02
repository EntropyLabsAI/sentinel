import datetime
from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar, Union, cast
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..types import UNSET, Unset

if TYPE_CHECKING:
    from ..models.tool_attributes import ToolAttributes


T = TypeVar("T", bound="Tool")


@_attrs_define
class Tool:
    """
    Attributes:
        name (str):
        description (str):
        id (Union[Unset, UUID]):
        attributes (Union[Unset, ToolAttributes]): Attributes of the tool that requests to this tool will have
        ignored_attributes (Union[Unset, List[str]]): Attributes of the tool that will not be shown in the UI for
            requests to this tool
        created_at (Union[Unset, datetime.datetime]):
    """

    name: str
    description: str
    id: Union[Unset, UUID] = UNSET
    attributes: Union[Unset, "ToolAttributes"] = UNSET
    ignored_attributes: Union[Unset, List[str]] = UNSET
    created_at: Union[Unset, datetime.datetime] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        name = self.name

        description = self.description

        id: Union[Unset, str] = UNSET
        if not isinstance(self.id, Unset):
            id = str(self.id)

        attributes: Union[Unset, Dict[str, Any]] = UNSET
        if not isinstance(self.attributes, Unset):
            attributes = self.attributes.to_dict()

        ignored_attributes: Union[Unset, List[str]] = UNSET
        if not isinstance(self.ignored_attributes, Unset):
            ignored_attributes = self.ignored_attributes

        created_at: Union[Unset, str] = UNSET
        if not isinstance(self.created_at, Unset):
            created_at = self.created_at.isoformat()

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "name": name,
                "description": description,
            }
        )
        if id is not UNSET:
            field_dict["id"] = id
        if attributes is not UNSET:
            field_dict["attributes"] = attributes
        if ignored_attributes is not UNSET:
            field_dict["ignored_attributes"] = ignored_attributes
        if created_at is not UNSET:
            field_dict["created_at"] = created_at

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_attributes import ToolAttributes

        d = src_dict.copy()
        name = d.pop("name")

        description = d.pop("description")

        _id = d.pop("id", UNSET)
        id: Union[Unset, UUID]
        if isinstance(_id, Unset):
            id = UNSET
        else:
            id = UUID(_id)

        _attributes = d.pop("attributes", UNSET)
        attributes: Union[Unset, ToolAttributes]
        if isinstance(_attributes, Unset):
            attributes = UNSET
        else:
            attributes = ToolAttributes.from_dict(_attributes)

        ignored_attributes = cast(List[str], d.pop("ignored_attributes", UNSET))

        _created_at = d.pop("created_at", UNSET)
        created_at: Union[Unset, datetime.datetime]
        if isinstance(_created_at, Unset):
            created_at = UNSET
        else:
            created_at = isoparse(_created_at)

        tool = cls(
            name=name,
            description=description,
            id=id,
            attributes=attributes,
            ignored_attributes=ignored_attributes,
            created_at=created_at,
        )

        tool.additional_properties = d
        return tool

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
