from typing import TYPE_CHECKING, Any, Dict, List, Type, TypeVar
from uuid import UUID

from attrs import define as _attrs_define
from attrs import field as _attrs_field

if TYPE_CHECKING:
    from ..models.supervision_request import SupervisionRequest
    from ..models.supervision_result import SupervisionResult
    from ..models.supervision_status import SupervisionStatus


T = TypeVar("T", bound="ExecutionSupervisions")


@_attrs_define
class ExecutionSupervisions:
    """
    Attributes:
        execution_id (UUID):
        requests (List['SupervisionRequest']):
        results (List['SupervisionResult']):
        statuses (List['SupervisionStatus']):
    """

    execution_id: UUID
    requests: List["SupervisionRequest"]
    results: List["SupervisionResult"]
    statuses: List["SupervisionStatus"]
    additional_properties: Dict[str, Any] = _attrs_field(init=False, factory=dict)

    def to_dict(self) -> Dict[str, Any]:
        execution_id = str(self.execution_id)

        requests = []
        for requests_item_data in self.requests:
            requests_item = requests_item_data.to_dict()
            requests.append(requests_item)

        results = []
        for results_item_data in self.results:
            results_item = results_item_data.to_dict()
            results.append(results_item)

        statuses = []
        for statuses_item_data in self.statuses:
            statuses_item = statuses_item_data.to_dict()
            statuses.append(statuses_item)

        field_dict: Dict[str, Any] = {}
        field_dict.update(self.additional_properties)
        field_dict.update(
            {
                "execution_id": execution_id,
                "requests": requests,
                "results": results,
                "statuses": statuses,
            }
        )

        return field_dict

    @classmethod
    def from_dict(cls: Type[T], src_dict: Dict[str, Any]) -> T:
        from ..models.supervision_request import SupervisionRequest
        from ..models.supervision_result import SupervisionResult
        from ..models.supervision_status import SupervisionStatus

        d = src_dict.copy()
        execution_id = UUID(d.pop("execution_id"))

        requests = []
        _requests = d.pop("requests")
        for requests_item_data in _requests:
            requests_item = SupervisionRequest.from_dict(requests_item_data)

            requests.append(requests_item)

        results = []
        _results = d.pop("results")
        for results_item_data in _results:
            results_item = SupervisionResult.from_dict(results_item_data)

            results.append(results_item)

        statuses = []
        _statuses = d.pop("statuses")
        for statuses_item_data in _statuses:
            statuses_item = SupervisionStatus.from_dict(statuses_item_data)

            statuses.append(statuses_item)

        execution_supervisions = cls(
            execution_id=execution_id,
            requests=requests,
            results=results,
            statuses=statuses,
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
