import datetime
from typing import Any, Dict, List, Type, TypeVar, Union
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.status import Status
from ..types import UNSET, Unset

T = TypeVar("T", bound="Execution")


@_attrs_define
class Execution:
    """
    Attributes:
        id (UUID):
        run_id (Union[Unset, UUID]):
        tool_id (Union[Unset, UUID]):
        created_at (Union[Unset, datetime.datetime]):
        status (Union[Unset, Status]):
    """

    id: UUID
    run_id: Union[Unset, UUID] = UNSET
    tool_id: Union[Unset, UUID] = UNSET
    created_at: Union[Unset, datetime.datetime] = UNSET
    status: Union[Unset, Status] = UNSET
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        id = str(self.id)

        run_id: Union[Unset, str] = UNSET
        if not isinstance(self.run_id, Unset):
            run_id = str(self.run_id)

        tool_id: Union[Unset, str] = UNSET
        if not isinstance(self.tool_id, Unset):
            tool_id = str(self.tool_id)

        created_at: Union[Unset, str] = UNSET
        if not isinstance(self.created_at, Unset):
            created_at = self.created_at.isoformat()

        status: Union[Unset, str] = UNSET
        if not isinstance(self.status, Unset):
            status = self.status.value

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
            }
        )
        if run_id is not UNSET:
            field_dict["run_id"] = run_id
        if tool_id is not UNSET:
            field_dict["tool_id"] = tool_id
        if created_at is not UNSET:
            field_dict["created_at"] = created_at
        if status is not UNSET:
            field_dict["status"] = status

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        d = src_dict.copy()
        id = UUID(d.pop("id"))

        _run_id = d.pop("run_id", UNSET)
        run_id: Union[Unset, UUID]
        if isinstance(_run_id, Unset):
            run_id = UNSET
        else:
            run_id = UUID(_run_id)

        _tool_id = d.pop("tool_id", UNSET)
        tool_id: Union[Unset, UUID]
        if isinstance(_tool_id, Unset):
            tool_id = UNSET
        else:
            tool_id = UUID(_tool_id)

        _created_at = d.pop("created_at", UNSET)
        created_at: Union[Unset, datetime.datetime]
        if isinstance(_created_at, Unset):
            created_at = UNSET
        else:
            created_at = isoparse(_created_at)

        _status = d.pop("status", UNSET)
        status: Union[Unset, Status]
        if isinstance(_status, Unset):
            status = UNSET
        else:
            status = Status(_status)

        execution = cls(
            id=id,
            run_id=run_id,
            tool_id=tool_id,
            created_at=created_at,
            status=status,
        )

        execution.additional_properties = d
        return execution

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
