from http import HTTPStatus
from typing import Any, Dict, List, Optional, Union, cast
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.error_response import ErrorResponse
from ...types import Response


def _get_kwargs(
    run_id: UUID,
    tool_id: UUID,
    *,
    body: List[List[UUID]],
) -> Dict[str, Any]:
    headers: Dict[str, Any] = {}

    _kwargs: Dict[str, Any] = {
        "method": "post",
        "url": f"/api/runs/{run_id}/tools/{tool_id}/supervisors",
    }

    _body = []
    for componentsschemas_supervisor_chain_assignment_item_data in body:
        componentsschemas_supervisor_chain_assignment_item = []
        for (
            componentsschemas_supervisor_chain_assignment_item_item_data
        ) in componentsschemas_supervisor_chain_assignment_item_data:
            componentsschemas_supervisor_chain_assignment_item_item = str(
                componentsschemas_supervisor_chain_assignment_item_item_data
            )
            componentsschemas_supervisor_chain_assignment_item.append(
                componentsschemas_supervisor_chain_assignment_item_item
            )

        _body.append(componentsschemas_supervisor_chain_assignment_item)

    _kwargs["json"] = _body
    headers["Content-Type"] = "application/json"

    _kwargs["headers"] = headers
    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[Any, ErrorResponse]]:
    if response.status_code == 200:
        response_200 = cast(Any, None)
        return response_200
    if response.status_code == 400:
        response_400 = ErrorResponse.from_dict(response.json())

        return response_400
    if client.raise_on_unexpected_status:
        raise errors.UnexpectedStatus(response.status_code, response.content)
    else:
        return None


def _build_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Response[Union[Any, ErrorResponse]]:
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
    body: List[List[UUID]],
) -> Response[Union[Any, ErrorResponse]]:
    """Assign supervisors to a tool for a given run

     Specify an array of arrays of supervisors in supervision order. Each array represents a list of
    supervisors that will be called in parallel, with the first supervisor in each array being called
    first, and so on. These supervisors will be called in parallel when this tool is invoked for the
    remainder of the run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[List[UUID]]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, ErrorResponse]]
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


def sync(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
    body: List[List[UUID]],
) -> Optional[Union[Any, ErrorResponse]]:
    """Assign supervisors to a tool for a given run

     Specify an array of arrays of supervisors in supervision order. Each array represents a list of
    supervisors that will be called in parallel, with the first supervisor in each array being called
    first, and so on. These supervisors will be called in parallel when this tool is invoked for the
    remainder of the run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[List[UUID]]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, ErrorResponse]
    """

    return sync_detailed(
        run_id=run_id,
        tool_id=tool_id,
        client=client,
        body=body,
    ).parsed


async def asyncio_detailed(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
    body: List[List[UUID]],
) -> Response[Union[Any, ErrorResponse]]:
    """Assign supervisors to a tool for a given run

     Specify an array of arrays of supervisors in supervision order. Each array represents a list of
    supervisors that will be called in parallel, with the first supervisor in each array being called
    first, and so on. These supervisors will be called in parallel when this tool is invoked for the
    remainder of the run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[List[UUID]]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, ErrorResponse]]
    """

    kwargs = _get_kwargs(
        run_id=run_id,
        tool_id=tool_id,
        body=body,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    run_id: UUID,
    tool_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
    body: List[List[UUID]],
) -> Optional[Union[Any, ErrorResponse]]:
    """Assign supervisors to a tool for a given run

     Specify an array of arrays of supervisors in supervision order. Each array represents a list of
    supervisors that will be called in parallel, with the first supervisor in each array being called
    first, and so on. These supervisors will be called in parallel when this tool is invoked for the
    remainder of the run.

    Args:
        run_id (UUID):
        tool_id (UUID):
        body (List[List[UUID]]):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, ErrorResponse]
    """

    return (
        await asyncio_detailed(
            run_id=run_id,
            tool_id=tool_id,
            client=client,
            body=body,
        )
    ).parsed
