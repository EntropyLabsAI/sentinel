"""
Unit tests for the initialization module.
"""

import unittest
from unittest.mock import patch, Mock
from sentinel import init
from sentinel.config import settings

class TestInitialization(unittest.TestCase):
    def test_init_with_invalid_api_key(self):
        with self.assertRaises(ValueError):
            init(api_key="")

