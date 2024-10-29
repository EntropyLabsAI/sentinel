from http import HTTPStatus
from typing import Any, Dict, Optional, Union, cast
from uuid import UUID

import httpx

from ... import errors
from ...client import AuthenticatedClient, Client
from ...models.supervision_status import SupervisionStatus
from ...types import Response


def _get_kwargs(
    review_id: UUID,
) -> Dict[str, Any]:
    _kwargs: Dict[str, Any] = {
        "method": "get",
        "url": f"/api/reviews/{review_id}/status",
    }

    return _kwargs


def _parse_response(
    *, client: Union[AuthenticatedClient, Client], response: httpx.Response
) -> Optional[Union[Any, SupervisionStatus]]:
    if response.status_code == 200:
        response_200 = SupervisionStatus.from_dict(response.json())

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
) -> Response[Union[Any, SupervisionStatus]]:
    return Response(
        status_code=HTTPStatus(response.status_code),
        content=response.content,
        headers=response.headers,
        parsed=_parse_response(client=client, response=response),
    )


def sync_detailed(
    review_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[Any, SupervisionStatus]]:
    """Get supervision status

    Args:
        review_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, SupervisionStatus]]
    """

    kwargs = _get_kwargs(
        review_id=review_id,
    )

    response = client.get_httpx_client().request(
        **kwargs,
    )

    return _build_response(client=client, response=response)


def sync(
    review_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[Any, SupervisionStatus]]:
    """Get supervision status

    Args:
        review_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, SupervisionStatus]
    """

    return sync_detailed(
        review_id=review_id,
        client=client,
    ).parsed


async def asyncio_detailed(
    review_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Response[Union[Any, SupervisionStatus]]:
    """Get supervision status

    Args:
        review_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Response[Union[Any, SupervisionStatus]]
    """

    kwargs = _get_kwargs(
        review_id=review_id,
    )

    response = await client.get_async_httpx_client().request(**kwargs)

    return _build_response(client=client, response=response)


async def asyncio(
    review_id: UUID,
    *,
    client: Union[AuthenticatedClient, Client],
) -> Optional[Union[Any, SupervisionStatus]]:
    """Get supervision status

    Args:
        review_id (UUID):

    Raises:
        errors.UnexpectedStatus: If the server returns an undocumented status code and Client.raise_on_unexpected_status is True.
        httpx.TimeoutException: If the request takes longer than Client.timeout.

    Returns:
        Union[Any, SupervisionStatus]
    """

    return (
        await asyncio_detailed(
            review_id=review_id,
            client=client,
        )
    ).parsed
