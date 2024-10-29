import datetime
from typing import Any, Dict, List, Type, TypeVar, Union
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.supervisor_type import SupervisorType
from ..types import UNSET, Unset

T = TypeVar("T", bound="Supervisor")


@_attrs_define
class Supervisor:
    """
    Attributes:
        description (str):
        created_at (datetime.datetime):
        type (SupervisorType): The type of supervisor. ClientSupervisor means that the supervision is done client side
            and the server is merely informed. Other supervisor types are handled serverside, e.g. HumanSupervisor means
            that a human will review the request via the Sentinel UI.
        name (str):
        id (Union[Unset, UUID]):
        code (Union[Unset, str]):
    """

    description: str
    created_at: datetime.datetime
    type: SupervisorType
    name: str
    id: Union[Unset, UUID] = UNSET
    code: Union[Unset, str] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        description = self.description

        created_at = self.created_at.isoformat()

        type = self.type.value

        name = self.name

        id: Union[Unset, str] = UNSET
        if not isinstance(self.id, Unset):
            id = str(self.id)

        code = self.code

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "description": description,
                "created_at": created_at,
                "type": type,
                "name": name,
            }
        )
        if id is not UNSET:
            field_dict["id"] = id
        if code is not UNSET:
            field_dict["code"] = code

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        d = src_dict.copy()
        description = d.pop("description")

        created_at = isoparse(d.pop("created_at"))

        type = SupervisorType(d.pop("type"))

        name = d.pop("name")

        _id = d.pop("id", UNSET)
        id: Union[Unset, UUID]
        if isinstance(_id, Unset):
            id = UNSET
        else:
            id = UUID(_id)

        code = d.pop("code", UNSET)

        supervisor = cls(
            description=description,
            created_at=created_at,
            type=type,
            name=name,
            id=id,
            code=code,
        )

        supervisor.additional_properties = d
        return supervisor

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
