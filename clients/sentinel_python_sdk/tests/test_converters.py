"""
Unit tests for the format converter.
"""

import unittest
from sentinel.converters.format_converter import convert_to_standard_format

class TestFormatConverter(unittest.TestCase):
    def test_convert_complete_data(self):
        original_data = {
            "messages": [{"role": "user", "content": "Hello"}],
            "model": "gpt-4o-mini",
            "temperature": 0.7,
            "max_tokens": 100,
            "frequency_penalty": 0.5
        }
        standard_data = convert_to_standard_format(original_data)
        self.assertEqual(standard_data["messages"], [{"role": "user", "content": "Hello"}])
        self.assertEqual(standard_data["model"], "gpt-4o-mini")
        self.assertEqual(standard_data["temperature"], 0.7)
        self.assertEqual(standard_data["max_tokens"], 100)
        self.assertEqual(standard_data["additional_params"], {"frequency_penalty": 0.5})

    def test_convert_partial_data(self):
        original_data = {
            "model": "gpt-4o-mini",
        }
        standard_data = convert_to_standard_format(original_data)
        self.assertEqual(standard_data["messages"], [])
        self.assertEqual(standard_data["model"], "gpt-4o-mini")
        self.assertEqual(standard_data["temperature"], 1.0)
        self.assertEqual(standard_data["max_tokens"], 256)
        self.assertEqual(standard_data["additional_params"], {})
