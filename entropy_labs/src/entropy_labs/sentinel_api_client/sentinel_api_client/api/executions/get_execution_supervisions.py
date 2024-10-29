from http import HTTPStatus
from typing import Any, Dict, Optional, Union
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.execution_supervisions import ExecutionSupervisions
from ...types import Response


def _get_kwargs(
    execution_id: UUID,
) -> Dict[str, Any]:
    _kwargs: Dict[str, Any] = {
        "method": "get",
        "url": f"/api/executions/{execution_id}/supervisions",
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[ExecutionSupervisions]:
    if response.status_code == 200:
        response_200 = ExecutionSupervisions.from_dict(response.json())

        return response_200
    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[ExecutionSupervisions]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    execution_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[ExecutionSupervisions]:
    """Get supervision info for an execution

    Args:
        execution_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ExecutionSupervisions]
    """

    kwargs = _get_kwargs(
        execution_id=execution_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    execution_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[ExecutionSupervisions]:
    """Get supervision info for an execution

    Args:
        execution_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ExecutionSupervisions
    """

    return sync_detailed(
        execution_id=execution_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    execution_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[ExecutionSupervisions]:
    """Get supervision info for an execution

    Args:
        execution_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[ExecutionSupervisions]
    """

    kwargs = _get_kwargs(
        execution_id=execution_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    execution_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[ExecutionSupervisions]:
    """Get supervision info for an execution

    Args:
        execution_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        ExecutionSupervisions
    """

    return (
        await asyncio_detailed(
            execution_id=execution_id,
            client=client,
        )
    ).parsed
