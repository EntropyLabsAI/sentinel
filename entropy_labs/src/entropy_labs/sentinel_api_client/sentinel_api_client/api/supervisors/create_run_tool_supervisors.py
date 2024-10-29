from http import HTTPStatus
from typing import Any, Dict, List, Optional, Union
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...types import Response


def _get_kwargs(
    run_id: UUID,
    tool_id: UUID,
    *,
    body: List[UUID],
) -> Dict[str, Any]:
    headers: Dict[str, Any] = {}

    _kwargs: Dict[str, Any] = {
        "method": "post",
        "url": f"/api/runs/{run_id}/tools/{tool_id}/supervisors",
    }

    _body = []
    for body_item_data in body:
        body_item = str(body_item_data)
        _body.append(body_item)

    _kwargs["json"] = _body
    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(*, client: Union[AuthenticatedClient, Client], response: httpx.Response) -> Optional[Any]:
    if response.status_code == 200:
        return None
    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(*, client: Union[AuthenticatedClient, Client], response: httpx.Response) -> Response[Any]:
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
    body: List[UUID],
) -> Response[Any]:
    """Assign a list of supervisors to a tool for a given run

     Specify an array of supervisors in supervision order. The first supervisor will be called first, and
    so on. These supervisors will be called in order when this tool is invoked for the remainder of the
    run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[UUID]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any]
    """

    kwargs = _get_kwargs(
        run_id=run_id,
        tool_id=tool_id,
        body=body,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


async def asyncio_detailed(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
    body: List[UUID],
) -> Response[Any]:
    """Assign a list of supervisors to a tool for a given run

     Specify an array of supervisors in supervision order. The first supervisor will be called first, and
    so on. These supervisors will be called in order when this tool is invoked for the remainder of the
    run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[UUID]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Any]
    """

    kwargs = _get_kwargs(
        run_id=run_id,
        tool_id=tool_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)
