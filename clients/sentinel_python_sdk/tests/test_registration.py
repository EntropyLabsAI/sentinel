"""
Unit tests for project and task registration.
"""

import unittest
from unittest.mock import patch, Mock
from sentinel.registration.project import register_project
from sentinel.registration.task import register_task
from sentinel.config import settings

class TestRegistration(unittest.TestCase):
    def setUp(self):
        settings.api_key = "test_api_key"

    @patch('sentinel.api.client.APIClient.post')
    def test_register_project_success(self, mock_post):
        mock_post.return_value = Mock(
            status_code=200,
            json=lambda: {"project_id": "new_project_id"}
        )
        project_id = register_project(name="Test Project")
        self.assertEqual(project_id, "new_project_id")
        mock_post.assert_called_once()

    @patch('sentinel.api.client.APIClient.post')
    def test_register_task_success(self, mock_post):
        mock_post.return_value = Mock(
            status_code=200,
            json=lambda: {"task_id": "new_task_id"}
        )
        task_id = register_task(name="Test Task", project_id="existing_project_id")
        self.assertEqual(task_id, "new_task_id")
        mock_post.assert_called_once()

    @patch('sentinel.api.client.APIClient.post')
    def test_register_task_failure(self, mock_post):
        mock_post.side_effect = Exception("API error")
        with self.assertRaises(ValueError):
            register_task(name="Test Task", project_id="invalid_project_id")
