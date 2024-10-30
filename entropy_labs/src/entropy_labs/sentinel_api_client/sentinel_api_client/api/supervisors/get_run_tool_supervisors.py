from http import HTTPStatus
from typing import Any, Dict, List, Optional, Union, cast
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.supervisor import Supervisor
from ...types import Response


def _get_kwargs(
    run_id: UUID,
    tool_id: UUID,
) -> Dict[str, Any]:
    _kwargs: Dict[str, Any] = {
        "method": "get",
        "url": f"/api/runs/{run_id}/tools/{tool_id}/supervisors",
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[Any, List["Supervisor"]]]:
    if response.status_code == 200:
        response_200 = []
        _response_200 = response.json()
        for response_200_item_data in _response_200:
            response_200_item = Supervisor.from_dict(response_200_item_data)

            response_200.append(response_200_item)

        return response_200
    if response.status_code == 404:
        response_404 = cast(Any, None)
        return response_404
    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[Any, List["Supervisor"]]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[Any, List["Supervisor"]]]:
    """Get the supervisors assigned to a tool

    Args:
        run_id (UUID):
        tool_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, List['Supervisor']]]
    """

    kwargs = _get_kwargs(
        run_id=run_id,
        tool_id=tool_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[Any, List["Supervisor"]]]:
    """Get the supervisors assigned to a tool

    Args:
        run_id (UUID):
        tool_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, List['Supervisor']]
    """

    return sync_detailed(
        run_id=run_id,
        tool_id=tool_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[Any, List["Supervisor"]]]:
    """Get the supervisors assigned to a tool

    Args:
        run_id (UUID):
        tool_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, List['Supervisor']]]
    """

    kwargs = _get_kwargs(
        run_id=run_id,
        tool_id=tool_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[Any, List["Supervisor"]]]:
    """Get the supervisors assigned to a tool

    Args:
        run_id (UUID):
        tool_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, List['Supervisor']]
    """

    return (
        await asyncio_detailed(
            run_id=run_id,
            tool_id=tool_id,
            client=client,
        )
    ).parsed