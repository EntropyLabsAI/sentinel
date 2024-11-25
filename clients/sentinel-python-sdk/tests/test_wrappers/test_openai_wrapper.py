"""
Unit tests for the OpenAI client wrapper.
"""

import unittest
from unittest.mock import patch, MagicMock
from openai import OpenAI

from sentinel.wrappers.openai import patch_openai, _original_completion_create
from sentinel.config import settings

client = OpenAI()

class TestOpenAIWrapper(unittest.TestCase):
    def setUp(self):
        settings.api_key = "test_api_key"
        settings.project_id = "test_project_id"
        patch_openai()

    def tearDown(self):
        # Restore original methods after tests
        client.chat.completions.create = _original_completion_create

    @patch('sentinel.wrappers.openai._original_completion_create')
    @patch('sentinel.wrappers.openai.submit_request')
    @patch('sentinel.wrappers.openai.submit_response')
    def test_wrapped_completion_create(self, mock_submit_response, mock_submit_request, mock_original_create):
        # Mock the original OpenAI method
        mock_original_create.return_value = {"id": "test_completion_id"}

        response = client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{"role": "user", "content": "Hello, world!"}],
            max_tokens=5)

        # Verify that submit_request and submit_response were called
        mock_submit_request.assert_called_once()
        mock_submit_response.assert_called_once()

        # Check if response is as expected
        self.assertEqual(response['id'], "test_completion_id")


    @patch('sentinel.wrappers.openai.APIClient.post')
    def test_submit_request(self, mock_post):
        from sentinel.wrappers.openai import submit_request
        mock_post.return_value = MagicMock(status_code=200)
        request_data = {
            "messages": [{"role": "user", "content": "Hi"}],
            "model": "gpt-4o-mini",
            "temperature": 0.5,
            "max_tokens": 50,
            "additional_params": {}
        }
        submit_request(request_data)
        mock_post.assert_called_once_with("/requests", json=request_data)

    @patch('sentinel.wrappers.openai.APIClient.post')
    def test_submit_response(self, mock_post):
        from sentinel.wrappers.openai import submit_response
        mock_post.return_value = MagicMock(status_code=200)
        response_data = {"id": "test_response_id", "choices": []}
        submit_response(response_data)
        mock_post.assert_called_once_with("/responses", json=response_data)
