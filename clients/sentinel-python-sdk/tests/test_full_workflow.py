"""
Integration tests for the full Sentinel SDK workflow.
"""

import unittest
from unittest.mock import patch, MagicMock
from openai import OpenAI
from sentinel import init
from sentinel.config import settings

class TestFullWorkflow(unittest.TestCase):
    @patch('sentinel.api.client.APIClient.post')
    @patch('sentinel.init.register_project')
    def test_full_workflow(self, mock_register_project, mock_api_client_post):
        # Mock the API client's post method
        mock_api_client_post.return_value = MagicMock(status_code=200)
        mock_register_project.return_value = "test_project_id"

        # Initialize the SDK
        init(api_key="test_api_key")

        client = OpenAI()

        response = client.chat.completions.create(
            model="gpt-4o-mini",
            messages=[{"role": "user", "content": "Hello, world!"}],
            max_tokens=50,
            temperature=0.5
        )

        print(response)

        # Verify that the response is as expected
        # self.assertEqual(response['id'], 'test_chat_id')
        # self.assertEqual(response['choices'], [])

        # # Verify that the API client post method was called for request/response submissions
        # # Expecting 2 calls: one for submit_request, one for submit_response
        # self.assertEqual(mock_api_client_post.call_count, 2)

        # # Check the endpoints called
        # called_endpoints = [call_args[0][0] for call_args in mock_api_client_post.call_args_list]
        # self.assertIn("/requests", called_endpoints)
        # self.assertIn("/responses", called_endpoints)
