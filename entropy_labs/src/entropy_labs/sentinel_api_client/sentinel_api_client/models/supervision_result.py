import datetime
from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field
from dateutil.parser import isoparse

from ..models.decision import Decision

if TYPE_CHECKING:
    from ..models.tool_request import ToolRequest


T = TypeVar("T", bound="SupervisionResult")


@_attrs_define
class SupervisionResult:
    """
    Attributes:
        id (UUID):
        supervision_request_id (UUID):
        created_at (datetime.datetime):
        decision (Decision):
        toolrequest (ToolRequest): A tool request is a request to use a tool. It must be approved by a supervisor.
        reasoning (str):
    """

    id: UUID
    supervision_request_id: UUID
    created_at: datetime.datetime
    decision: Decision
    toolrequest: "ToolRequest"
    reasoning: str
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        id = str(self.id)

        supervision_request_id = str(self.supervision_request_id)

        created_at = self.created_at.isoformat()

        decision = self.decision.value

        toolrequest = self.toolrequest.to_dict()

        reasoning = self.reasoning

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "id": id,
                "supervision_request_id": supervision_request_id,
                "created_at": created_at,
                "decision": decision,
                "toolrequest": toolrequest,
                "reasoning": reasoning,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.tool_request import ToolRequest

        d = src_dict.copy()
        id = UUID(d.pop("id"))

        supervision_request_id = UUID(d.pop("supervision_request_id"))

        created_at = isoparse(d.pop("created_at"))

        decision = Decision(d.pop("decision"))

        toolrequest = ToolRequest.from_dict(d.pop("toolrequest"))

        reasoning = d.pop("reasoning")

        supervision_result = cls(
            id=id,
            supervision_request_id=supervision_request_id,
            created_at=created_at,
            decision=decision,
            toolrequest=toolrequest,
            reasoning=reasoning,
        )

        supervision_result.additional_properties = d
        return supervision_result

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
