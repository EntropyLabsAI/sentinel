from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.supervision_result import SupervisionResult
    from ..models.tool_request import ToolRequest


T = TypeVar("T", bound="CreateSupervisionResult")


@_attrs_define
class CreateSupervisionResult:
    """
    Attributes:
        execution_id (UUID):
        run_id (UUID):
        tool_id (UUID):
        supervisor_id (UUID):
        supervision_result (SupervisionResult):
        tool_request (ToolRequest): A tool request is a request to use a tool. It must be approved by a supervisor.
    """

    execution_id: UUID
    run_id: UUID
    tool_id: UUID
    supervisor_id: UUID
    supervision_result: "SupervisionResult"
    tool_request: "ToolRequest"
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        execution_id = str(self.execution_id)

        run_id = str(self.run_id)

        tool_id = str(self.tool_id)

        supervisor_id = str(self.supervisor_id)

        supervision_result = self.supervision_result.to_dict()

        tool_request = self.tool_request.to_dict()

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "execution_id": execution_id,
                "run_id": run_id,
                "tool_id": tool_id,
                "supervisor_id": supervisor_id,
                "supervision_result": supervision_result,
                "tool_request": tool_request,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.supervision_result import SupervisionResult
        from ..models.tool_request import ToolRequest

        d = src_dict.copy()
        execution_id = UUID(d.pop("execution_id"))

        run_id = UUID(d.pop("run_id"))

        tool_id = UUID(d.pop("tool_id"))

        supervisor_id = UUID(d.pop("supervisor_id"))

        supervision_result = SupervisionResult.from_dict(d.pop("supervision_result"))

        tool_request = ToolRequest.from_dict(d.pop("tool_request"))

        create_supervision_result = cls(
            execution_id=execution_id,
            run_id=run_id,
            tool_id=tool_id,
            supervisor_id=supervisor_id,
            supervision_result=supervision_result,
            tool_request=tool_request,
        )

        create_supervision_result.additional_properties = d
        return create_supervision_result

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
