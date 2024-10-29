from http import HTTPStatus
from typing import Any, Dict, List, Optional, Union

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.supervision_request import SupervisionRequest
from ...models.supervisor_type import SupervisorType
from ...types import UNSET, Response, Unset


def _get_kwargs(
    *,
    type: Union[Unset, SupervisorType] = UNSET,
) -> Dict[str, Any]:
    params: Dict[str, Any] = {}

    json_type: Union[Unset, str] = UNSET
    if not isinstance(type, Unset):
        json_type = type.value

    params["type"] = json_type

    params = {k: v for k, v in params.items() if v is not UNSET and v is not None}

    _kwargs: Dict[str, Any] = {
        "method": "get",
        "url": "/api/reviews",
        "params": params,
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[List["SupervisionRequest"]]:
    if response.status_code == 200:
        response_200 = []
        _response_200 = response.json()
        for response_200_item_data in _response_200:
            response_200_item = SupervisionRequest.from_dict(response_200_item_data)

            response_200.append(response_200_item)

        return response_200
    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[List["SupervisionRequest"]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
    type: Union[Unset, SupervisorType] = UNSET,
) -> Response[List["SupervisionRequest"]]:
    """List all supervisor requests

    Args:
        type (Union[Unset, SupervisorType]): The type of supervisor. ClientSupervisor means that
            the supervision is done client side and the server is merely informed. Other supervisor
            types are handled serverside, e.g. HumanSupervisor means that a human will review the
            request via the Sentinel UI.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[List['SupervisionRequest']]
    """

    kwargs = _get_kwargs(
        type=type,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    *,
    client: Union[AuthenticatedClient, Client],
    type: Union[Unset, SupervisorType] = UNSET,
) -> Optional[List["SupervisionRequest"]]:
    """List all supervisor requests

    Args:
        type (Union[Unset, SupervisorType]): The type of supervisor. ClientSupervisor means that
            the supervision is done client side and the server is merely informed. Other supervisor
            types are handled serverside, e.g. HumanSupervisor means that a human will review the
            request via the Sentinel UI.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        List['SupervisionRequest']
    """

    return sync_detailed(
        client=client,
        type=type,
    ).parsed


async def asyncio_detailed(
    *,
    client: Union[AuthenticatedClient, Client],
    type: Union[Unset, SupervisorType] = UNSET,
) -> Response[List["SupervisionRequest"]]:
    """List all supervisor requests

    Args:
        type (Union[Unset, SupervisorType]): The type of supervisor. ClientSupervisor means that
            the supervision is done client side and the server is merely informed. Other supervisor
            types are handled serverside, e.g. HumanSupervisor means that a human will review the
            request via the Sentinel UI.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[List['SupervisionRequest']]
    """

    kwargs = _get_kwargs(
        type=type,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    *,
    client: Union[AuthenticatedClient, Client],
    type: Union[Unset, SupervisorType] = UNSET,
) -> Optional[List["SupervisionRequest"]]:
    """List all supervisor requests

    Args:
        type (Union[Unset, SupervisorType]): The type of supervisor. ClientSupervisor means that
            the supervision is done client side and the server is merely informed. Other supervisor
            types are handled serverside, e.g. HumanSupervisor means that a human will review the
            request via the Sentinel UI.

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        List['SupervisionRequest']
    """

    return (
        await asyncio_detailed(
            client=client,
            type=type,
        )
    ).parsed
