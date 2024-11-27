"""
Unit tests for the API client.
"""

import unittest
from unittest.mock import patch, Mock
from requests.exceptions import RequestException

from sentinel.clients.api_client import APIClient
class TestAPIClient(unittest.TestCase):
    def setUp(self):
        self.client = APIClient(api_key="test_api_key")

    @patch('requests.Session.post')
    def test_post_success(self, mock_post):
        mock_post.return_value = Mock(status_code=200, json=lambda: {})
        response = self.client.post("/test-endpoint", json={"key": "value"})
        self.assertEqual(response.status_code, 200)
        mock_post.assert_called_once()

    @patch('requests.Session.post')
    def test_post_failure(self, mock_post):
        mock_post.side_effect = RequestException("Network error")
        with self.assertRaises(RequestException):
            self.client.post("/test-endpoint", json={"key": "value"})
