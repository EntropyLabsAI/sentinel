"""
Unit tests for the configuration module.
"""

import unittest
from sentinel.config import Settings

class TestConfig(unittest.TestCase):
    def test_settings_initialization(self):
        settings = Settings()
        self.assertIsNone(settings.api_key)
        self.assertIsNone(settings.project_id)

    def test_settings_assignment(self):
        settings = Settings()
        settings.api_key = "test_api_key"
        settings.project_id = "test_project_id"
        self.assertEqual(settings.api_key, "test_api_key")
        self.assertEqual(settings.project_id, "test_project_id")
