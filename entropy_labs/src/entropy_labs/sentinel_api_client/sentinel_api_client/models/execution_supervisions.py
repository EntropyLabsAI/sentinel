from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.supervision import Supervision


T = TypeVar("T", bound="ExecutionSupervisions")


@_attrs_define
class ExecutionSupervisions:
    """
    Attributes:
        execution_id (UUID):
        supervisions (List[List['Supervision']]): A list of supervision chains. Each chain is a list of supervisions
            that took place in a chain.
    """

    execution_id: UUID
    supervisions: List[List["Supervision"]]
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        execution_id = str(self.execution_id)

        supervisions = []
        for supervisions_item_data in self.supervisions:
            supervisions_item = []
            for componentsschemas_supervision_chain_item_data in supervisions_item_data:
                componentsschemas_supervision_chain_item = componentsschemas_supervision_chain_item_data.to_dict()
                supervisions_item.append(componentsschemas_supervision_chain_item)

            supervisions.append(supervisions_item)

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "execution_id": execution_id,
                "supervisions": supervisions,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.supervision import Supervision

        d = src_dict.copy()
        execution_id = UUID(d.pop("execution_id"))

        supervisions = []
        _supervisions = d.pop("supervisions")
        for supervisions_item_data in _supervisions:
            supervisions_item = []
            _supervisions_item = supervisions_item_data
            for componentsschemas_supervision_chain_item_data in _supervisions_item:
                componentsschemas_supervision_chain_item = Supervision.from_dict(
                    componentsschemas_supervision_chain_item_data
                )

                supervisions_item.append(componentsschemas_supervision_chain_item)

            supervisions.append(supervisions_item)

        execution_supervisions = cls(
            execution_id=execution_id,
            supervisions=supervisions,
        )

        execution_supervisions.additional_properties = d
        return execution_supervisions

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
